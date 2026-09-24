package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseReleaseTagRequiresPrefix(t *testing.T) {
	if _, err := parseReleaseTag("1.4.0"); err == nil {
		t.Fatal("expected missing v prefix error")
	}
}

func TestParseReleaseTagAcceptsPrerelease(t *testing.T) {
	v, err := parseReleaseTag("v1.4.0-rc.1")
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "1.4.0-rc.1" {
		t.Fatalf("version = %q", v.String())
	}
}

func TestReleaseTagUsesProjectPrefix(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".kt"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".kt", "project.yaml"), []byte("template: cli\napp: tool\nkind: cli\nservices: []\nrelease:\n  tag_prefix: release-\n"), 0644); err != nil {
		t.Fatal(err)
	}
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	v, err := parseReleaseTag("release-1.5.0")
	if err != nil {
		t.Fatal(err)
	}
	if releaseTag(v) != "release-1.5.0" {
		t.Fatalf("release tag = %q", releaseTag(v))
	}
}

func TestParseReleaseOptions(t *testing.T) {
	opts := parseReleaseOptions([]string{"--pre", "rc", "--json"})
	if opts.err != nil {
		t.Fatal(opts.err)
	}
	if opts.preLabel != "rc" || !opts.json {
		t.Fatalf("opts = %+v", opts)
	}
}
