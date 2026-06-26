package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveBuildMetadata_PrefersBuildInfoVersion(t *testing.T) {
	orig := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v1.3.0-rc.1"},
		}, true
	}
	t.Cleanup(func() { readBuildInfo = orig })

	version, commit, date := resolveBuildMetadata("dev", "unknown", "unknown")
	if version != "v1.3.0-rc.1" {
		t.Fatalf("version = %q", version)
	}
	if commit != "unknown" || date != "unknown" {
		t.Fatalf("commit/date = %q %q", commit, date)
	}
}

func TestResolveBuildMetadata_UsesVCSSettings(t *testing.T) {
	orig := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "(devel)"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.time", Value: "2026-06-26T12:00:00Z"},
			},
		}, true
	}
	t.Cleanup(func() { readBuildInfo = orig })

	version, commit, date := resolveBuildMetadata("dev", "unknown", "unknown")
	if version != "dev" {
		t.Fatalf("version = %q", version)
	}
	if commit != "abc123" || date != "2026-06-26T12:00:00Z" {
		t.Fatalf("commit/date = %q %q", commit, date)
	}
}
