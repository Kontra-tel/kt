package deploycheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.kontra.tel/kontra.tel/Kt/internal/ktconfig"
)

func TestPlanUsesManifestCommandsAndServices(t *testing.T) {
	project := ktconfig.Project{
		App:  "suite",
		Kind: "mixed",
		Commands: []ktconfig.Command{
			{Name: "suite", Path: "deploy/bin/suite"},
			{Name: "suite-admin", Path: "deploy/bin/suite-admin"},
		},
		ServiceEntries: []ktconfig.Service{{
			Name:   "suite-service",
			Runner: "deploy/run/suite-service",
			Unit:   "deploy/systemd/suite-service.service",
		}},
		Config: ktconfig.ConfigInfo{Dir: "deploy/config", InstallDir: "/etc/suite", ExampleSuffix: ".example"},
	}

	plan := Plan(project)
	want := []PackageEntry{
		{Source: "dist/app", Destination: "/usr/lib/suite"},
		{Source: "deploy/bin/suite", Destination: "/usr/bin/suite", Mode: 0o755},
		{Source: "deploy/bin/suite-admin", Destination: "/usr/bin/suite-admin", Mode: 0o755},
		{Source: "deploy/run/suite-service", Destination: "/usr/lib/suite/bin/suite-service", Mode: 0o755},
		{Source: "deploy/config/*.example", Destination: "/etc/suite/", Type: "config|noreplace"},
		{Source: "deploy/systemd/suite-service.service", Destination: "/usr/lib/systemd/system/suite-service.service"},
	}
	if len(plan.Entries) != len(want) {
		t.Fatalf("entries = %#v", plan.Entries)
	}
	for i := range want {
		if plan.Entries[i] != want[i] {
			t.Errorf("entry[%d] = %#v, want %#v", i, plan.Entries[i], want[i])
		}
	}
}

func TestSyncPackageContentsChangesOnlyManagedBlock(t *testing.T) {
	file := filepath.Join(t.TempDir(), "nfpm.yaml")
	original := `name: suite
contents:
  # kt:contents:start
  - src: stale
    dst: /stale
  # kt:contents:end
  # User-owned entries stay below the marker.
  - src: resources
    dst: /usr/lib/suite/resources
scripts:
  postinstall: deploy/scripts/postinstall.sh
`
	if err := os.WriteFile(file, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	plan := PackagePlan{App: "suite", Entries: []PackageEntry{{Source: "deploy/bin/suite", Destination: "/usr/bin/suite", Mode: 0o755}}}

	preview, changed, err := PreviewPackageContents(file, plan)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !strings.Contains(preview, "src: deploy/bin/suite") || !strings.Contains(preview, "src: resources") {
		t.Fatalf("preview did not preserve the user-owned entry:\n%s", preview)
	}
	if changed, err := SyncPackageContents(file, plan); err != nil || !changed {
		t.Fatalf("SyncPackageContents() = (%t, %v)", changed, err)
	}
	updated, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(updated), "src: stale") || !strings.Contains(string(updated), "postinstall:") {
		t.Fatalf("sync changed content outside the managed block:\n%s", updated)
	}
}

func TestSyncPackageContentsRejectsLegacyManifest(t *testing.T) {
	file := filepath.Join(t.TempDir(), "nfpm.yaml")
	if err := os.WriteFile(file, []byte("contents: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := SyncPackageContents(file, PackagePlan{}); err == nil || !strings.Contains(err.Error(), "no kt-managed contents block") {
		t.Fatalf("SyncPackageContents() error = %v", err)
	}
}

func TestCheckProjectReportsMissingPackageMapping(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".kt", "project.yaml"), "template: cli\napp: tool\nkind: cli\nservices: []\ncommands:\n  - name: tool\n    path: deploy/bin/tool\n")
	writeFile(t, filepath.Join(root, "nfpm.yaml"), "contents:\n  - src: dist/app\n    dst: /usr/lib/tool\n  - src: deploy/config/*.example\n    dst: /etc/tool/\n    type: config|noreplace\n")
	writeFile(t, filepath.Join(root, "deploy", "config", "app.env.example"), "APP_ENV=production\n")
	writeExec(t, filepath.Join(root, "deploy", "bin", "tool"), "#!/bin/sh\n")

	checks, err := CheckProject(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range checks {
		if check.Name == "package mapping" && strings.Contains(check.Message, "deploy/bin/tool") {
			return
		}
	}
	t.Fatalf("missing package mapping was not reported: %+v", checks)
}
