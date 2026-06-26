package ktconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func setup(t *testing.T, content string) {
	t.Helper()
	f := filepath.Join(t.TempDir(), "project.yaml")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	orig := path
	path = f
	t.Cleanup(func() { path = orig })
}

func TestGet(t *testing.T) {
	setup(t, "template: app\napp: my-api\nuser: svc\n")
	tests := []struct{ key, want string }{
		{"template", "app"},
		{"app", "my-api"},
		{"user", "svc"},
	}
	for _, tc := range tests {
		got, err := Get(tc.key)
		if err != nil {
			t.Errorf("Get(%q): unexpected error: %v", tc.key, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Get(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestGet_MissingKey(t *testing.T) {
	setup(t, "app: my-api\n")
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestGet_MissingFile(t *testing.T) {
	orig := path
	path = "/nonexistent/.kt/project.yaml"
	defer func() { path = orig }()
	_, err := Get("app")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestGet_IgnoresComments(t *testing.T) {
	setup(t, "# comment\napp: my-api\n")
	got, err := Get("app")
	if err != nil {
		t.Fatal(err)
	}
	if got != "my-api" {
		t.Errorf("got %q, want %q", got, "my-api")
	}
}

func TestSet_UpdateExisting(t *testing.T) {
	setup(t, "app: my-api\nport: 4002\n")
	if err := Set("port", "9000"); err != nil {
		t.Fatal(err)
	}
	got, err := Get("port")
	if err != nil {
		t.Fatal(err)
	}
	if got != "9000" {
		t.Errorf("got %q, want %q", got, "9000")
	}
	// Other keys must be unchanged.
	app, _ := Get("app")
	if app != "my-api" {
		t.Errorf("Set modified unrelated key: app = %q", app)
	}
}

func TestSet_AppendNew(t *testing.T) {
	setup(t, "app: my-api\n")
	if err := Set("newkey", "newval"); err != nil {
		t.Fatal(err)
	}
	got, err := Get("newkey")
	if err != nil {
		t.Fatal(err)
	}
	if got != "newval" {
		t.Errorf("got %q, want %q", got, "newval")
	}
}

func TestSet_Idempotent(t *testing.T) {
	setup(t, "app: my-api\n")
	if err := Set("app", "my-api"); err != nil {
		t.Fatal(err)
	}
	got, _ := Get("app")
	if got != "my-api" {
		t.Errorf("got %q, want %q", got, "my-api")
	}
}

func TestAll(t *testing.T) {
	setup(t, "template: app\napp: my-api\nuser: svc\n")
	pairs, err := All()
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]string{
		{"template", "app"},
		{"app", "my-api"},
		{"user", "svc"},
	}
	if len(pairs) != len(want) {
		t.Fatalf("got %d pairs, want %d", len(pairs), len(want))
	}
	for i, w := range want {
		if pairs[i] != w {
			t.Errorf("pair[%d]: got %v, want %v", i, pairs[i], w)
		}
	}
}

func TestAll_SkipsComments(t *testing.T) {
	setup(t, "# project config\napp: my-api\n# end\n")
	pairs, err := All()
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 {
		t.Errorf("got %d pairs, want 1 (comments should be skipped)", len(pairs))
	}
}

func TestLoad_NormalizesServiceProject(t *testing.T) {
	setup(t, "template: service\napp: my-api\nkind: service\nuser: svc\ngroup: ops\n")
	project, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if project.Kind != "service" {
		t.Fatalf("kind = %q, want service", project.Kind)
	}
	if project.Services != "my-api" {
		t.Fatalf("services = %q, want my-api", project.Services)
	}
	got := project.ServicesList()
	if len(got) != 1 || got[0] != "my-api" {
		t.Fatalf("ServicesList() = %v, want [my-api]", got)
	}
}

func TestLoad_NormalizesLegacyMultiProject(t *testing.T) {
	setup(t, "template: multi\napp: suite\n")
	project, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if project.Kind != "multi-service" {
		t.Fatalf("kind = %q, want multi-service", project.Kind)
	}
	if project.Services != "suite-backend,suite-frontend" {
		t.Fatalf("services = %q", project.Services)
	}
}

func TestLoad_NormalizesCLIProject(t *testing.T) {
	setup(t, "template: cli\napp: tool\nkind: cli\nservices:\n")
	project, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if project.HasServices() {
		t.Fatalf("HasServices() = true, want false")
	}
}
