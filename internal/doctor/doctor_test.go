package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckReportsMissingTools(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("DOCTOR_TOOLS := git definitely-not-a-tool\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || len(result.Tools) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if result.Tools[0].Name != "git" || !result.Tools[0].Found || result.Tools[1].Found {
		t.Fatalf("tools = %+v", result.Tools)
	}
}

func TestDoctorToolsAcceptsMakeAssignments(t *testing.T) {
	for _, assignment := range []string{":=", "?=", "+=", "="} {
		got := doctorTools("DOCTOR_TOOLS " + assignment + " git make\n")
		if len(got) != 2 || got[0] != "git" || got[1] != "make" {
			t.Errorf("doctorTools(%q) = %v", assignment, got)
		}
	}
}
