package deploycheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.kontra.tel/kontra.tel/Kt/internal/ktconfig"
)

type Check struct {
	Level   string `json:"level"`
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
}

type Info struct {
	App       string             `json:"app"`
	Kind      string             `json:"kind"`
	Services  []ktconfig.Service `json:"services"`
	ConfigDir string             `json:"config_dir"`
	DataDir   string             `json:"data_dir"`
	LogDir    string             `json:"log_dir"`
	Units     []string           `json:"units"`
	Runners   []string           `json:"runners"`
}

func Inspect(dir string) (Info, error) {
	project, err := ktconfig.LoadFile(filepath.Join(dir, ".kt", "project.yaml"))
	if err != nil {
		return Info{}, err
	}
	services := project.ServiceDetails()
	units := make([]string, 0, len(services))
	runners := make([]string, 0, len(services))
	for i := range services {
		services[i] = fillServiceDefaults(project, services[i])
		units = append(units, filepath.Base(services[i].Unit))
		runners = append(runners, filepath.Join("/usr/lib", project.App, "bin", filepath.Base(services[i].Runner)))
	}
	return Info{
		App:       project.App,
		Kind:      project.Kind,
		Services:  services,
		ConfigDir: project.Config.InstallDir,
		DataDir:   filepath.Join("/var/lib", project.App),
		LogDir:    filepath.Join("/var/log", project.App),
		Units:     units,
		Runners:   runners,
	}, nil
}

func CheckProject(dir string) ([]Check, error) {
	projectPath := filepath.Join(dir, ".kt", "project.yaml")
	project, err := ktconfig.LoadFile(projectPath)
	if err != nil {
		return nil, err
	}
	var checks []Check
	add := func(level, name, path, msg string) {
		checks = append(checks, Check{Level: level, Name: name, Path: cleanDisplayPath(dir, path), Message: msg})
	}

	issues := ktconfig.Validate(project)
	if len(issues) == 0 {
		add("ok", "project manifest", projectPath, "")
	} else {
		for _, issue := range issues {
			add("error", "project manifest", projectPath, issue)
		}
	}

	if exists(filepath.Join(dir, "nfpm.yaml")) {
		add("ok", "package manifest", filepath.Join(dir, "nfpm.yaml"), "")
	} else {
		add("error", "package manifest", filepath.Join(dir, "nfpm.yaml"), "missing nfpm.yaml")
	}

	configDir := project.Config.Dir
	if configDir == "" {
		configDir = filepath.Join("deploy", "config")
	}
	if hasConfigExample(filepath.Join(dir, configDir), project.Config.ExampleSuffix) {
		add("ok", "config examples", filepath.Join(dir, configDir), "")
	} else {
		add("warn", "config examples", filepath.Join(dir, configDir), "no *"+project.Config.ExampleSuffix+" files found")
	}

	if project.Kind == "cli" {
		cmd := filepath.Join(dir, "deploy", "bin", project.App)
		checkExecutable(&checks, dir, "command", cmd)
		return checks, nil
	}

	for _, svc := range project.ServiceDetails() {
		service := fillServiceDefaults(project, svc)
		runner := filepath.Join(dir, service.Runner)
		unit := filepath.Join(dir, service.Unit)
		checkExecutable(&checks, dir, "service runner", runner)
		checkFile(&checks, dir, "service unit", unit)
		checkUnit(&checks, dir, project.App, service.Name, unit)
	}
	for _, stale := range []string{
		filepath.Join(dir, ".kt", "scripts", "postinstall-systemd.sh"),
		filepath.Join(dir, ".kt", "scripts", "preremove-systemd.sh"),
	} {
		if exists(stale) {
			add("warn", "stale lifecycle script", stale, "package hooks should delegate to /etc/<app>/hooks instead")
		}
	}
	return checks, nil
}

func HasErrors(checks []Check) bool {
	for _, check := range checks {
		if check.Level == "error" {
			return true
		}
	}
	return false
}

func fillServiceDefaults(p ktconfig.Project, s ktconfig.Service) ktconfig.Service {
	if s.Runner == "" && s.Name != "" {
		s.Runner = filepath.Join("deploy", "run", s.Name)
	}
	if s.Unit == "" && s.Name != "" {
		s.Unit = filepath.Join("deploy", "systemd", s.Name+".service")
	}
	if s.User == "" {
		s.User = p.User
	}
	if s.Group == "" {
		s.Group = p.Group
	}
	return s
}

func checkFile(checks *[]Check, root, name, path string) {
	if exists(path) {
		*checks = append(*checks, Check{Level: "ok", Name: name, Path: cleanDisplayPath(root, path)})
		return
	}
	*checks = append(*checks, Check{Level: "error", Name: name, Path: cleanDisplayPath(root, path), Message: "missing file"})
}

func checkExecutable(checks *[]Check, root, name, path string) {
	info, err := os.Stat(path)
	if err != nil {
		*checks = append(*checks, Check{Level: "error", Name: name, Path: cleanDisplayPath(root, path), Message: "missing file"})
		return
	}
	if info.Mode()&0111 == 0 {
		*checks = append(*checks, Check{Level: "error", Name: name, Path: cleanDisplayPath(root, path), Message: "not executable"})
		return
	}
	*checks = append(*checks, Check{Level: "ok", Name: name, Path: cleanDisplayPath(root, path)})
}

func checkUnit(checks *[]Check, root, app, service, unit string) {
	data, err := os.ReadFile(unit)
	if err != nil {
		return
	}
	text := string(data)
	wantExec := "ExecStart=/usr/lib/" + app + "/bin/" + service
	if strings.Contains(text, wantExec) {
		*checks = append(*checks, Check{Level: "ok", Name: "unit entrypoint", Path: cleanDisplayPath(root, unit)})
	} else {
		*checks = append(*checks, Check{Level: "error", Name: "unit entrypoint", Path: cleanDisplayPath(root, unit), Message: fmt.Sprintf("missing %s", wantExec)})
	}
	for _, hardening := range []string{"NoNewPrivileges=true", "PrivateTmp=true", "ProtectSystem=strict"} {
		if !strings.Contains(text, hardening) {
			*checks = append(*checks, Check{Level: "warn", Name: "unit hardening", Path: cleanDisplayPath(root, unit), Message: "missing " + hardening})
		}
	}
}

func hasConfigExample(dir, suffix string) bool {
	if suffix == "" {
		suffix = ".example"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func cleanDisplayPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}
