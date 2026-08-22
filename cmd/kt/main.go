package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.kontra.tel/kontra.tel/Kt/internal/assets"
	"git.kontra.tel/kontra.tel/Kt/internal/deploycheck"
	"git.kontra.tel/kontra.tel/Kt/internal/ktconfig"
	"git.kontra.tel/kontra.tel/Kt/internal/scaffold"
	"git.kontra.tel/kontra.tel/Kt/internal/tui"
	"git.kontra.tel/kontra.tel/Kt/internal/updater"
	"git.kontra.tel/kontra.tel/Kt/internal/versioning"
)

var (
	version    = "dev"
	commit     = "unknown"
	date       = "unknown"
	releaseAPI = "https://git.kontra.tel/api/v1/repos/kontra.tel/Kt"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	s := scaffold.Scaffolder{FS: assets.FS}
	switch os.Args[1] {
	case "init":
		cmdInit(s, os.Args[2:])
	case "templates":
		cmdTemplates(s)
	case "install-tools":
		cmdInstallTools(s, os.Args[2:], false)
	case "update-tools":
		cmdInstallTools(s, os.Args[2:], true)
	case "config":
		cmdConfig(os.Args[2:])
	case "deploy":
		cmdDeploy(os.Args[2:])
	case "release":
		cmdRelease(os.Args[2:])
	case "doctor":
		cmdDoctor()
	case "update":
		cmdUpdate(os.Args[2:])
	case "version":
		cmdVersion()
	case "help", "--help", "-h":
		usage()
	default:
		tui.Err("unknown command: " + os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	tui.Title("kt", "project scaffolding for Make, nFPM, systemd, and tag-based releases")

	tui.Header("Project")
	tui.Table([]string{"command", "description"}, [][]string{
		{"kt init <template> <app> [--dir .] [--force]", "create a new project"},
		{"kt templates", "list scaffold templates"},
		{"kt config get|set|show|shape|validate", "inspect and validate .kt/project.yaml"},
		{"kt config init|diff|check", "manage deploy/config examples"},
		{"kt deploy inspect [--json]", "show deploy metadata"},
		{"kt deploy check [--json]", "validate deploy files"},
	})

	tui.Header("Release")
	tui.Table([]string{"command", "description"}, [][]string{
		{"kt release next <patch|minor|major|pre|stable> [--pre rc]", "print next version"},
		{"kt release plan <kind|version> [--pre rc] [--json]", "preview tag, dirty state, conflicts"},
		{"kt release validate <vversion> [--github-output]", "validate CI release tag"},
		{"kt release tag|push <version>", "create immutable annotated tag"},
	})

	tui.Header("Tooling")
	tui.Table([]string{"command", "description"}, [][]string{
		{"kt install-tools [--dir .] [--force]", "install .kt/mk helpers"},
		{"kt update-tools [--dir .] [--force]", "refresh .kt/mk helpers"},
		{"kt update [--check] [--prerelease]", "update kt itself"},
		{"kt doctor", "run project doctor checks"},
		{"kt version", "print build metadata"},
	})

	tui.Header("Examples")
	fmt.Println("  kt init service my-api")
	fmt.Println("  kt release plan minor")
	fmt.Println("  kt deploy check")
}

func cmdInit(s scaffold.Scaffolder, args []string) {
	var positional []string
	dir := "."
	force := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--dir":
			i++
			if i < len(args) {
				dir = args[i]
			}
		case "--force":
			force = true
		default:
			positional = append(positional, a)
		}
	}

	var tmplName, appName string
	if len(positional) >= 2 {
		tmplName, appName = positional[0], positional[1]
	} else {
		tmplName, appName = promptInit(s, positional)
	}
	if !ktconfig.SafeName(appName) {
		tui.Err("app name must match [a-z0-9][a-z0-9-]*")
		os.Exit(1)
	}

	ctx := scaffold.Context{Template: tmplName, App: appName}
	tui.Header("Initializing " + ctx.App)
	if err := s.Init(dir, ctx, force); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	if tmplName == "app" {
		tui.Warn("template 'app' is kept for compatibility; prefer 'service' for new projects")
	}
	tui.OK("created project structure")
	tui.Info("next: " + initNextHint(dir))
}

