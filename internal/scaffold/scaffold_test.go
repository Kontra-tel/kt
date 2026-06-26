package scaffold_test

import (
	"os"
	"os/exec"
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

	for _, name := range []string{"app", "cli", "mixed", "multi", "service"} {
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
// including the metadata command, service runner, and service unit.
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
		"deploy/run/my-tool",
		"deploy/systemd/my-tool.service",
		"deploy/config/app.env.example",
		"deploy/hooks-examples/postinstall.local.sh",
		"deploy/hooks-examples/preremove.local.sh",
		"deploy/scripts/postinstall.sh",
		"deploy/scripts/preremove.sh",
	)
}

// TestInit_AppScriptsExecutable checks that the app command and runner are executable.
func TestInit_AppScriptsExecutable(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "app", App: "my-tool"}, false); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"deploy/bin/my-tool", "deploy/run/my-tool"} {
		info, err := os.Stat(filepath.Join(dir, rel))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("%s is not executable", rel)
		}
	}
}

// TestInit_CLIFiles checks that the cli template produces the expected files.
func TestInit_CLIFiles(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "cli", App: "my-tool"}, false); err != nil {
		t.Fatal(err)
	}

	mustExist(t, dir,
		"Makefile",
		"nfpm.yaml",
		".kt/project.yaml",
		"deploy/bin/my-tool",
		"deploy/config/app.env.example",
	)
}

func TestInit_MixedFiles(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	ctx := scaffold.Context{Template: "mixed", App: "my-suite", ServiceUser: "svc", ServiceGroup: "svc"}
	if err := s.Init(dir, ctx, false); err != nil {
		t.Fatal(err)
	}

	mustExist(t, dir,
		"Makefile",
		"nfpm.yaml",
		".kt/project.yaml",
		"deploy/bin/my-suite",
		"deploy/bin/my-suite-service",
		"deploy/run/my-suite-service",
		"deploy/systemd/my-suite-service.service",
		"deploy/config/app.env.example",
		"deploy/hooks-examples/postinstall.local.sh",
		"deploy/hooks-examples/preremove.local.sh",
		"deploy/scripts/postinstall.sh",
		"deploy/scripts/preremove.sh",
	)
}

// TestInit_MultiFiles checks that the multi template produces the expected files,
// including the metadata command, service runners, and service units.
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
		"deploy/bin/my-app",
		"deploy/run/my-app-backend",
		"deploy/run/my-app-frontend",
		"deploy/systemd/my-app-backend.service",
		"deploy/systemd/my-app-frontend.service",
		"deploy/config/app.env.example",
		"deploy/hooks-examples/postinstall.local.sh",
		"deploy/hooks-examples/preremove.local.sh",
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

	for _, want := range []string{"app: my-api", "kind: service", "services: my-api", "user: svc", "group: ops", "template: app"} {
		if !strings.Contains(content, want) {
			t.Errorf("project.yaml missing %q:\n%s", want, content)
		}
	}
}

func TestInit_ServiceAliasWritesTemplateName(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "service", App: "my-api"}, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".kt/project.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "template: service") {
		t.Fatalf("project.yaml missing service alias template name:\n%s", content)
	}
	if !strings.Contains(content, "kind: service") {
		t.Fatalf("project.yaml missing normalized kind:\n%s", content)
	}
}

