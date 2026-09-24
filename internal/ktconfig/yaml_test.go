package ktconfig

import "testing"

func TestLoadSupportsQuotedAndMultilineYAML(t *testing.T) {
	setup(t, `schema: kt.project/v1
template: cli
app: tool
kind: cli
package:
  description: "tool: a command-line utility"
services: []
commands:
  - name: tool
    path: deploy/bin/tool
config:
  dir: deploy/config
  install_dir: /etc/tool
  example_suffix: .example
release:
  tag_prefix: release-
kt:
  scaffold_version: "1.5"
notes: |
  valid YAML with: a colon
`)

	project, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if project.Package.Description != "tool: a command-line utility" {
		t.Fatalf("description = %q", project.Package.Description)
	}
	prefix, err := Get("release.tag_prefix")
	if err != nil || prefix != "release-" {
		t.Fatalf("Get(release.tag_prefix) = %q, %v", prefix, err)
	}
}

func TestSetRejectsStructuredYAMLField(t *testing.T) {
	setup(t, "app: tool\npackage:\n  name: tool\n")
	if err := Set("package", "tool"); err == nil {
		t.Fatal("expected Set to reject a structured field")
	}
}
