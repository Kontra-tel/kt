package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	globalJSON bool
)

func main() {
	args := parseGlobalArgs(os.Args[1:])
	if len(args) < 1 {
		usage()
		return
	}
	s := scaffold.Scaffolder{FS: assets.FS}
	switch args[0] {
	case "init":
		cmdInit(s, args[1:])
	case "templates":
		cmdTemplates(s)
	case "completion":
		cmdCompletion(args[1:])
	case "install-tools":
		cmdInstallTools(s, args[1:], false)
	case "update-tools":
		cmdInstallTools(s, args[1:], true)
	case "config":
		cmdConfig(args[1:])
	case "deploy":
		cmdDeploy(args[1:])
	case "release":
		cmdRelease(args[1:])
	case "doctor":
		cmdDoctor()
	case "update":
		cmdUpdate(args[1:])
	case "version":
		cmdVersion()
	case "help", "--help", "-h":
		usageTopic(args[1:])
	default:
		suggestion := suggest(args[0], rootCommands())
		if suggestion != "" {
			tui.Err("unknown command: " + args[0] + " (did you mean " + suggestion + "?)")
		} else {
			tui.Err("unknown command: " + args[0])
		}
		usage()
		os.Exit(2)
	}
}

func parseGlobalArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--quiet", "-q":
			tui.SetQuiet(true)
		case "--no-color":
			tui.SetColor(false)
		case "--color":
			i++
			if i >= len(args) {
				tui.Err("--color requires auto, always, or never")
				os.Exit(2)
			}
			switch args[i] {
			case "never":
				tui.SetColor(false)
			case "always", "auto":
				tui.SetColor(true)
			default:
				tui.Err("--color requires auto, always, or never")
				os.Exit(2)
			}
		case "--json":
			globalJSON = true
			tui.SetColor(false)
		default:
			out = append(out, args[i])
		}
	}
	return out
}

func usage() {
	tui.Title("kt", "scaffold Make, nFPM, systemd, and tag-based releases")

	tui.Header("Usage")
	fmt.Println("  kt [global flags] <command> [arguments]")
	fmt.Println("  kt help [topic]")

	tui.Header("Commands")
	tui.Table([]string{"group", "commands"}, [][]string{
		{"project", "init, templates, config, deploy"},
		{"release", "release next|plan|notes|validate|tag|push"},
		{"tooling", "install-tools, update-tools, completion, update, doctor, version"},
	})

	tui.Header("Global flags")
	tui.Table([]string{"flag", "description"}, [][]string{
		{"--json", "machine-readable output where supported"},
		{"--quiet, -q", "suppress non-error status output"},
		{"--no-color", "disable styling"},
		{"--color auto|always|never", "control styling"},
	})

	fmt.Println("  details: kt help <topic>, man kt, or docs/commands.md")
}

func usageTopic(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	switch args[0] {
	case "init":
		tui.Header("kt init")
		fmt.Println("  kt init [template] [app] [--dir DIR] [--force] [--dry-run]")
		fmt.Println("  interactive when template or app is omitted")
	case "config":
		tui.Header("kt config")
		fmt.Println("  kt config get <key> | set <key> <value> | show [--json] | shape | validate")
		fmt.Println("  kt config edit | schema | migrate --to kt.project/v1")
		fmt.Println("  kt config init | diff | check")
	case "deploy":
		tui.Header("kt deploy")
		fmt.Println("  kt deploy inspect [--json]")
		fmt.Println("  kt deploy metadata [--json] [--output FILE]")
		fmt.Println("  kt deploy check [--json]")
	case "release":
		tui.Header("kt release")
		fmt.Println("  kt release next <patch|minor|major|pre|stable> [--pre rc]")
		fmt.Println("  kt release plan <kind|version> [--pre rc] [--json]")
		fmt.Println("  kt release notes [range|--since latest]")
		fmt.Println("  kt release validate <vversion> [--github-output]")
		fmt.Println("  kt release tag|push <version>")
	case "completion":
		tui.Header("kt completion")
		fmt.Println("  kt completion bash|zsh|fish")
	case "tools", "install-tools", "update-tools":
		tui.Header("kt tooling")
		fmt.Println("  kt install-tools [--dir .] [--force] [--check|--diff] [--apply]")
		fmt.Println("  kt update-tools [--dir .] [--force] [--check|--diff] [--apply]")
	default:
		tui.Err("unknown help topic: " + args[0])
		os.Exit(2)
	}
}