func promptInit(s scaffold.Scaffolder, positional []string) (tmplName, appName string) {
	infos, err := s.TemplatesWithDesc()
	if err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	labels := make([]string, len(infos))
	maxLen := 0
	for _, t := range infos {
		if len(t.Name) > maxLen {
			maxLen = len(t.Name)
		}
	}
	for i, t := range infos {
		labels[i] = fmt.Sprintf("%-*s  %s", maxLen, t.Name, tui.Muted(t.Desc))
	}
	idx := tui.Select("Choose a template", labels)
	tmplName = infos[idx].Name

	if len(positional) >= 1 {
		appName = positional[0]
	} else {
		appName = tui.Input("App name", "")
		for appName == "" {
			tui.Err("app name is required")
			appName = tui.Input("App name", "")
		}
	}
	return
}

func cmdTemplates(s scaffold.Scaffolder) {
	tui.Header("Available templates")
	infos, err := s.TemplatesWithDesc()
	if err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	rows := make([][]string, 0, len(infos))
	for _, t := range infos {
		rows = append(rows, []string{t.Name, t.Desc})
	}
	tui.Table([]string{"template", "description"}, rows)
}

func cmdInstallTools(s scaffold.Scaffolder, args []string, update bool) {
	name := "install-tools"
	header := "Installing local kt tooling"
	success := "installed .kt/mk"
	forceDefault := false
	forceUsage := "overwrite existing files"
	if update {
		name = "update-tools"
		header = "Updating local kt tooling"
		success = "updated .kt/mk"
		forceDefault = true
		forceUsage = "overwrite existing files (default true for updates)"
	}
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	dir := fs.String("dir", ".", "target directory")
	force := fs.Bool("force", forceDefault, forceUsage)
	_ = fs.Parse(args)
	tui.Header(header)
	if err := s.InstallTools(*dir, *force); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	tui.OK(success)
}

func cmdConfig(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt config get <key> | set <key> <value> | show [--json] | shape | validate | init|diff|check")
		os.Exit(2)
	}
	switch args[0] {
	case "get":
		if len(args) < 2 {
			tui.Err("usage: kt config get <key>")
			os.Exit(2)
		}
		val, err := ktconfig.Get(args[1])
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		fmt.Println(val)
	case "set":
		if len(args) < 3 {
			tui.Err("usage: kt config set <key> <value>")
			os.Exit(2)
		}
		if err := ktconfig.Set(args[1], args[2]); err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tui.OK(args[1] + " = " + args[2])
	case "show":
		if len(args) > 1 && args[1] == "--json" {
			project, err := ktconfig.Load()
			if err != nil {
				tui.Err(err.Error())
				os.Exit(1)
			}
			out := map[string]any{
				"schema":   project.Schema,
				"template": project.Template,
				"app":      project.App,
				"kind":     project.Kind,
				"services": project.ServicesList(),
				"user":     project.User,
				"group":    project.Group,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(out); err != nil {
				tui.Err(err.Error())
				os.Exit(1)
			}
			return
		}
		pairs, err := ktconfig.All()
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tui.Header("Project config")
		for _, p := range pairs {
			tui.Info(p[0] + ": " + p[1])
		}
	case "shape":
		project, err := ktconfig.Load()
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tui.Header("Project shape")
		services := "none"
		if project.HasServices() {
			services = strings.Join(project.ServicesList(), ", ")
		}
		rows := [][]string{
			{"schema", project.Schema},
			{"template", project.Template},
			{"app", project.App},
			{"kind", project.Kind},
			{"services", services},
		}
		if project.User != "" {
			rows = append(rows, []string{"user", project.User})
		}
		if project.Group != "" {
			rows = append(rows, []string{"group", project.Group})
		}
		tui.Table([]string{"field", "value"}, rows)
	case "validate":
		project, err := ktconfig.Load()
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		issues := ktconfig.Validate(project)
		if len(issues) == 0 {
			tui.OK("project manifest is valid")
			return
		}
		for _, issue := range issues {
			tui.Err(issue)
		}
		os.Exit(1)
	default:
		runMake("config-" + args[0])
	}
}

