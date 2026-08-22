package main

import "testing"

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

func TestParseReleaseOptions(t *testing.T) {
	opts := parseReleaseOptions([]string{"--pre", "rc", "--json"})
	if opts.err != nil {
		t.Fatal(opts.err)
	}
	if opts.preLabel != "rc" || !opts.json {
		t.Fatalf("opts = %+v", opts)
	}
}