func rootCommands() []string {
	return []string{"init", "templates", "install-tools", "update-tools", "config", "deploy", "release", "completion", "doctor", "update", "version", "help"}
}

func cmdInit(s scaffold.Scaffolder, args []string) {
	var positional []string
	dir := "."
	dirSet := false
	force := false
	dryRun := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--dir":
			i++
			if i < len(args) {
				dir = args[i]
				dirSet = true
			}
		case "--force":
			force = true
		case "--dry-run":
			dryRun = true
		case "help", "--help", "-h":
			usageTopic([]string{"init"})
			return
		default:
			positional = append(positional, a)
		}
	}

	interactive := len(positional) < 2
	var tmplName, appName string
	if !interactive {
		tmplName, appName = positional[0], positional[1]
	} else {
		tmplName, appName = promptInit(s, positional)
		if !dirSet {
			dir = tui.Input("Target directory", dir)
		}
	}
	if !validTemplate(s, tmplName) {
		tui.Err("unknown template: " + tmplName)
		os.Exit(1)
	}
	if !ktconfig.SafeName(appName) {
		tui.Err("app name must match [a-z0-9][a-z0-9-]*")
		os.Exit(1)
	}

	ctx := scaffold.Context{Template: tmplName, App: appName}
	if interactive {
		ctx.Author = tui.Input("Package maintainer", defaultMaintainer())
		if templateHasService(tmplName) {
			ctx.ServiceUser = tui.Input("Service user", appName)
			ctx.ServiceGroup = tui.Input("Service group", ctx.ServiceUser)
		}
		tui.Header("Create project")
		tui.Table([]string{"field", "value"}, [][]string{
			{"template", tmplName},
			{"app", appName},
			{"directory", dir},
			{"maintainer", ctx.Author},
			{"service user", ctx.ServiceUser},
			{"service group", ctx.ServiceGroup},
		})
	}

	if dryRun {
		if err := dryRunInit(s, dir, ctx, force); err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		return
	}

	tui.Header("Initializing " + ctx.App)
	if err := s.Init(dir, ctx, force); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	if tmplName == "app" {
		tui.Warn("template 'app' is kept for compatibility; prefer 'service' for new projects")
	}
	tui.OK("created project structure")
	tui.Info(initNextHint(dir))
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
	if len(positional) >= 1 {
		tmplName = positional[0]
	} else {
		idx := tui.Select("Choose a template", labels)
		tmplName = infos[idx].Name
	}

	if len(positional) >= 2 {
		appName = positional[1]
	} else {
		for {
			appName = tui.Input("App name", "")
			if ktconfig.SafeName(appName) {
				break
			}
			tui.Err("app name must match [a-z0-9][a-z0-9-]*")
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

func cmdCompletion(args []string) {
	if len(args) != 1 {
		tui.Err("usage: kt completion bash|zsh|fish")
		os.Exit(2)
	}
	commands := strings.Join(rootCommands(), " ")
	switch args[0] {
	case "bash":
		fmt.Printf("_kt_complete() {\n  local cur=\"${COMP_WORDS[COMP_CWORD]}\"\n  COMPREPLY=( $(compgen -W %q -- \"$cur\") )\n}\ncomplete -F _kt_complete kt\n", commands)
	case "zsh":
		fmt.Printf("#compdef kt\n_arguments '1:command:(%s)'\n", commands)
	case "fish":
		for _, command := range rootCommands() {
			fmt.Printf("complete -c kt -f -a %s\n", command)
		}
	default:
		tui.Err("usage: kt completion bash|zsh|fish")
		os.Exit(2)
	}
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
	checkOnly := fs.Bool("check", false, "exit 1 if local .kt/mk differs from embedded tooling")
	diffOnly := fs.Bool("diff", false, "print diff between local .kt/mk and embedded tooling")
	apply := fs.Bool("apply", false, "apply update after check/diff")
	_ = fs.Parse(args)
	if *checkOnly || *diffOnly {
		changed, err := toolDiff(s, *dir, *diffOnly)
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		if changed {
			if *apply {
				tui.Warn("local .kt/mk differs; applying embedded tooling")
				if err := s.InstallTools(*dir, true); err != nil {
					tui.Err(err.Error())
					os.Exit(1)
				}
				tui.OK("updated .kt/mk")
				return
			}
			tui.Warn("local .kt/mk differs from embedded tooling")
			os.Exit(1)
		}
		tui.OK("local .kt/mk matches embedded tooling")
		return
	}
	tui.Header(header)
	if err := s.InstallTools(*dir, *force); err != nil {
		tui.Err(err.Error())
		os.Exit(1)
	}
	tui.OK(success)
}

func toolDiff(s scaffold.Scaffolder, dir string, print bool) (bool, error) {
	tmp, err := os.MkdirTemp("", "kt-tools-*")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmp)
	if err := s.InstallTools(tmp, true); err != nil {
		return false, err
	}
	want := filepath.Join(tmp, ".kt", "mk")
	got := filepath.Join(dir, ".kt", "mk")
	if _, err := os.Stat(got); err != nil {
		if print {
			fmt.Printf("Only in embedded tooling: .kt/mk\n")
		}
		return true, nil
	}
	cmd := exec.Command("diff", "-ru", got, want)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 && print {
		fmt.Print(string(out))
	}
	if err == nil {
		return false, nil
	}
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
		return true, nil
	}
	return true, err
}

