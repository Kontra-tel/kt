package ktconfig

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var path = ".kt/project.yaml"

var appNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type Service struct {
	Name   string `json:"name"`
	Role   string `json:"role,omitempty"`
	Runner string `json:"runner,omitempty"`
	Unit   string `json:"unit,omitempty"`
	User   string `json:"user,omitempty"`
	Group  string `json:"group,omitempty"`
}

type Command struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type PackageInfo struct {
	Name        string `json:"name,omitempty"`
	Maintainer  string `json:"maintainer,omitempty"`
	Description string `json:"description,omitempty"`
	Section     string `json:"section,omitempty"`
	License     string `json:"license,omitempty"`
}

type ConfigInfo struct {
	Dir           string `json:"dir,omitempty"`
	InstallDir    string `json:"install_dir,omitempty"`
	ExampleSuffix string `json:"example_suffix,omitempty"`
}

type ReleaseInfo struct {
	TagPrefix string `json:"tag_prefix,omitempty"`
}

type KTInfo struct {
	ScaffoldVersion string `json:"scaffold_version,omitempty"`
}

type Project struct {
	Schema   string
	Template string
	App      string
	Kind     string
	Services string
	User     string
	Group    string

	ServiceEntries []Service
	Commands       []Command
	Package        PackageInfo
	Config         ConfigInfo
	Release        ReleaseInfo
	KT             KTInfo
}