func cmdDeploy(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt deploy inspect [--json] | check [--json]")
		os.Exit(2)
	}
	jsonOut := len(args) > 1 && args[1] == "--json"
	switch args[0] {
	case "inspect":
		info, err := deploycheck.Inspect(".")
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		if jsonOut {
			writeJSON(info)
			return
		}
		tui.Header("Deploy")
		tui.Table([]string{"field", "value"}, [][]string{
			{"app", info.App},
			{"kind", info.Kind},
			{"services", strings.Join(info.Services, ", ")},
			{"config", info.ConfigDir},
			{"data", info.DataDir},
			{"logs", info.LogDir},
		})
	case "check":
		checks, err := deploycheck.CheckProject(".")
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		if jsonOut {
			writeJSON(map[string]any{"ok": !deploycheck.HasErrors(checks), "checks": checks})
			if deploycheck.HasErrors(checks) {
				os.Exit(1)
			}
			return
		}
		tui.Header("Deploy check")
		for _, check := range checks {
			msg := check.Name
			if check.Path != "" {
				msg += "  " + tui.Muted(check.Path)
			}
			if check.Message != "" {
				msg += " — " + check.Message
			}
			switch check.Level {
			case "ok":
				tui.OK(msg)
			case "warn":
				tui.Warn(msg)
			default:
				tui.Err(msg)
			}
		}
		if deploycheck.HasErrors(checks) {
			os.Exit(1)
		}
	default:
		tui.Err("usage: kt deploy inspect [--json] | check [--json]")
		os.Exit(2)
	}
}

func cmdRelease(args []string) {
	if len(args) < 1 {
		releaseUsage()
		os.Exit(2)
	}
	switch args[0] {
	case "next":
		opts := parseReleaseOptions(args[2:])
		if len(args) < 2 || opts.err != nil {
			if opts.err != nil {
				tui.Err(opts.err.Error())
			}
			releaseUsage()
			os.Exit(2)
		}
		next, err := nextRelease(args[1], opts.preLabel)
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		fmt.Println(next.String())
	case "plan":
		opts := parseReleaseOptions(args[2:])
		if len(args) < 2 || opts.err != nil {
			if opts.err != nil {
				tui.Err(opts.err.Error())
			}
			releaseUsage()
			os.Exit(2)
		}
		plan, err := buildReleasePlan(args[1], opts.preLabel)
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		if opts.json {
			writeJSON(plan)
			return
		}
		printReleasePlan(plan)
	case "validate":
		opts := parseReleaseOptions(args[2:])
		if len(args) < 2 || opts.err != nil {
			if opts.err != nil {
				tui.Err(opts.err.Error())
			}
			releaseUsage()
			os.Exit(2)
		}
		v, err := parseReleaseTag(args[1])
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		prerelease := v.Pre != ""
		if opts.githubOutput {
			if err := appendGitHubOutput(v.String(), prerelease); err != nil {
				tui.Err(err.Error())
				os.Exit(1)
			}
		}
		tui.OK("valid release tag v" + v.String())
	case "tag", "push":
		if len(args) != 2 {
			releaseUsage()
			os.Exit(2)
		}
		v, err := versioning.Parse(args[1])
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tag := "v" + v.String()
		if err := createReleaseTag(tag); err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		if args[0] == "push" {
			if err := gitRun("push", "origin", tag); err != nil {
				tui.Err(err.Error())
				os.Exit(1)
			}
		}
		tui.OK("created " + tag)
	default:
		releaseUsage()
		os.Exit(2)
	}
}

type releaseOptions struct {
	preLabel     string
	json         bool
	githubOutput bool
	err          error
}

type releasePlan struct {
	Current     string `json:"current"`
	Next        string `json:"next"`
	Tag         string `json:"tag"`
	Prerelease  bool   `json:"prerelease"`
	Dirty       bool   `json:"dirty"`
	LocalTag    string `json:"local_tag"`
	RemoteTag   string `json:"remote_tag"`
	RemoteError string `json:"remote_error,omitempty"`
}

func releaseUsage() {
	tui.Err("usage: kt release next <patch|minor|major|pre|stable> [--pre rc] | plan <patch|minor|major|version> [--pre rc] [--json] | validate <vversion> [--github-output] | tag|push <version>")
}

func parseReleaseOptions(args []string) releaseOptions {
	var opts releaseOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--pre":
			i++
			if i >= len(args) || strings.TrimSpace(args[i]) == "" {
				opts.err = fmt.Errorf("--pre requires a label")
				return opts
			}
			opts.preLabel = args[i]
		case "--json":
			opts.json = true
		case "--github-output":
			opts.githubOutput = true
		default:
			opts.err = fmt.Errorf("unknown release option %q", args[i])
			return opts
		}
	}
	return opts
}

