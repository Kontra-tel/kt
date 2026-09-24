package deploycheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.kontra.tel/kontra.tel/Kt/internal/ktconfig"
	"gopkg.in/yaml.v3"
)

const (
	contentsStartMarker = "  # kt:contents:start"
	contentsEndMarker   = "  # kt:contents:end"
)

// PackageEntry is one source-to-destination mapping required by a project.
type PackageEntry struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Type        string `json:"type,omitempty"`
	Mode        int    `json:"mode,omitempty"`
}

// PackagePlan is the manifest-derived package contract.
type PackagePlan struct {
	App     string         `json:"app"`
	Entries []PackageEntry `json:"entries"`
}

// PlanProject returns the package mappings declared by a project's normalized manifest.
func PlanProject(dir string) (PackagePlan, error) {
	project, err := ktconfig.LoadFile(filepath.Join(dir, ".kt", "project.yaml"))
	if err != nil {
		return PackagePlan{}, err
	}
	return Plan(project), nil
}

// Plan derives package mappings without reading package configuration.
func Plan(project ktconfig.Project) PackagePlan {
	entries := []PackageEntry{{
		Source:      "dist/app",
		Destination: filepath.ToSlash(filepath.Join("/usr/lib", project.App)),
	}}
	for _, command := range project.CommandDetails() {
		entries = append(entries, PackageEntry{
			Source:      command.Path,
			Destination: filepath.ToSlash(filepath.Join("/usr/bin", command.Name)),
			Mode:        0o755,
		})
	}
	for _, service := range project.ServiceDetails() {
		service = fillServiceDefaults(project, service)
		entries = append(entries, PackageEntry{
			Source:      service.Runner,
			Destination: filepath.ToSlash(filepath.Join("/usr/lib", project.App, "bin", filepath.Base(service.Runner))),
			Mode:        0o755,
		})
	}
	entries = append(entries, PackageEntry{
		Source:      filepath.ToSlash(filepath.Join(project.Config.Dir, "*"+project.Config.ExampleSuffix)),
		Destination: project.Config.InstallDir + "/",
		Type:        "config|noreplace",
	})
	for _, service := range project.ServiceDetails() {
		service = fillServiceDefaults(project, service)
		entries = append(entries, PackageEntry{
			Source:      service.Unit,
			Destination: filepath.ToSlash(filepath.Join("/usr/lib/systemd/system", filepath.Base(service.Unit))),
		})
	}
	return PackagePlan{App: project.App, Entries: entries}
}

// ManagedContents returns the marker-owned nFPM entries for a package plan.
func ManagedContents(plan PackagePlan) string {
	var b strings.Builder
	b.WriteString(contentsStartMarker)
	b.WriteByte('\n')
	for _, entry := range plan.Entries {
		fmt.Fprintf(&b, "  - src: %s\n", entry.Source)
		fmt.Fprintf(&b, "    dst: %s\n", entry.Destination)
		if entry.Type != "" {
			fmt.Fprintf(&b, "    type: %s\n", entry.Type)
		}
		if entry.Mode != 0 {
			fmt.Fprintf(&b, "    file_info:\n      mode: 0%o\n", entry.Mode)
		}
		b.WriteByte('\n')
	}
	b.WriteString(contentsEndMarker)
	return b.String()
}

// PreviewPackageContents returns the updated nFPM document without writing it.
func PreviewPackageContents(file string, plan PackagePlan) (string, bool, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", false, err
	}
	text := string(data)
	start := strings.Index(text, contentsStartMarker)
	end := strings.Index(text, contentsEndMarker)
	if start < 0 || end < 0 || end < start {
		return "", false, fmt.Errorf("%s has no kt-managed contents block; add the v1.5 markers before syncing", file)
	}
	end += len(contentsEndMarker)
	updated := text[:start] + ManagedContents(plan) + text[end:]
	return updated, updated != text, nil
}

// SyncPackageContents refreshes only the marker-owned portion of nfpm.yaml.
func SyncPackageContents(file string, plan PackagePlan) (bool, error) {
	updated, changed, err := PreviewPackageContents(file, plan)
	if err != nil || !changed {
		return changed, err
	}
	if err := os.WriteFile(file, []byte(updated), 0644); err != nil {
		return false, err
	}
	return true, nil
}

type nfpmManifest struct {
	Contents []nfpmContent `yaml:"contents"`
}

type nfpmContent struct {
	Source      string `yaml:"src"`
	Destination string `yaml:"dst"`
	Type        string `yaml:"type"`
	FileInfo    struct {
		Mode int `yaml:"mode"`
	} `yaml:"file_info"`
}

func checkPackagePlan(checks *[]Check, root string, plan PackagePlan) {
	path := filepath.Join(root, "nfpm.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var manifest nfpmManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		*checks = append(*checks, Check{Level: "error", Name: "package manifest", Path: "nfpm.yaml", Message: "invalid YAML: " + err.Error()})
		return
	}
	for _, expected := range plan.Entries {
		found := false
		for _, actual := range manifest.Contents {
			if actual.Source != expected.Source || actual.Destination != expected.Destination {
				continue
			}
			if expected.Type != "" && actual.Type != expected.Type {
				continue
			}
			if expected.Mode != 0 && actual.FileInfo.Mode != expected.Mode {
				continue
			}
			found = true
			break
		}
		if !found {
			*checks = append(*checks, Check{
				Level:   "error",
				Name:    "package mapping",
				Path:    "nfpm.yaml",
				Message: fmt.Sprintf("missing %s -> %s", expected.Source, expected.Destination),
			})
		}
	}
}

func validProjectPath(path, root string) bool {
	if path == "" || filepath.IsAbs(path) {
		return false
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	return clean == root || strings.HasPrefix(clean, root+string(filepath.Separator))
}