func templateHasService(name string) bool {
	return name == "app" || name == "service" || name == "mixed" || name == "multi"
}

func defaultMaintainer() string {
	name, _ := exec.Command("git", "config", "user.name").Output()
	email, _ := exec.Command("git", "config", "user.email").Output()
	n := strings.TrimSpace(string(name))
	e := strings.TrimSpace(string(email))
	if n != "" && e != "" {
		return n + " <" + e + ">"
	}
	return n
}

func validTemplate(s scaffold.Scaffolder, name string) bool {
	templates, err := s.Templates()
	if err != nil {
		return false
	}
	for _, template := range templates {
		if template == name {
			return true
		}
	}
	return false
}

func dryRunInit(s scaffold.Scaffolder, dir string, ctx scaffold.Context, force bool) error {
	tmp, err := os.MkdirTemp("", "kt-init-dry-run-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := s.Init(tmp, ctx, true); err != nil {
		return err
	}
	var rows [][]string
	err = filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(tmp, path)
		if err != nil {
			return err
		}
		action := "create"
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			if force {
				action = "overwrite"
			} else {
				action = "keep"
			}
		}
		rows = append(rows, []string{action, rel})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][0] == rows[j][0] {
			return rows[i][1] < rows[j][1]
		}
		return rows[i][0] < rows[j][0]
	})
	tui.Header("Init dry run")
	tui.Table([]string{"action", "path"}, rows)
	return nil
}

func cmdConfig(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt config get <key> | set <key> <value> | show [--json] | shape | validate | edit | schema | migrate --to kt.project/v1 | init|diff|check")
		os.Exit(2)
	}
	switch args[0] {
	case "help", "--help", "-h":
		usageTopic([]string{"config"})
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
		if (len(args) > 1 && args[1] == "--json") || globalJSON {
			project, err := ktconfig.Load()
			if err != nil {
				tui.Err(err.Error())
				os.Exit(1)
			}
			writeJSON(projectJSON(project))
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
	case "edit":
		if err := editProjectConfig(); err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
	case "schema":
		writeJSON(projectSchema())
	case "migrate":
		if len(args) != 3 || args[1] != "--to" || args[2] != "kt.project/v1" {
			tui.Err("usage: kt config migrate --to kt.project/v1")
			os.Exit(2)
		}
		if err := migrateProjectConfig(); err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
	default:
		runMake("config-" + args[0])
	}
}

func projectJSON(project ktconfig.Project) map[string]any {
	return map[string]any{
		"schema":          project.Schema,
		"template":        project.Template,
		"app":             project.App,
		"kind":            project.Kind,
		"services":        project.ServicesList(),
		"service_details": project.ServiceDetails(),
		"commands":        project.Commands,
		"package":         project.Package,
		"config":          project.Config,
		"release":         project.Release,
		"kt":              project.KT,
		"user":            project.User,
		"group":           project.Group,
	}
}

func editProjectConfig() error {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command("sh", "-c", editor+" "+shellQuote(filepath.Join(".kt", "project.yaml")))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	project, err := ktconfig.Load()
	if err != nil {
		return err
	}
	issues := ktconfig.Validate(project)
	if len(issues) > 0 {
		return fmt.Errorf("project manifest has issues after edit: %s", strings.Join(issues, "; "))
	}
	tui.OK("project manifest is valid")
	return nil
}

