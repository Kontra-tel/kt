package ktconfig

import "testing"

func TestValidateAcceptsServiceProject(t *testing.T) {
	issues := Validate(Project{App: "my-api", Kind: "service", Services: "my-api", User: "my-api", Group: "my-api"})
	if len(issues) != 0 {
		t.Fatalf("Validate returned issues: %v", issues)
	}
}

func TestValidateRejectsUnsafeAppName(t *testing.T) {
	issues := Validate(Project{App: "bad/name", Kind: "cli"})
	if len(issues) == 0 {
		t.Fatal("expected unsafe app name issue")
	}
}

func TestValidateRejectsCLIServices(t *testing.T) {
	issues := Validate(Project{App: "tool", Kind: "cli", Services: "tool"})
	if len(issues) == 0 {
		t.Fatal("expected cli services issue")
	}
}
