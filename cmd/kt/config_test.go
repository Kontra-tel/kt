package main

import (
	"strings"
	"testing"

	"git.kontra.tel/kontra.tel/Kt/internal/ktconfig"
)

func TestRenderProjectYAML_MigratesLegacyServicesToStructured(t *testing.T) {
	project := ktconfig.Project{
		Template: "service",
		App:      "old-api",
		Kind:     "service",
		Services: "old-api",
		User:     "old",
		Group:    "old",
		Package:  ktconfig.PackageInfo{Name: "old-api"},
		Config:   ktconfig.ConfigInfo{Dir: "deploy/config", InstallDir: "/etc/old-api", ExampleSuffix: ".example"},
		Release:  ktconfig.ReleaseInfo{TagPrefix: "v"},
		KT:       ktconfig.KTInfo{ScaffoldVersion: "1.4"},
	}

	content := renderProjectYAML(project)
	for _, want := range []string{
		"schema: kt.project/v1",
		"services:\n  - name: old-api",
		"runner: deploy/run/old-api",
		"unit: deploy/systemd/old-api.service",
		"user: old",
		"group: old",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("renderProjectYAML missing %q:\n%s", want, content)
		}
	}
}

func TestProjectSchemaNamesKtProjectV1(t *testing.T) {
	schema := projectSchema()
	if schema["title"] != "kt.project/v1" {
		t.Fatalf("title = %v", schema["title"])
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties missing: %#v", schema["properties"])
	}
	if props["services"] == nil || props["config"] == nil || props["release"] == nil {
		t.Fatalf("schema missing structured properties: %#v", props)
	}
}