func nextRelease(kind, preLabel string) (versioning.Version, error) {
	current, err := latestReleaseVersion()
	if err != nil {
		return versioning.Version{}, err
	}
	switch kind {
	case "pre":
		if current.Pre == "" {
			return versioning.Version{}, fmt.Errorf("latest release is stable; use patch|minor|major --pre <label>")
		}
		current.PreN++
		return current, nil
	case "stable":
		if current.Pre == "" {
			return versioning.Version{}, fmt.Errorf("latest release is already stable")
		}
		current.Pre = ""
		current.PreN = 0
		return current, nil
	default:
		next, err := current.Next(kind)
		if err != nil {
			return versioning.Version{}, err
		}
		if preLabel != "" {
			next.Pre = preLabel
			next.PreN = 1
		}
		return next, nil
	}
}

func buildReleasePlan(input, preLabel string) (releasePlan, error) {
	current, err := latestReleaseVersion()
	if err != nil {
		return releasePlan{}, err
	}
	next, err := versioning.Parse(input)
	if err != nil {
		next, err = nextRelease(input, preLabel)
		if err != nil {
			return releasePlan{}, err
		}
	} else if preLabel != "" {
		next.Pre = preLabel
		next.PreN = 1
	}
	tag := "v" + next.String()
	dirty, _ := gitOutput("status", "--porcelain")
	plan := releasePlan{
		Current:    current.String(),
		Next:       next.String(),
		Tag:        tag,
		Prerelease: next.Pre != "",
		Dirty:      strings.TrimSpace(dirty) != "",
		LocalTag:   "available",
		RemoteTag:  "available",
	}
	if err := gitRunQuiet("show-ref", "--verify", "--quiet", "refs/tags/"+tag); err == nil {
		plan.LocalTag = "exists"
	}
	remote, err := gitOutput("ls-remote", "--tags", "origin", "refs/tags/"+tag)
	if err != nil {
		plan.RemoteTag = "unknown"
		plan.RemoteError = err.Error()
	} else if strings.TrimSpace(remote) != "" {
		plan.RemoteTag = "exists"
	}
	return plan, nil
}

func printReleasePlan(plan releasePlan) {
	tui.Header("Release plan")
	tui.Table([]string{"field", "value"}, [][]string{
		{"current", plan.Current},
		{"next", plan.Next},
		{"tag", plan.Tag},
		{"prerelease", fmt.Sprintf("%t", plan.Prerelease)},
		{"dirty", fmt.Sprintf("%t", plan.Dirty)},
		{"local tag", plan.LocalTag},
		{"remote tag", plan.RemoteTag},
	})
	if plan.RemoteError != "" {
		tui.Warn("remote tag check: " + plan.RemoteError)
	}
	tui.Info("next: kt release push " + plan.Next)
}

func parseReleaseTag(tag string) (versioning.Version, error) {
	if !strings.HasPrefix(tag, "v") {
		return versioning.Version{}, fmt.Errorf("release tag must be v<semver>")
	}
	return versioning.Parse(strings.TrimPrefix(tag, "v"))
}

func appendGitHubOutput(version string, prerelease bool) error {
	out := os.Getenv("GITHUB_OUTPUT")
	if out == "" {
		fmt.Printf("version=%s\nprerelease=%t\n", version, prerelease)
		return nil
	}
	f, err := os.OpenFile(out, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "version=%s\nprerelease=%t\n", version, prerelease)
	return err
}

func latestReleaseVersion() (versioning.Version, error) {
	out, err := gitOutput("for-each-ref", "--merged=HEAD", "--sort=-version:refname", "--format=%(refname:strip=2)", "refs/tags/v*")
	if err != nil {
		return versioning.Version{}, err
	}
	for _, tag := range strings.Fields(out) {
		v, err := versioning.Parse(tag)
		if err == nil {
			return v, nil
		}
	}
	return versioning.Version{}, fmt.Errorf("no semver release tag is reachable from HEAD")
}

