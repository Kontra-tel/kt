package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git.kontra.tel/kontra.tel/build-tools/internal/assets"
	"git.kontra.tel/kontra.tel/build-tools/internal/scaffold"
	"git.kontra.tel/kontra.tel/build-tools/internal/tui"
	"git.kontra.tel/kontra.tel/build-tools/internal/updater"
	"git.kontra.tel/kontra.tel/build-tools/internal/versioning"
)

var (
	version    = "dev"
	commit     = "unknown"
	date       = "unknown"
	releaseAPI = "https://git.kontra.tel/api/v1/repos/kontra.tel/kt"
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
  kt init <template> <app> [--dir .] [--port 8080] [--user app] [--group app]
  kt install-tools [--dir .] [--force]
  kt update-tools [--dir .] [--force]
  kt config init|diff|check
  kt release patch|minor|major
  kt update [--check]
  kt doctor
  kt version

Examples:
  kt init java-service kontra-api
  kt init multi-service knetlog --port 4002 --user kontra --group kontra
  make build
  make package`
	banner = strings.ReplaceAll(banner, "${CYAN}", tui.Cyan+tui.Bold)
	banner = strings.ReplaceAll(banner, "${RESET}", tui.Reset)
	fmt.Println(banner)
}

func cmdInit(s scaffold.Scaffolder, args []string) {
	var positional []string
	dir, port, user, group, author := ".", "8080", "", "", "Kontra"
	force := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--dir":
			i++
			if i < len(args) {
				dir = args[i]
			}
		case "--port":
			i++
			if i < len(args) {
				port = args[i]
			}
		case "--user":
			i++
			if i < len(args) {
				user = args[i]
			}
		case "--group":
			i++
			if i < len(args) {
				group = args[i]
			}
		case "--author":
			i++
			if i < len(args) {
				author = args[i]
			}
		case "--force":
			force = true
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) < 2 {
		tui.Err("usage: kt init <template> <app>")
		os.Exit(2)
	}
	ctx := scaffold.Context{Template: positional[0], App: positional[1], Port: port, ServiceUser: user, ServiceGroup: group, Author: author}
	tui.Header("Initializing " + ctx.App)
	if err := s.Init(dir, ctx, force); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	tui.OK("created project structure")
	tui.Info("next: make doctor && make build && make package")
}

func cmdTemplates(s scaffold.Scaffolder) {
	tui.Header("Available templates")
	t, err := s.Templates()
	if err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	for _, name := range t {
		tui.Info(name)
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
		tui.Err("usage: kt config init|diff|check")
		os.Exit(2)
	}
	runMake("config-" + args[0])
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