// Get reads a key from .kt/project.yaml.
func Get(key string) (string, error) {
	project, loadErr := Load()
	switch key {
	case "schema":
		if loadErr == nil && project.Schema != "" {
			return project.Schema, nil
		}
	case "template":
		if loadErr == nil && project.Template != "" {
			return project.Template, nil
		}
	case "app":
		if loadErr == nil && project.App != "" {
			return project.App, nil
		}
	case "kind":
		if loadErr == nil && project.Kind != "" {
			return project.Kind, nil
		}
	case "services":
		if loadErr == nil {
			return strings.Join(project.ServicesList(), ","), nil
		}
	case "user":
		if loadErr == nil && project.User != "" {
			return project.User, nil
		}
	case "group":
		if loadErr == nil && project.Group != "" {
			return project.Group, nil
		}
	}

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

// Set updates or appends a top-level scalar key in .kt/project.yaml.
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

// All returns all scalar key-value pairs from .kt/project.yaml, preserving order.
func All() ([][2]string, error) {
	return allFrom(path)
}

// Load reads and normalizes the project contract from .kt/project.yaml.
func Load() (Project, error) {
	return LoadFile(path)
}

// LoadFile reads and normalizes the project contract from the given file.
func LoadFile(file string) (Project, error) {
	lines, err := readLinesFrom(file)
	if err != nil {
		return Project{}, err
	}
	p := parseProject(lines)
	p.normalize()
	return p, nil
}

// ServicesList returns the configured service names.
func (p Project) ServicesList() []string {
	if len(p.ServiceEntries) > 0 {
		out := make([]string, 0, len(p.ServiceEntries))
		for _, service := range p.ServiceEntries {
			if strings.TrimSpace(service.Name) != "" {
				out = append(out, strings.TrimSpace(service.Name))
			}
		}
		return out
	}
	return splitList(p.Services)
}

func (p Project) ServiceDetails() []Service {
	if len(p.ServiceEntries) > 0 {
		return p.ServiceEntries
	}
	services := p.ServicesList()
	out := make([]Service, 0, len(services))
	for _, name := range services {
		out = append(out, Service{Name: name})
	}
	return out
}

func (p Project) HasServices() bool {
	return len(p.ServicesList()) > 0
}

func SafeName(name string) bool {
	return appNamePattern.MatchString(name)
}

// Validate checks the normalized project contract for values kt can safely use
// in package names, paths, service units, and generated scripts.
func Validate(p Project) []string {
	var issues []string
	if p.Schema != "" && p.Schema != "kt.project/v1" {
		issues = append(issues, "schema must be kt.project/v1")
	}
	if p.App == "" {
		issues = append(issues, "app is required")
	} else if !appNamePattern.MatchString(p.App) {
		issues = append(issues, "app must match [a-z0-9][a-z0-9-]*")
	}
	switch p.Kind {
	case "cli":
		if p.HasServices() {
			issues = append(issues, "cli projects must not declare services")
		}
	case "service", "mixed", "multi-service":
		if !p.HasServices() {
			issues = append(issues, p.Kind+" projects must declare at least one service")
		}
		if p.User == "" {
			issues = append(issues, "service-bearing projects should set user")
		}
		if p.Group == "" {
			issues = append(issues, "service-bearing projects should set group")
		}
	default:
		issues = append(issues, "kind must be cli, service, mixed, or multi-service")
	}
	structuredServices := len(p.ServiceEntries) > 0
	for _, service := range p.ServiceDetails() {
		if service.Name == "" {
			issues = append(issues, "service name is required")
			continue
		}
		if !appNamePattern.MatchString(service.Name) {
			issues = append(issues, "service "+service.Name+" must match [a-z0-9][a-z0-9-]*")
		}
		if p.Kind != "cli" && structuredServices {
			if service.Runner == "" {
				issues = append(issues, "service "+service.Name+" should set runner")
			}
			if service.Unit == "" {
				issues = append(issues, "service "+service.Name+" should set unit")
			}
		}
	}
	return issues
}

func parseProject(lines []string) Project {
	pairs := parseScalarPairs(lines)
	var p Project
	for _, pair := range pairs {
		switch pair[0] {
		case "schema":
			p.Schema = pair[1]
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
	p.ServiceEntries = parseServiceEntries(lines)
	p.Commands = parseCommandEntries(lines)
	p.Package = parsePackageInfo(lines)
	p.Config = parseConfigInfo(lines)
	p.Release = ReleaseInfo{TagPrefix: parseBlockValue(lines, "release", "tag_prefix")}
	p.KT = KTInfo{ScaffoldVersion: trimQuotes(parseBlockValue(lines, "kt", "scaffold_version"))}
	return p
}

func parseScalarPairs(lines []string) [][2]string {
	var out [][2]string
	for _, line := range lines {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if k, v, ok := strings.Cut(t, ":"); ok {
			out = append(out, [2]string{strings.TrimSpace(k), strings.TrimSpace(v)})
		}
	}
	return out
}

func parseServiceEntries(lines []string) []Service {
	block := blockLines(lines, "services")
	var services []Service
	var current *Service
	for _, line := range block {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "- ") {
			services = append(services, Service{})
			current = &services[len(services)-1]
			t = strings.TrimSpace(strings.TrimPrefix(t, "- "))
			if k, v, ok := strings.Cut(t, ":"); ok {
				setServiceField(current, strings.TrimSpace(k), strings.TrimSpace(v))
			}
			continue
		}
		if current == nil {
			continue
		}
		if k, v, ok := strings.Cut(t, ":"); ok {
			setServiceField(current, strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}
	return services
}

func parseCommandEntries(lines []string) []Command {
	block := blockLines(lines, "commands")
	var commands []Command
	var current *Command
	for _, line := range block {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "- ") {
			commands = append(commands, Command{})
			current = &commands[len(commands)-1]
			t = strings.TrimSpace(strings.TrimPrefix(t, "- "))
			if k, v, ok := strings.Cut(t, ":"); ok {
				setCommandField(current, strings.TrimSpace(k), strings.TrimSpace(v))
			}
			continue
		}
		if current == nil {
			continue
		}
		if k, v, ok := strings.Cut(t, ":"); ok {
			setCommandField(current, strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}
	return commands
}

func parsePackageInfo(lines []string) PackageInfo {
	return PackageInfo{
		Name:        parseBlockValue(lines, "package", "name"),
		Maintainer:  parseBlockValue(lines, "package", "maintainer"),
		Description: parseBlockValue(lines, "package", "description"),
		Section:     parseBlockValue(lines, "package", "section"),
		License:     parseBlockValue(lines, "package", "license"),
	}
}

func parseConfigInfo(lines []string) ConfigInfo {
	return ConfigInfo{
		Dir:           parseBlockValue(lines, "config", "dir"),
		InstallDir:    parseBlockValue(lines, "config", "install_dir"),
		ExampleSuffix: parseBlockValue(lines, "config", "example_suffix"),
	}
}

func blockLines(lines []string, name string) []string {
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == name+":" {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return nil
	}
	var out []string
	for _, line := range lines[start:] {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		out = append(out, line)
	}
	return out
}

func parseBlockValue(lines []string, block, key string) string {
	for _, line := range blockLines(lines, block) {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "- ") {
			continue
		}
		if k, v, ok := strings.Cut(t, ":"); ok && strings.TrimSpace(k) == key {
			return trimQuotes(strings.TrimSpace(v))
		}
	}
	return ""
}

func setServiceField(s *Service, key, value string) {
	value = trimQuotes(value)
	switch key {
	case "name":
		s.Name = value
	case "role":
		s.Role = value
	case "runner":
		s.Runner = value
	case "unit":
		s.Unit = value
	case "user":
		s.User = value
	case "group":
		s.Group = value
	}
}

func setCommandField(c *Command, key, value string) {
	value = trimQuotes(value)
	switch key {
	case "name":
		c.Name = value
	case "path":
		c.Path = value
	}
}

func allFrom(file string) ([][2]string, error) {
	lines, err := readLinesFrom(file)
	if err != nil {
		return nil, err
	}
	return parseScalarPairs(lines), nil
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
	if p.Schema == "" {
		p.Schema = "kt.project/v1"
	}
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
	if strings.TrimSpace(p.Services) == "" && len(p.ServiceEntries) == 0 {
		switch p.Kind {
		case "service":
			if p.App != "" {
				p.ServiceEntries = []Service{{Name: p.App, Runner: "deploy/run/" + p.App, Unit: "deploy/systemd/" + p.App + ".service", User: p.App, Group: p.App}}
			}
		case "multi-service":
			if p.App != "" {
				p.ServiceEntries = []Service{
					{Name: p.App + "-backend", Role: "backend", Runner: "deploy/run/" + p.App + "-backend", Unit: "deploy/systemd/" + p.App + "-backend.service", User: p.App, Group: p.App},
					{Name: p.App + "-frontend", Role: "frontend", Runner: "deploy/run/" + p.App + "-frontend", Unit: "deploy/systemd/" + p.App + "-frontend.service", User: p.App, Group: p.App},
				}
			}
		case "mixed":
			if p.App != "" {
				p.ServiceEntries = []Service{{Name: p.App + "-service", Role: "service", Runner: "deploy/run/" + p.App + "-service", Unit: "deploy/systemd/" + p.App + "-service.service", User: p.App, Group: p.App}}
			}
		}
	}
	if p.Services == "" && len(p.ServiceEntries) > 0 {
		p.Services = strings.Join(p.ServicesList(), ",")
	}
	if p.User == "" && len(p.ServiceEntries) > 0 {
		p.User = p.ServiceEntries[0].User
	}
	if p.Group == "" && len(p.ServiceEntries) > 0 {
		p.Group = p.ServiceEntries[0].Group
	}
	if p.Config.Dir == "" {
		p.Config.Dir = "deploy/config"
	}
	if p.Config.InstallDir == "" && p.App != "" {
		p.Config.InstallDir = "/etc/" + p.App
	}
	if p.Config.ExampleSuffix == "" {
		p.Config.ExampleSuffix = ".example"
	}
	if p.Release.TagPrefix == "" {
		p.Release.TagPrefix = "v"
	}
	if p.KT.ScaffoldVersion == "" {
		p.KT.ScaffoldVersion = "1.4"
	}
	if p.Package.Name == "" {
		p.Package.Name = p.App
	}
}

func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && part != "[]" {
			out = append(out, part)
		}
	}
	return out
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}
