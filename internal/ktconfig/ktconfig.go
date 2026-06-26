package ktconfig

import (
	"fmt"
	"os"
	"strings"
)

var path = ".kt/project.yaml"

type Project struct {
	Template string
	App      string
	Kind     string
	Services string
	User     string
	Group    string
}

// Get reads a key from .kt/project.yaml.
func Get(key string) (string, error) {
	lines, err := readLines()
	if err != nil {
		return "", err
	}
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if k, v, ok := strings.Cut(t, ":"); ok && strings.TrimSpace(k) == key {
			return strings.TrimSpace(v), nil
		}
	}
	return "", fmt.Errorf("key %q not found in %s", key, path)
}

// Set updates or appends a key in .kt/project.yaml.
func Set(key, value string) error {
	lines, err := readLines()
	if err != nil {
		return err
	}
	prefix := key + ":"
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			lines[i] = key + ": " + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+": "+value)
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// All returns all key-value pairs from .kt/project.yaml, preserving order.
func All() ([][2]string, error) {
	return allFrom(path)
}

// Load reads and normalizes the project contract from .kt/project.yaml.
func Load() (Project, error) {
	return LoadFile(path)
}

// LoadFile reads and normalizes the project contract from the given file.
func LoadFile(file string) (Project, error) {
	pairs, err := allFrom(file)
	if err != nil {
		return Project{}, err
	}
	var p Project
	for _, pair := range pairs {
		switch pair[0] {
		case "template":
			p.Template = pair[1]
		case "app":
			p.App = pair[1]
		case "kind":
			p.Kind = pair[1]
		case "services":
			p.Services = pair[1]
		case "user":
			p.User = pair[1]
		case "group":
			p.Group = pair[1]
		}
	}
	p.normalize()
	return p, nil
}

// ServicesList returns the configured service names.
func (p Project) ServicesList() []string {
	if strings.TrimSpace(p.Services) == "" {
		return nil
	}
	parts := strings.Split(p.Services, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (p Project) HasServices() bool {
	return len(p.ServicesList()) > 0
}

func allFrom(file string) ([][2]string, error) {
	lines, err := readLinesFrom(file)
	if err != nil {
		return nil, err
	}
	var out [][2]string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if k, v, ok := strings.Cut(t, ":"); ok {
			out = append(out, [2]string{strings.TrimSpace(k), strings.TrimSpace(v)})
		}
	}
	return out, nil
}

func readLines() ([]string, error) {
	return readLinesFrom(path)
}

func readLinesFrom(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s not found — run kt init first", file)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return lines, nil
}

func (p *Project) normalize() {
	if p.Kind == "" {
		switch p.Template {
		case "app", "service":
			p.Kind = "service"
		case "multi":
			p.Kind = "multi-service"
		case "mixed":
			p.Kind = "mixed"
		case "cli":
			p.Kind = "cli"
		}
	}
	if strings.TrimSpace(p.Services) == "" {
		switch p.Kind {
		case "service":
			if p.App != "" {
				p.Services = p.App
			}
		case "multi-service":
			if p.App != "" {
				p.Services = p.App + "-backend," + p.App + "-frontend"
			}
		case "mixed":
			if p.App != "" {
				p.Services = p.App + "-service"
			}
		}
	}
}