// TestInit_TemplateVarsRendered checks that no unrendered Go template syntax
// remains in output files.
func TestInit_TemplateVarsRendered(t *testing.T) {
	s := newScaffolder()
	templates := []string{"app", "cli", "mixed", "multi", "service"}
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

// TestInit_FHSLayout checks that generated service templates use separate
// locations for packaged artifacts, mutable data, logs, packaged units, and runners.
func TestInit_FHSLayout(t *testing.T) {
	s := newScaffolder()
	tests := []struct {
		tmpl  string
		wants map[string][]string
	}{
		{
			tmpl: "app",
			wants: map[string][]string{
				"nfpm.yaml": {
					"dst: /usr/lib/testapp",
					"dst: /usr/lib/testapp/bin/testapp",
					"dst: /usr/lib/systemd/system/testapp.service",
				},
				"deploy/systemd/testapp.service": {
					"WorkingDirectory=/var/lib/testapp",
					"ReadWritePaths=/var/lib/testapp /var/log/testapp",
					"ExecStart=/usr/lib/testapp/bin/testapp",
				},
				"deploy/scripts/postinstall.sh": {
					"--home /var/lib/testapp",
					"/var/lib/testapp /var/log/testapp",
				},
			},
		},
		{
			tmpl: "multi",
			wants: map[string][]string{
				"nfpm.yaml": {
					"dst: /usr/lib/testapp",
					"dst: /usr/lib/testapp/bin/testapp-backend",
					"dst: /usr/lib/testapp/bin/testapp-frontend",
					"dst: /usr/lib/systemd/system/testapp-backend.service",
					"dst: /usr/lib/systemd/system/testapp-frontend.service",
				},
				"deploy/systemd/testapp-backend.service": {
					"WorkingDirectory=/var/lib/testapp",
					"ReadWritePaths=/var/lib/testapp /var/log/testapp",
					"ExecStart=/usr/lib/testapp/bin/testapp-backend",
				},
				"deploy/systemd/testapp-frontend.service": {
					"WorkingDirectory=/var/lib/testapp",
					"ReadWritePaths=/var/lib/testapp /var/log/testapp",
					"ExecStart=/usr/lib/testapp/bin/testapp-frontend",
				},
				"deploy/scripts/postinstall.sh": {
					"--home /var/lib/testapp",
					"/var/lib/testapp /var/log/testapp",
				},
			},
		},
		{
			tmpl: "mixed",
			wants: map[string][]string{
				"nfpm.yaml": {
					"dst: /usr/lib/testapp",
					"dst: /usr/lib/testapp/bin/testapp-service",
					"dst: /usr/lib/systemd/system/testapp-service.service",
				},
				"deploy/systemd/testapp-service.service": {
					"WorkingDirectory=/var/lib/testapp",
					"ReadWritePaths=/var/lib/testapp /var/log/testapp",
					"ExecStart=/usr/lib/testapp/bin/testapp-service",
				},
				"deploy/scripts/postinstall.sh": {
					"--home /var/lib/testapp",
					"/etc/testapp/hooks",
					"/var/lib/testapp /var/log/testapp",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.tmpl, func(t *testing.T) {
			dir := t.TempDir()
			ctx := scaffold.Context{Template: tc.tmpl, App: "testapp", ServiceUser: "svc", ServiceGroup: "svc"}
			if err := s.Init(dir, ctx, false); err != nil {
				t.Fatal(err)
			}
			for rel, wants := range tc.wants {
				data, err := os.ReadFile(filepath.Join(dir, rel))
				if err != nil {
					t.Fatal(err)
				}
				content := string(data)
				for _, want := range wants {
					if !strings.Contains(content, want) {
						t.Errorf("%s missing %q:\n%s", rel, want, content)
					}
				}
			}
		})
	}
}

// TestInit_CLINoServiceBits checks that the cli template does not scaffold service assets.
func TestInit_CLINoServiceBits(t *testing.T) {
	s := newScaffolder()
	dir := t.TempDir()
	if err := s.Init(dir, scaffold.Context{Template: "cli", App: "testapp"}, false); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		"deploy/systemd",
		"deploy/scripts",
		"deploy/run",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Errorf("%s should not exist in cli template", rel)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, ".kt/project.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, unwanted := range []string{"user:", "group:"} {
		if strings.Contains(content, unwanted) {
			t.Errorf("cli project.yaml should not contain %q:\n%s", unwanted, content)
		}
	}
}

func TestInit_MetadataCommandsSupportJSON(t *testing.T) {
	s := newScaffolder()
	tests := map[string]string{
		"app":   "deploy/bin/testapp",
		"multi": "deploy/bin/testapp",
		"mixed": "deploy/bin/testapp-service",
	}
	for tmpl, rel := range tests {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			if err := s.Init(dir, scaffold.Context{Template: tmpl, App: "testapp"}, false); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			if !strings.Contains(content, "--json") {
				t.Fatalf("%s missing --json support:\n%s", rel, content)
			}
		})
	}
}

