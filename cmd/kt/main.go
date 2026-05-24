package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.kontra.tel/kontra.tel/build-tools/internal/assets"
	"git.kontra.tel/kontra.tel/build-tools/internal/ktconfig"
	"git.kontra.tel/kontra.tel/build-tools/internal/scaffold"
	"git.kontra.tel/kontra.tel/build-tools/internal/tui"
	"git.kontra.tel/kontra.tel/build-tools/internal/updater"
	"git.kontra.tel/kontra.tel/build-tools/internal/versioning"
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
  kt config show
  kt config init|diff|check
  kt release patch|minor|major
  kt update [--check]
  kt doctor
  kt version

Examples:
  kt init java-service my-api
  kt init multi-service my-app
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
	tui.OK("created project structure")
	tui.Info("next: cd " + appName + " && make doctor && make build && make package")
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
		tui.Err("usage: kt config get <key> | set <key> <value> | show | init|diff|check")
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
		pairs, err := ktconfig.All()
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		tui.Header("Project config")
		for _, p := range pairs {
			tui.Info(p[0] + ": " + p[1])
		}
	default:
		runMake("config-" + args[0])
	}
}

func cmdRelease(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt release patch|minor|major")
		os.Exit(2)
	}
	v, err := versioning.Bump("version.txt", args[0])
	if err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	tui.OK("new version: " + v)
}

func cmdDoctor() { runMake("doctor") }

func cmdUpdate(args []string) {
	checkOnly := len(args) > 0 && args[0] == "--check"

	if version == "dev" {
		tui.Warn("skipping update check for dev build")
		return
	}

	// If the install location isn't writable, re-exec transparently with sudo.
	if !checkOnly {
		if exe, err := updater.ExecutablePath(); err == nil {
			if !canWriteDir(filepath.Dir(exe)) {
				cmd := exec.Command("sudo", exe, "update")
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
	latest, newer, err := updater.Check(releaseAPI, version)
	if err != nil {
		tui.Err("check failed: " + err.Error())
		os.Exit(1)
	}
	if !newer {
		tui.OK("already up to date (" + version + ")")
		return
	}
	tui.Info("new version available: " + latest + " (current: " + version + ")")

	if checkOnly {
		os.Exit(1)
	}

	tui.Header("Updating")
	if err := updater.Apply(releaseAPI); err != nil {
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
