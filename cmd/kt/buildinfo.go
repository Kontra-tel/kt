package main

import "runtime/debug"

var readBuildInfo = debug.ReadBuildInfo

func init() {
	version, commit, date = resolveBuildMetadata(version, commit, date)
}

func resolveBuildMetadata(defaultVersion, defaultCommit, defaultDate string) (string, string, string) {
	info, ok := readBuildInfo()
	if !ok {
		return defaultVersion, defaultCommit, defaultDate
	}

	version := defaultVersion
	if version == "" || version == "dev" {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	commit := defaultCommit
	date := defaultDate
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if commit == "" || commit == "unknown" {
				commit = setting.Value
			}
		case "vcs.time":
			if date == "" || date == "unknown" {
				date = setting.Value
			}
		}
	}

	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if date == "" {
		date = "unknown"
	}
	return version, commit, date
}