func createReleaseTag(tag string) error {
	dirty, err := gitOutput("status", "--porcelain")
	if err != nil {
		return err
	}
	if dirty != "" {
		return fmt.Errorf("working tree must be clean before creating a release tag")
	}
	if err := gitRun("rev-parse", "--verify", "HEAD"); err != nil {
		return fmt.Errorf("HEAD is not a commit: %w", err)
	}
	if err := gitRun("show-ref", "--verify", "--quiet", "refs/tags/"+tag); err == nil {
		return fmt.Errorf("tag %s already exists locally", tag)
	}
	remote, err := gitOutput("ls-remote", "--tags", "origin", "refs/tags/"+tag)
	if err != nil {
		return fmt.Errorf("check remote tag: %w", err)
	}
	if strings.TrimSpace(remote) != "" {
		return fmt.Errorf("tag %s already exists on origin", tag)
	}
	return gitRun("tag", "-a", tag, "-m", "Release "+tag)
}

func gitOutput(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

func gitRun(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitRunQuiet(args ...string) error {
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func cmdDoctor() { runMake("doctor") }

func cmdUpdate(args []string) {
	checkOnly := false
	includePrerelease := false
	for _, arg := range args {
		switch arg {
		case "--check":
			checkOnly = true
		case "--prerelease":
			includePrerelease = true
		default:
			tui.Err("usage: kt update [--check] [--prerelease]")
			os.Exit(2)
		}
	}
	if checkOnly && includePrerelease {
		tui.Err("usage: kt update --check | kt update [--prerelease]")
		os.Exit(2)
	}

	if version == "dev" {
		tui.Warn("skipping update check for dev build")
		return
	}

	// If the install location isn't writable, re-exec transparently with sudo.
	if !checkOnly {
		if exe, err := updater.ExecutablePath(); err == nil {
			if !canWriteDir(filepath.Dir(exe)) {
				sudoArgs := []string{exe, "update"}
				if includePrerelease {
					sudoArgs = append(sudoArgs, "--prerelease")
				}
				cmd := exec.Command("sudo", sudoArgs...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				if err := cmd.Run(); err != nil {
					os.Exit(1)
				}
				return
			}
		}
	}

	tui.Header("Checking for updates")
	latest, newer, err := updater.Check(releaseAPI, version, includePrerelease)
	if err != nil {
		tui.Err("check failed: " + err.Error())
		os.Exit(1)
	}
	if !newer {
		tui.OK("already up to date (" + version + ")")
		if checkOnly {
			if preLatest, preNewer, preErr := updater.Check(releaseAPI, version, true); preErr == nil && preNewer && preLatest != latest {
				tui.Info("prerelease available: " + preLatest + " (use kt update --prerelease)")
			}
		}
		return
	}
	tui.Info("new version available: " + latest + " (current: " + version + ")")
	if includePrerelease {
		tui.Info("channel: prerelease enabled")
	} else if checkOnly {
		if preLatest, preNewer, preErr := updater.Check(releaseAPI, version, true); preErr == nil && preNewer && preLatest != latest {
			tui.Info("new prerelease also available: " + preLatest + " (use kt update --prerelease)")
		}
	}

	if checkOnly {
		os.Exit(1)
	}

	tui.Header("Updating")
	if err := updater.Apply(releaseAPI, includePrerelease); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	tui.OK("updated to " + latest + " — restart kt to use the new version")
}

func canWriteDir(dir string) bool {
	tmp, err := os.CreateTemp(dir, ".kt-write-check-*")
	if err != nil {
		return false
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return true
}

func cmdVersion() {
	tui.Header("kt")
	tui.Info("version: " + version)
	tui.Info("commit:  " + commit)
	tui.Info("date:    " + date)
}

func runMake(target string) {
	if _, err := os.Stat("Makefile"); err != nil {
		tui.Err("Makefile not found in " + mustGetwd())
		os.Exit(1)
	}
	cmd := exec.Command("make", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func mustGetwd() string { wd, _ := os.Getwd(); return wd }

func initNextHint(dir string) string {
	projectFile := filepath.Join(dir, ".kt", "project.yaml")
	project, err := ktconfig.LoadFile(projectFile)
	if err != nil {
		return "make doctor && make build && make package"
	}
	var steps []string
	if dir != "." {
		steps = append(steps, "cd "+dir)
	}
	steps = append(steps, "make doctor", "make build")
	switch project.Kind {
	case "cli":
		steps = append(steps, "make run")
	case "service", "multi-service", "mixed":
		steps = append(steps, "make print-info")
	}
	steps = append(steps, "make package")
	return strings.Join(steps, " && ")
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
}