func projectSchema() map[string]any {
	return map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title":   "kt.project/v1",
		"type":    "object",
		"required": []string{
			"schema", "template", "app", "kind",
		},
		"properties": map[string]any{
			"schema":   map[string]any{"const": "kt.project/v1"},
			"template": map[string]any{"type": "string"},
			"app":      map[string]any{"type": "string", "pattern": "^[a-z0-9][a-z0-9-]*$"},
			"kind":     map[string]any{"enum": []string{"cli", "service", "mixed", "multi-service"}},
			"package":  map[string]any{"type": "object"},
			"services": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
			"commands": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
			"config":   map[string]any{"type": "object"},
			"release":  map[string]any{"type": "object"},
			"kt":       map[string]any{"type": "object"},
		},
	}
}

func migrateProjectConfig() error {
	project, err := ktconfig.Load()
	if err != nil {
		return err
	}
	content := renderProjectYAML(project)
	if err := os.WriteFile(filepath.Join(".kt", "project.yaml"), []byte(content), 0644); err != nil {
		return err
	}
	tui.OK("migrated .kt/project.yaml to kt.project/v1")
	return nil
}

func renderProjectYAML(project ktconfig.Project) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Project contract used by kt, Make, and packaging.\n")
	fmt.Fprintf(&b, "# Safe to edit after scaffold.\n")
	fmt.Fprintf(&b, "schema: kt.project/v1\n")
	fmt.Fprintf(&b, "template: %s\n", project.Template)
	fmt.Fprintf(&b, "app: %s\n", project.App)
	fmt.Fprintf(&b, "kind: %s\n", project.Kind)
	fmt.Fprintf(&b, "package:\n")
	fmt.Fprintf(&b, "  name: %s\n", nonEmpty(project.Package.Name, project.App))
	if project.Package.Maintainer != "" {
		fmt.Fprintf(&b, "  maintainer: %s\n", project.Package.Maintainer)
	}
	if project.Package.Description != "" {
		fmt.Fprintf(&b, "  description: %s\n", project.Package.Description)
	}
	if project.Package.Section != "" {
		fmt.Fprintf(&b, "  section: %s\n", project.Package.Section)
	}
	if project.Package.License != "" {
		fmt.Fprintf(&b, "  license: %s\n", project.Package.License)
	}
	services := project.ServiceDetails()
	if len(services) == 0 {
		fmt.Fprintf(&b, "services: []\n")
	} else {
		fmt.Fprintf(&b, "services:\n")
		for _, service := range services {
			service = serviceWithDefaults(project, service)
			fmt.Fprintf(&b, "  - name: %s\n", service.Name)
			if service.Role != "" {
				fmt.Fprintf(&b, "    role: %s\n", service.Role)
			}
			if service.Runner != "" {
				fmt.Fprintf(&b, "    runner: %s\n", service.Runner)
			}
			if service.Unit != "" {
				fmt.Fprintf(&b, "    unit: %s\n", service.Unit)
			}
			if service.User != "" {
				fmt.Fprintf(&b, "    user: %s\n", service.User)
			}
			if service.Group != "" {
				fmt.Fprintf(&b, "    group: %s\n", service.Group)
			}
		}
	}
	commands := project.Commands
	if len(commands) == 0 && project.App != "" {
		commands = []ktconfig.Command{{Name: project.App, Path: "deploy/bin/" + project.App}}
		if project.Kind == "mixed" {
			commands = append(commands, ktconfig.Command{Name: project.App + "-service", Path: "deploy/bin/" + project.App + "-service"})
		}
	}
	if len(commands) > 0 {
		fmt.Fprintf(&b, "commands:\n")
		for _, command := range commands {
			fmt.Fprintf(&b, "  - name: %s\n", command.Name)
			fmt.Fprintf(&b, "    path: %s\n", command.Path)
		}
	}
	fmt.Fprintf(&b, "config:\n")
	fmt.Fprintf(&b, "  dir: %s\n", project.Config.Dir)
	fmt.Fprintf(&b, "  install_dir: %s\n", project.Config.InstallDir)
	fmt.Fprintf(&b, "  example_suffix: %s\n", project.Config.ExampleSuffix)
	fmt.Fprintf(&b, "release:\n")
	fmt.Fprintf(&b, "  tag_prefix: %s\n", project.Release.TagPrefix)
	fmt.Fprintf(&b, "kt:\n")
	fmt.Fprintf(&b, "  scaffold_version: \"%s\"\n", project.KT.ScaffoldVersion)
	return b.String()
}

func nonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func hasArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func serviceWithDefaults(project ktconfig.Project, service ktconfig.Service) ktconfig.Service {
	if service.Runner == "" && service.Name != "" {
		service.Runner = filepath.Join("deploy", "run", service.Name)
	}
	if service.Unit == "" && service.Name != "" {
		service.Unit = filepath.Join("deploy", "systemd", service.Name+".service")
	}
	if service.User == "" {
		service.User = project.User
	}
	if service.Group == "" {
		service.Group = project.Group
	}
	return service
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
func cmdDeploy(args []string) {
	if len(args) < 1 {
		tui.Err("usage: kt deploy inspect [--json] | metadata [--json] [--output FILE] | check [--json]")
		os.Exit(2)
	}
	jsonOut := globalJSON || hasArg(args[1:], "--json")
	switch args[0] {
	case "help", "--help", "-h":
		usageTopic([]string{"deploy"})
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
			{"services", strings.Join(serviceNames(info.Services), ", ")},
			{"config", info.ConfigDir},
			{"data", info.DataDir},
			{"logs", info.LogDir},
		})
	case "metadata":
		info, err := deploycheck.Inspect(".")
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		output := ""
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--json":
			case "--output":
				i++
				if i >= len(args) {
					tui.Err("--output requires a file")
					os.Exit(2)
				}
				output = args[i]
			default:
				tui.Err("usage: kt deploy metadata [--json] [--output FILE]")
				os.Exit(2)
			}
		}
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		data = append(data, '\n')
		if output != "" {
			if err := os.WriteFile(output, data, 0644); err != nil {
				tui.Err(err.Error())
				os.Exit(1)
			}
			tui.OK("wrote " + output)
			return
		}
		fmt.Print(string(data))
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
		tui.Err("usage: kt deploy inspect [--json] | metadata [--json] [--output FILE] | check [--json]")
		os.Exit(2)
	}
}

func serviceNames(services []ktconfig.Service) []string {
	names := make([]string, 0, len(services))
	for _, service := range services {
		names = append(names, service.Name)
	}
	return names
}

func cmdRelease(args []string) {
	if len(args) < 1 {
		releaseUsage()
		os.Exit(2)
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		usageTopic([]string{"release"})
		return
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
		if opts.json || globalJSON {
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
	case "notes":
		notes, err := releaseNotes(args[1:])
		if err != nil {
			tui.Err(err.Error())
			os.Exit(1)
		}
		fmt.Print(notes)
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
		suggestion := suggest(args[0], []string{"next", "plan", "notes", "validate", "tag", "push"})
		if suggestion != "" {
			tui.Err("unknown release command: " + args[0] + " (did you mean " + suggestion + "?)")
		}
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
	tui.Err("usage: kt release next <patch|minor|major|pre|stable> [--pre rc] | plan <patch|minor|major|version> [--pre rc] [--json] | notes [range|--since latest] | validate <vversion> [--github-output] | tag|push <version>")
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
			if s := suggest(kind, []string{"patch", "minor", "major", "pre", "stable"}); s != "" {
				return versioning.Version{}, fmt.Errorf("%w (did you mean %s?)", err, s)
			}
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

func releaseNotes(args []string) (string, error) {
	rangeSpec := ""
	if len(args) >= 2 && args[0] == "--since" && args[1] == "latest" {
		last, err := gitOutput("describe", "--tags", "--abbrev=0", "HEAD^")
		if err == nil && strings.TrimSpace(last) != "" {
			rangeSpec = strings.TrimSpace(last) + "..HEAD"
		}
	} else if len(args) >= 1 {
		rangeSpec = args[0]
	} else {
		last, err := gitOutput("describe", "--tags", "--abbrev=0", "HEAD^")
		if err == nil && strings.TrimSpace(last) != "" {
			rangeSpec = strings.TrimSpace(last) + "..HEAD"
		}
	}
	gitArgs := []string{"log", "--oneline"}
	if rangeSpec != "" {
		gitArgs = append(gitArgs, rangeSpec)
	}
	out, err := gitOutput(gitArgs...)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return "No changes.\n", nil
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String(), nil
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
func suggest(input string, options []string) string {
	best := ""
	bestDistance := 3
	for _, option := range options {
		d := editDistance(input, option)
		if d < bestDistance {
			bestDistance = d
			best = option
		}
	}
	return best
}

func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
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
