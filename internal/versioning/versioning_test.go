package versioning_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.kontra.tel/kontra.tel/build-tools/internal/versioning"
)

func writeVersion(t *testing.T, v string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "version.txt")
	if err := os.WriteFile(f, []byte(v+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestBump(t *testing.T) {
	tests := []struct {
		start, kind, want string
	}{
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "major", "2.0.0"},
		{"0.0.9", "patch", "0.0.10"},
		{"1.9.9", "minor", "1.10.0"},
		{"9.9.9", "major", "10.0.0"},
	}
	for _, tc := range tests {
		f := writeVersion(t, tc.start)
		got, err := versioning.Bump(f, tc.kind)
		if err != nil {
			t.Errorf("Bump(%q, %q): unexpected error: %v", tc.start, tc.kind, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Bump(%q, %q) = %q, want %q", tc.start, tc.kind, got, tc.want)
		}
	}
}

func TestBump_WritesPersisted(t *testing.T) {
	f := writeVersion(t, "1.0.0")
	if _, err := versioning.Bump(f, "patch"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "1.0.1" {
		t.Errorf("file contains %q, want %q", strings.TrimSpace(string(data)), "1.0.1")
	}
}

func TestBump_InvalidFormat(t *testing.T) {
	tests := []string{"notaversion", "1.2", "1.2.3.4", "a.b.c"}
	for _, bad := range tests {
		f := writeVersion(t, bad)
		_, err := versioning.Bump(f, "patch")
		if err == nil {
			t.Errorf("Bump(%q, patch): expected error, got nil", bad)
		}
	}
}

func TestBump_UnknownKind(t *testing.T) {
	f := writeVersion(t, "1.0.0")
	_, err := versioning.Bump(f, "hotfix")
	if err == nil {
		t.Fatal("expected error for unknown bump kind, got nil")
	}
}

func TestBump_MissingFile(t *testing.T) {
	_, err := versioning.Bump("/nonexistent/version.txt", "patch")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