func TestInit_TreeShape(t *testing.T) {
	s := newScaffolder()
	tests := map[string][]string{
		"cli": {
			".gitignore",
			".kt/project.yaml",
			"Makefile",
			"deploy/bin/treeapp",
			"deploy/config/app.env.example",
			"nfpm.yaml",
			"version.txt",
		},
		"service": {
			".gitignore",
			".kt/project.yaml",
			"Makefile",
			"deploy/bin/treeapp",
			"deploy/config/app.env.example",
			"deploy/hooks-examples/postinstall.local.sh",
			"deploy/hooks-examples/preremove.local.sh",
			"deploy/run/treeapp",
			"deploy/scripts/postinstall.sh",
			"deploy/scripts/preremove.sh",
			"deploy/systemd/treeapp.service",
			"nfpm.yaml",
			"version.txt",
		},
		"mixed": {
			".gitignore",
			".kt/project.yaml",
			"Makefile",
			"deploy/bin/treeapp",
			"deploy/bin/treeapp-service",
			"deploy/config/app.env.example",
			"deploy/hooks-examples/postinstall.local.sh",
			"deploy/hooks-examples/preremove.local.sh",
			"deploy/run/treeapp-service",
			"deploy/scripts/postinstall.sh",
			"deploy/scripts/preremove.sh",
			"deploy/systemd/treeapp-service.service",
			"nfpm.yaml",
			"version.txt",
		},
		"multi": {
			".gitignore",
			".kt/project.yaml",
			"Makefile",
			"deploy/bin/treeapp",
			"deploy/config/app.env.example",
			"deploy/hooks-examples/postinstall.local.sh",
			"deploy/hooks-examples/preremove.local.sh",
			"deploy/run/treeapp-backend",
			"deploy/run/treeapp-frontend",
			"deploy/scripts/postinstall.sh",
			"deploy/scripts/preremove.sh",
			"deploy/systemd/treeapp-backend.service",
			"deploy/systemd/treeapp-frontend.service",
			"nfpm.yaml",
			"version.txt",
		},
	}
	for tmpl, wants := range tests {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			if err := s.Init(dir, scaffold.Context{Template: tmpl, App: "treeapp"}, false); err != nil {
				t.Fatal(err)
			}
			mustExist(t, dir, wants...)
		})
	}
}

// TestInit_NoLegacyOptPaths checks that newly scaffolded projects no longer
// install application files or writable state below /opt.
func TestInit_NoLegacyOptPaths(t *testing.T) {
	s := newScaffolder()
	for _, tmpl := range []string{"app", "cli", "mixed", "multi", "service"} {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			ctx := scaffold.Context{Template: tmpl, App: "testapp", ServiceUser: "svc", ServiceGroup: "svc"}
			if err := s.Init(dir, ctx, false); err != nil {
				t.Fatal(err)
			}
			assertTreeExcludes(t, dir, "/opt/")
		})
	}
}

