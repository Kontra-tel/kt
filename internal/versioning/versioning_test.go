package versioning_test

import (
	"testing"

	"git.kontra.tel/kontra.tel/Kt/internal/versioning"
)

func TestNext(t *testing.T) {
	tests := []struct {
		start, kind, want string
	}{
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "major", "2.0.0"},
		{"1.2.3-rc.1", "patch", "1.2.4"},
	}
	for _, tc := range tests {
		v, err := versioning.Parse(tc.start)
		if err != nil {
			t.Fatal(err)
		}
		got, err := v.Next(tc.kind)
		if err != nil {
			t.Errorf("Next(%q, %q): %v", tc.start, tc.kind, err)
		}
		if got.String() != tc.want {
			t.Errorf("Next(%q, %q) = %q, want %q", tc.start, tc.kind, got, tc.want)
		}
	}
}

func TestParseRejectsNonSemver(t *testing.T) {
	for _, bad := range []string{"notaversion", "1.2", "1.2.3.4", "1.2.3-rc", "1.2.3-rc.0", "01.2.3", "1.-2.3", "1.2.3-rc.01"} {
		if _, err := versioning.Parse(bad); err == nil {
			t.Errorf("Parse(%q): expected error, got nil", bad)
		}
	}
}

func TestNextUnknownKind(t *testing.T) {
	v, err := versioning.Parse("1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Next("hotfix"); err == nil {
		t.Fatal("expected error for unknown release kind")
	}
}

func TestParseAndCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"2.0.0-rc.1", "2.0.0-beta.2", 1},
		{"2.0.0", "2.0.0-rc.9", 1},
		{"2.0.0-rc.2", "2.0.0-rc.10", -1},
		{"2.1.0-alpha.1", "2.0.9", 1},
	}
	for _, tc := range tests {
		a, err := versioning.Parse(tc.a)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.a, err)
		}
		b, err := versioning.Parse(tc.b)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.b, err)
		}
		if got := a.Compare(b); got != tc.want {
			t.Fatalf("%s.Compare(%s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
