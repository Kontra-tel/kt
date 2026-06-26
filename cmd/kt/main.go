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
		cmdInstallTools(s, os.Args[2:])
	case "update-tools":
		cmdInstallTools(s, os.Args[2:])
	case "config":
		cmdConfig(os.Args[2:])
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
	banner := `${CYAN}kt${RESET} - tiny project scaffolding for Make + nFPM + systemd

Usage:
  kt templates
  kt init <template> <app> [--dir .] [--force]
  kt install-tools [--dir .] [--force]
  kt update-tools [--dir .] [--force]
  kt config get <key>
  kt config set <key> <value>
  kt config show [--json]
  kt config shape
  kt config init|diff|check
  kt release patch|minor|major
  kt release set <version>
  kt update [--check] [--prerelease]
  kt doctor
  kt version

Examples:
  kt init service my-api
  kt init cli my-tool
  kt init mixed my-suite
  kt init multi my-platform
  make build
  make package`
	banner = strings.ReplaceAll(banner, "${CYAN}", tui.Cyan+tui.Bold)
	banner = strings.ReplaceAll(banner, "${RESET}", tui.Reset)
	fmt.Println(banner)
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
		labels[i] = fmt.Sprintf("%-*s  %s%s%s", maxLen, t.Name, tui.Dim, t.Desc, tui.Reset)
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
	maxLen := 0
	for _, t := range infos {
		if len(t.Name) > maxLen {
			maxLen = len(t.Name)
		}
	}
	for _, t := range infos {
		tui.Info(fmt.Sprintf("%-*s  %s", maxLen, t.Name, t.Desc))
	}
}

func cmdInstallTools(s scaffold.Scaffolder, args []string) {
	fs := flag.NewFlagSet("install-tools", flag.ExitOnError)
	dir := fs.String("dir", ".", "target directory")
	force := fs.Bool("force", false, "overwrite existing files")
	_ = fs.Parse(args)
	tui.Header("Installing local kt tooling")
	if err := s.InstallTools(*dir, *force); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	tui.OK("installed .kt/mk and .kt/scripts")
}

func cmdConfig(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt config get <key> | set <key> <value> | show [--json] | shape | init|diff|check")
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
		tui.Info("template: " + project.Template)
		tui.Info("app:      " + project.App)
		tui.Info("kind:     " + project.Kind)
		if project.HasServices() {
			tui.Info("services: " + strings.Join(project.ServicesList(), ", "))
		} else {
			tui.Info("services: none")
		}
		if project.User != "" {
			tui.Info("user:     " + project.User)
		}
		if project.Group != "" {
			tui.Info("group:    " + project.Group)
		}
	default:
		runMake("config-" + args[0])
	}
}

func cmdRelease(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt release patch|minor|major | set <version>")
		os.Exit(2)
	}
	switch args[0] {
	case "patch", "minor", "major":
		v, err := versioning.Bump("version.txt", args[0])
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tui.OK("new version: " + v)
	case "set":
		if len(args) < 2 {
			tui.Err("usage: kt release set <version>")
			os.Exit(2)
		}
		v, err := versioning.Set("version.txt", args[1])
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tui.OK("new version: " + v)
	default:
		tui.Err("usage: kt release patch|minor|major | set <version>")
		os.Exit(2)
	}
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
