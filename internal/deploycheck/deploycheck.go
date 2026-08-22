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
	App       string   `json:"app"`
	Kind      string   `json:"kind"`
	Services  []string `json:"services"`
	ConfigDir string   `json:"config_dir"`
	DataDir   string   `json:"data_dir"`
	LogDir    string   `json:"log_dir"`
}

func Inspect(dir string) (Info, error) {
	project, err := ktconfig.LoadFile(filepath.Join(dir, ".kt", "project.yaml"))
	if err != nil {
		return Info{}, err
	}
	return Info{
		App:       project.App,
		Kind:      project.Kind,
		Services:  project.ServicesList(),
		ConfigDir: filepath.Join("/etc", project.App),
		DataDir:   filepath.Join("/var/lib", project.App),
		LogDir:    filepath.Join("/var/log", project.App),
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

	if hasConfigExample(filepath.Join(dir, "deploy", "config")) {
		add("ok", "config examples", filepath.Join(dir, "deploy", "config"), "")
	} else {
		add("warn", "config examples", filepath.Join(dir, "deploy", "config"), "no *.example files found")
	}

	if project.Kind == "cli" {
		cmd := filepath.Join(dir, "deploy", "bin", project.App)
		checkExecutable(&checks, dir, "command", cmd)
		return checks, nil
	}

	for _, service := range project.ServicesList() {
		runner := filepath.Join(dir, "deploy", "run", service)
		unit := filepath.Join(dir, "deploy", "systemd", service+".service")
		checkExecutable(&checks, dir, "service runner", runner)
		checkFile(&checks, dir, "service unit", unit)
		checkUnit(&checks, dir, project.App, service, unit)
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

func hasConfigExample(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".example") {
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
