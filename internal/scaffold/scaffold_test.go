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

	for _, name := range []string{"app", "multi"} {
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
	err := s.Init(t.TempDir(), scaffold.Context{Template: "app"}, false)
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

// TestInit_AppFiles checks that the app template produces the expected files,
// including the renamed launch script and service unit.
func TestInit_AppFiles(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "app", App: "my-tool"}, false); err != nil {
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
		"deploy/bin/my-tool",
		"deploy/systemd/my-tool.service",
		"deploy/config/app.env.example",
		"deploy/scripts/postinstall.sh",
		"deploy/scripts/preremove.sh",
	)
}

// TestInit_AppBinExecutable checks that the deploy/bin/ launch script is executable.
func TestInit_AppBinExecutable(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "app", App: "my-tool"}, false); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, "deploy/bin/my-tool"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("deploy/bin/my-tool is not executable")
	}
}

// TestInit_MultiFiles checks that the multi template produces the expected files,
// including backend/frontend launch scripts and service units.
func TestInit_MultiFiles(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	ctx := scaffold.Context{Template: "multi", App: "my-app", ServiceUser: "svc", ServiceGroup: "svc"}
	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}

	mustExist(t, dir,
		"Makefile",
		"nfpm.yaml",
		".kt/project.yaml",
		"deploy/bin/my-app-backend",
		"deploy/bin/my-app-frontend",
		"deploy/systemd/my-app-backend.service",
		"deploy/systemd/my-app-frontend.service",
		"deploy/config/app.env.example",
		"deploy/scripts/postinstall.sh",
		"deploy/scripts/preremove.sh",
	)
}

// TestInit_ProjectYAML checks that .kt/project.yaml is rendered with the correct values.
func TestInit_ProjectYAML(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	ctx := scaffold.Context{Template: "app", App: "my-api", ServiceUser: "svc", ServiceGroup: "ops"}
	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".kt/project.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{"app: my-api", "user: svc", "group: ops", "template: app"} {
		if !strings.Contains(content, want) {
			t.Errorf("project.yaml missing %q:\n%s", want, content)
		}
	}
}

// TestInit_TemplateVarsRendered checks that no unrendered Go template syntax
// remains in output files.
func TestInit_TemplateVarsRendered(t *testing.T) {
	s := newScaffolder()
	templates := []string{"app", "multi"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			ctx := scaffold.Context{Template: tmpl, App: "testapp", ServiceUser: "svc", ServiceGroup: "svc"}
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

// TestInit_ServiceFileRename checks that service units and launch scripts are
// renamed correctly for all templates.
func TestInit_ServiceFileRename(t *testing.T) {
	s := newScaffolder()
	tests := []struct {
		tmpl  string
		files []string
	}{
		{"app", []string{
			"deploy/systemd/myapp.service",
			"deploy/bin/myapp",
		}},
		{"multi", []string{
			"deploy/systemd/myapp-backend.service",
			"deploy/systemd/myapp-frontend.service",
			"deploy/bin/myapp-backend",
			"deploy/bin/myapp-frontend",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.tmpl, func(t *testing.T) {
			dir := t.TempDir()
			ctx := scaffold.Context{Template: tc.tmpl, App: "myapp", ServiceUser: "svc", ServiceGroup: "svc"}
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
	ctx := scaffold.Context{Template: "app", App: "my-tool"}

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
