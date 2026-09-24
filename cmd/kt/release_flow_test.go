package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseNotesSinceLatestExcludesCurrentTag(t *testing.T) {
	dir := initReleaseRepo(t)
	gitTest(t, dir, "tag", "-a", "v1.4.0", "-m", "v1.4.0")
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("second\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, dir, "add", "notes.txt")
	gitTest(t, dir, "commit", "-m", "second release change")
	gitTest(t, dir, "tag", "-a", "v1.5.0", "-m", "v1.5.0")

	withTestDir(t, dir)
	notes, err := releaseNotes([]string{"--since", "latest"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(notes, "second release change") || strings.Contains(notes, "initial release change") {
		t.Fatalf("release notes = %q", notes)
	}
}

func TestPushReleaseBranchPublishesHeadBeforeTag(t *testing.T) {
	dir := initReleaseRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	gitTest(t, ".", "init", "--bare", remote)
	gitTest(t, dir, "remote", "add", "origin", remote)

	withTestDir(t, dir)
	if err := pushReleaseBranch(); err != nil {
		t.Fatal(err)
	}
	local := strings.TrimSpace(gitTest(t, dir, "rev-parse", "HEAD"))
	remoteHead := strings.TrimSpace(gitTest(t, ".", "--git-dir", remote, "rev-parse", "refs/heads/main"))
	if remoteHead != local {
		t.Fatalf("remote main = %s, want %s", remoteHead, local)
	}
}

func initReleaseRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitTest(t, dir, "init")
	gitTest(t, dir, "config", "user.name", "Release Test")
	gitTest(t, dir, "config", "user.email", "release@example.invalid")
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("first\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, dir, "add", "notes.txt")
	gitTest(t, dir, "commit", "-m", "initial release change")
	gitTest(t, dir, "branch", "-M", "main")
	return dir
}

func withTestDir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
}

func gitTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
