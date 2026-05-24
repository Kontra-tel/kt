package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.kontra.tel/kontra.tel/build-tools/internal/assets"
	"git.kontra.tel/kontra.tel/build-tools/internal/scaffold"
)

func newScaffolder() scaffold.Scaffolder {
	return scaffold.Scaffolder{FS: assets.FS}
}

// TestTemplatesWithDesc checks that all expected templates are present,
// sorted, and have descriptions.
func TestTemplatesWithDesc(t *testing.T) {
	s := newScaffolder()
	infos, err := s.TemplatesWithDesc()
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]string)
	for _, ti := range infos {
		byName[ti.Name] = ti.Desc
	}

	for _, name := range []string{"generic-service", "generic-cli", "go-cli", "java-service", "node-service", "multi-service"} {
		if _, ok := byName[name]; !ok {
			t.Errorf("template %q not found", name)
		}
	}
	for _, ti := range infos {
		if ti.Desc == "" {
			t.Errorf("template %q has empty description", ti.Name)
		}
	}
	for i := 1; i < len(infos); i++ {
		if infos[i].Name < infos[i-1].Name {
			t.Errorf("templates not sorted: %q comes after %q", infos[i].Name, infos[i-1].Name)
		}
	}
}

// TestTemplates checks that the names-only list is consistent with TemplatesWithDesc.
func TestTemplates(t *testing.T) {
	s := newScaffolder()
	names, err := s.Templates()
	if err != nil {
		t.Fatal(err)
	}
	infos, _ := s.TemplatesWithDesc()
	if len(names) != len(infos) {
		t.Errorf("Templates() returned %d names, TemplatesWithDesc() returned %d infos", len(names), len(infos))
	}
	for i, n := range names {
		if n != infos[i].Name {
			t.Errorf("names[%d] = %q, infos[%d].Name = %q", i, n, i, infos[i].Name)
		}
	}
}

// TestInit_MissingApp checks that Init fails without an app name.
func TestInit_MissingApp(t *testing.T) {
	s := newScaffolder()
	err := s.Init(t.TempDir(), scaffold.Context{Template: "go-cli"}, false)
	if err == nil {
		t.Fatal("expected error for missing app name, got nil")
	}
}

// TestInit_UnknownTemplate checks that Init fails for an unknown template.
func TestInit_UnknownTemplate(t *testing.T) {
	s := newScaffolder()
	err := s.Init(t.TempDir(), scaffold.Context{Template: "nonexistent", App: "myapp"}, false)
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
}

// TestInit_GoCliFiles checks that the go-cli template produces the expected files.
func TestInit_GoCliFiles(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "go-cli", App: "my-tool"}, false); err != nil {
		t.Fatal(err)
	}

	mustExist(t, dir,
		"Makefile",
		"nfpm.yaml",
		"version.txt",
		".gitignore",
		".kt/project.yaml",
		".kt/mk/common.mk",
		".kt/mk/nfpm.mk",
		"go.mod",
		"cmd/my-tool/main.go",
		"deploy/config/app.env.example",
	)
}

// TestInit_JavaServiceFiles checks that the java-service template produces the expected files,
// including the correctly renamed service unit.
func TestInit_JavaServiceFiles(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	ctx := scaffold.Context{Template: "java-service", App: "my-api", Port: "4002", ServiceUser: "svc", ServiceGroup: "svc"}
	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}

	mustExist(t, dir,
		"Makefile",
		"nfpm.yaml",
		".kt/project.yaml",
		"deploy/systemd/my-api.service",
		"deploy/scripts/postinstall.sh",
		"deploy/scripts/preremove.sh",
		"deploy/config/app.env.example",
	)
}

// TestInit_ProjectYAML checks that .kt/project.yaml is rendered with the correct values.
func TestInit_ProjectYAML(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	ctx := scaffold.Context{Template: "java-service", App: "my-api", Port: "4002", ServiceUser: "svc", ServiceGroup: "ops"}
	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".kt/project.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{"app: my-api", "port: 4002", "user: svc", "group: ops", "template: java-service"} {
		if !strings.Contains(content, want) {
			t.Errorf("project.yaml missing %q:\n%s", want, content)
		}
	}
}

// TestInit_TemplateVarsRendered checks that no unrendered Go template syntax
// remains in output files.
func TestInit_TemplateVarsRendered(t *testing.T) {
	s := newScaffolder()
	templates := []string{"java-service", "node-service", "go-cli", "generic-service", "generic-cli"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			ctx := scaffold.Context{Template: tmpl, App: "testapp", Port: "8080", ServiceUser: "svc", ServiceGroup: "svc"}
			if err := s.Init(dir, ctx, false); err != nil {
				t.Fatal(err)
			}
			err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if strings.Contains(string(data), "{{") {
					t.Errorf("file %s contains unrendered template syntax", path)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestInit_ServiceFileRename checks that service units are renamed for all service templates.
func TestInit_ServiceFileRename(t *testing.T) {
	s := newScaffolder()
	tests := []struct {
		tmpl  string
		files []string
	}{
		{"java-service", []string{"deploy/systemd/myapp.service"}},
		{"node-service", []string{"deploy/systemd/myapp.service"}},
		{"generic-service", []string{"deploy/systemd/myapp.service"}},
		{"multi-service", []string{"deploy/systemd/myapp-backend.service", "deploy/systemd/myapp-frontend.service"}},
	}
	for _, tc := range tests {
		t.Run(tc.tmpl, func(t *testing.T) {
			dir := t.TempDir()
			ctx := scaffold.Context{Template: tc.tmpl, App: "myapp", Port: "8080", ServiceUser: "svc", ServiceGroup: "svc"}
			if err := s.Init(dir, ctx, false); err != nil {
				t.Fatal(err)
			}
			mustExist(t, dir, tc.files...)
		})
	}
}

// TestInit_NoOverwriteWithoutForce checks that Init does not clobber existing files
// unless --force is set.
func TestInit_NoOverwriteWithoutForce(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	ctx := scaffold.Context{Template: "go-cli", App: "my-tool"}

	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}
	makefile := filepath.Join(dir, "Makefile")
	if err := os.WriteFile(makefile, []byte("custom"), 0644); err != nil {
		t.Fatal(err)
	}

	// Second init without force must not overwrite.
	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(makefile)
	if string(data) != "custom" {
		t.Error("Init without --force overwrote existing Makefile")
	}

	// Second init with force must overwrite.
	if err := s.Init(dir, ctx, true); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(makefile)
	if string(data) == "custom" {
		t.Error("Init with --force did not overwrite existing Makefile")
	}
}

// mustExist fails the test if any of the given paths (relative to base) do not exist.
func mustExist(t *testing.T, base string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		full := filepath.Join(base, p)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected file not found: %s", p)
		}
	}
}
