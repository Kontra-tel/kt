package deploycheck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckProjectCLI(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".kt", "project.yaml"), "template: cli\napp: tool\nkind: cli\nservices:\n")
	writeFile(t, filepath.Join(root, "nfpm.yaml"), "name: ${APP}\n")
	writeFile(t, filepath.Join(root, "deploy", "config", "app.env.example"), "APP_ENV=production\n")
	writeExec(t, filepath.Join(root, "deploy", "bin", "tool"), "#!/bin/sh\n")

	checks, err := CheckProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if HasErrors(checks) {
		t.Fatalf("expected no errors: %+v", checks)
	}
}

func TestCheckProjectServiceMissingRunner(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".kt", "project.yaml"), "template: service\napp: api\nkind: service\nservices: api\nuser: api\ngroup: api\n")
	writeFile(t, filepath.Join(root, "nfpm.yaml"), "name: ${APP}\n")
	writeFile(t, filepath.Join(root, "deploy", "config", "app.env.example"), "APP_ENV=production\n")
	writeFile(t, filepath.Join(root, "deploy", "systemd", "api.service"), "[Service]\nExecStart=/usr/lib/api/bin/api\nNoNewPrivileges=true\nPrivateTmp=true\nProtectSystem=strict\n")

	checks, err := CheckProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if !HasErrors(checks) {
		t.Fatalf("expected missing runner error: %+v", checks)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeExec(t *testing.T, path, content string) {
	t.Helper()
	writeFile(t, path, content)
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatal(err)
	}
}