// TestInit_HooksAreShellValid checks the generated lifecycle hooks with bash.
func TestInit_HooksAreShellValid(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found")
	}
	s := newScaffolder()
	for _, tmpl := range []string{"app", "cli", "mixed", "multi", "service"} {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			if err := s.Init(dir, scaffold.Context{Template: tmpl, App: "testapp"}, false); err != nil {
				t.Fatal(err)
			}
			paths := []string{
				"deploy/bin/testapp",
				".kt/scripts/postinstall-systemd.sh",
				".kt/scripts/preremove-systemd.sh",
			}
			switch tmpl {
			case "app", "service":
				paths = append(paths,
					"deploy/run/testapp",
					"deploy/scripts/postinstall.sh",
					"deploy/scripts/preremove.sh",
					"deploy/hooks-examples/postinstall.local.sh",
					"deploy/hooks-examples/preremove.local.sh",
				)
			case "mixed":
				paths = append(paths,
					"deploy/bin/testapp-service",
					"deploy/run/testapp-service",
					"deploy/scripts/postinstall.sh",
					"deploy/scripts/preremove.sh",
					"deploy/hooks-examples/postinstall.local.sh",
					"deploy/hooks-examples/preremove.local.sh",
				)
			case "multi":
				paths = append(paths,
					"deploy/run/testapp-backend",
					"deploy/run/testapp-frontend",
					"deploy/scripts/postinstall.sh",
					"deploy/scripts/preremove.sh",
					"deploy/hooks-examples/postinstall.local.sh",
					"deploy/hooks-examples/preremove.local.sh",
				)
			}
			for _, rel := range paths {
				cmd := exec.Command(bash, "-n", filepath.Join(dir, rel))
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Errorf("%s is not valid bash: %v\n%s", rel, err, out)
				}
			}
		})
	}
}

// TestInit_HooksLeaveLifecycleToDeployment checks that package hooks do not
// enable, restart, stop, or disable services during install or upgrade.
func TestInit_HooksLeaveLifecycleToDeployment(t *testing.T) {
	s := newScaffolder()
	for _, tmpl := range []string{"app", "mixed", "multi", "service"} {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			if err := s.Init(dir, scaffold.Context{Template: tmpl, App: "testapp"}, false); err != nil {
				t.Fatal(err)
			}
			for _, rel := range []string{
				"deploy/scripts/postinstall.sh",
				"deploy/scripts/preremove.sh",
				".kt/scripts/postinstall-systemd.sh",
				".kt/scripts/preremove-systemd.sh",
			} {
				data, err := os.ReadFile(filepath.Join(dir, rel))
				if err != nil {
					t.Fatal(err)
				}
				content := string(data)
				for _, command := range []string{
					"systemctl enable",
					"systemctl restart",
					"systemctl stop",
					"systemctl disable",
				} {
					if strings.Contains(content, command) {
						t.Errorf("%s contains lifecycle command %q", rel, command)
					}
				}
			}
		})
	}
}

// TestInit_HooksSupportLocalExtensions checks that generated service hooks can
// delegate to optional local scripts under /etc/<app>/hooks/.
func TestInit_HooksSupportLocalExtensions(t *testing.T) {
	s := newScaffolder()
	for _, tmpl := range []string{"app", "mixed", "multi", "service"} {
		t.Run(tmpl, func(t *testing.T) {
			dir := t.TempDir()
			if err := s.Init(dir, scaffold.Context{Template: tmpl, App: "testapp"}, false); err != nil {
				t.Fatal(err)
			}
			cases := map[string]string{
				"deploy/scripts/postinstall.sh": "/etc/testapp/hooks/postinstall.local.sh",
				"deploy/scripts/preremove.sh":   "/etc/testapp/hooks/preremove.local.sh",
			}
			for rel, want := range cases {
				data, err := os.ReadFile(filepath.Join(dir, rel))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), want) {
					t.Errorf("%s missing local hook reference %q", rel, want)
				}
			}
		})
	}
}

// TestInit_ServiceFileRename checks that service units and generated commands are
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
			"deploy/run/myapp",
		}},
		{"cli", []string{
			"deploy/bin/myapp",
		}},
		{"mixed", []string{
			"deploy/systemd/myapp-service.service",
			"deploy/bin/myapp",
			"deploy/bin/myapp-service",
			"deploy/run/myapp-service",
		}},
		{"multi", []string{
			"deploy/systemd/myapp-backend.service",
			"deploy/systemd/myapp-frontend.service",
			"deploy/bin/myapp",
			"deploy/run/myapp-backend",
			"deploy/run/myapp-frontend",
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

func assertTreeExcludes(t *testing.T, root, unwanted string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), unwanted) {
			t.Errorf("%s contains %q", path, unwanted)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
