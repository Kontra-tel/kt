package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.kontra.tel/kontra.tel/Kt/internal/scaffold"
)

func TestInitAddsManagedPackageMarkersAndVerifyTarget(t *testing.T) {
	s := newScaffolder()
	for _, template := range []string{"app", "cli", "mixed", "multi", "service"} {
		t.Run(template, func(t *testing.T) {
			dir := t.TempDir()
			if err := s.Init(dir, scaffold.Context{Template: template, App: "demo"}, false); err != nil {
				t.Fatal(err)
			}
			manifest, err := os.ReadFile(filepath.Join(dir, "nfpm.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			for _, marker := range []string{"# kt:contents:start", "# kt:contents:end"} {
				if !strings.Contains(string(manifest), marker) {
					t.Errorf("nfpm.yaml missing %q", marker)
				}
			}
			makefile, err := os.ReadFile(filepath.Join(dir, "Makefile"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(makefile), "verify: config-check") {
				t.Fatal("Makefile missing verify target")
			}
		})
	}
}
