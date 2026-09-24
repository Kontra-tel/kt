package ktconfig

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var path = ".kt/project.yaml"

var (
	appNamePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	tagPrefixPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]*$`)
)

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

// Get reads a scalar key from .kt/project.yaml. Nested fields use dot notation.
func Get(key string) (string, error) {
	project, loadErr := Load()
	if loadErr == nil {
		switch key {
		case "schema":
			return project.Schema, nil
		case "template":
			return project.Template, nil
		case "app":
			return project.App, nil
		case "kind":
			return project.Kind, nil
		case "services":
			return strings.Join(project.ServicesList(), ","), nil
		case "user":
			return project.User, nil
		case "group":
			return project.Group, nil
		}
	}
	root, err := readDocument(path)
	if err != nil {
		return "", err
	}
	value := lookup(root, strings.Split(key, "."))
	if value == nil || value.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("key %q not found in %s", key, path)
	}
	return value.Value, nil
}

// Set updates or appends a top-level scalar key in .kt/project.yaml.
func Set(key, value string) error {
	if strings.Contains(key, ".") {
		return fmt.Errorf("set only supports top-level scalar keys")
	}
	root, err := readDocument(path)
	if err != nil {
		return err
	}
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("%s must contain a mapping", path)
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != key {
			continue
		}
		if root.Content[i+1].Kind != yaml.ScalarNode {
			return fmt.Errorf("key %q is not a scalar", key)
		}
		root.Content[i+1].Tag = "!!str"
		root.Content[i+1].Value = value
		return writeDocument(path, root)
	}
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
	)
	return writeDocument(path, root)
}

// All returns all top-level scalar key-value pairs, preserving document order.
func All() ([][2]string, error) {
	root, err := readDocument(path)
	if err != nil {
		return nil, err
	}
	return scalarPairs(root), nil
}

// Load reads and normalizes the project contract from .kt/project.yaml.
func Load() (Project, error) {
	return LoadFile(path)
}

// LoadFile reads and normalizes the project contract from the given file.
func LoadFile(file string) (Project, error) {
	root, err := readDocument(file)
	if err != nil {
		return Project{}, err
	}
	project := parseProject(root)
	project.normalize()
	return project, nil
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

// CommandDetails returns declared commands, including legacy template defaults.
func (p Project) CommandDetails() []Command {
	if len(p.Commands) > 0 {
		return p.Commands
	}
	if p.App == "" {
		return nil
	}
	commands := []Command{{Name: p.App, Path: "deploy/bin/" + p.App}}
	if p.Kind == "mixed" {
		commands = append(commands, Command{Name: p.App + "-service", Path: "deploy/bin/" + p.App + "-service"})
	}
	return commands
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
	if p.Release.TagPrefix != "" && !tagPrefixPattern.MatchString(p.Release.TagPrefix) {
		issues = append(issues, "release.tag_prefix must match [A-Za-z][A-Za-z0-9._-]*")
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
	for _, command := range p.CommandDetails() {
		if !SafeName(command.Name) {
			issues = append(issues, "command "+command.Name+" must match [a-z0-9][a-z0-9-]*")
		}
		if command.Path == "" {
			issues = append(issues, "command "+command.Name+" must set path")
		}
	}
	return issues
}

func parseProject(root *yaml.Node) Project {
	project := Project{
		Schema:   scalar(root, "schema"),
		Template: scalar(root, "template"),
		App:      scalar(root, "app"),
		Kind:     scalar(root, "kind"),
		Services: scalar(root, "services"),
		User:     scalar(root, "user"),
		Group:    scalar(root, "group"),
	}
	project.ServiceEntries = parseServices(lookup(root, []string{"services"}))
	project.Commands = parseCommands(lookup(root, []string{"commands"}))
	project.Package = PackageInfo{
		Name:        scalar(root, "package", "name"),
		Maintainer:  scalar(root, "package", "maintainer"),
		Description: scalar(root, "package", "description"),
		Section:     scalar(root, "package", "section"),
		License:     scalar(root, "package", "license"),
	}
	project.Config = ConfigInfo{
		Dir:           scalar(root, "config", "dir"),
		InstallDir:    scalar(root, "config", "install_dir"),
		ExampleSuffix: scalar(root, "config", "example_suffix"),
	}
	project.Release = ReleaseInfo{TagPrefix: scalar(root, "release", "tag_prefix")}
	project.KT = KTInfo{ScaffoldVersion: scalar(root, "kt", "scaffold_version")}
	return project
}

func parseServices(node *yaml.Node) []Service {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	services := make([]Service, 0, len(node.Content))
	for _, entry := range node.Content {
		if entry.Kind != yaml.MappingNode {
			continue
		}
		services = append(services, Service{
			Name:   scalar(entry, "name"),
			Role:   scalar(entry, "role"),
			Runner: scalar(entry, "runner"),
			Unit:   scalar(entry, "unit"),
			User:   scalar(entry, "user"),
			Group:  scalar(entry, "group"),
		})
	}
	return services
}

func parseCommands(node *yaml.Node) []Command {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	commands := make([]Command, 0, len(node.Content))
	for _, entry := range node.Content {
		if entry.Kind != yaml.MappingNode {
			continue
		}
		commands = append(commands, Command{Name: scalar(entry, "name"), Path: scalar(entry, "path")})
	}
	return commands
}

func scalar(node *yaml.Node, keys ...string) string {
	value := lookup(node, keys)
	if value == nil || value.Kind != yaml.ScalarNode {
		return ""
	}
	return value.Value
}

func lookup(node *yaml.Node, keys []string) *yaml.Node {
	if node == nil {
		return nil
	}
	for _, key := range keys {
		if node.Kind != yaml.MappingNode {
			return nil
		}
		var next *yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				next = node.Content[i+1]
				break
			}
		}
		if next == nil {
			return nil
		}
		node = next
	}
	return node
}

func scalarPairs(root *yaml.Node) [][2]string {
	if root == nil || root.Kind != yaml.MappingNode {
		return nil
	}
	pairs := make([][2]string, 0, len(root.Content)/2)
	for i := 0; i+1 < len(root.Content); i += 2 {
		value := root.Content[i+1]
		if value.Kind == yaml.ScalarNode {
			pairs = append(pairs, [2]string{root.Content[i].Value, value.Value})
		}
	}
	return pairs
}

func readDocument(file string) (*yaml.Node, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s not found — run kt init first", file)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s must contain a mapping", file)
	}
	return document.Content[0], nil
}

func writeDocument(file string, root *yaml.Node) error {
	document := yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	data, err := yaml.Marshal(&document)
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
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
		p.KT.ScaffoldVersion = "1.5"
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
