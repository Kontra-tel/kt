// Package doctor reports whether a generated project's required build tools are available.
package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Tool describes one required build tool.
type Tool struct {
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
	Found   bool   `json:"found"`
}

// Result is the machine-readable doctor contract.
type Result struct {
	OK    bool   `json:"ok"`
	Tools []Tool `json:"tools"`
}

// Check reads DOCTOR_TOOLS from dir/Makefile and checks each command on PATH.
func Check(dir string) (Result, error) {
	data, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		return Result{}, fmt.Errorf("read Makefile: %w", err)
	}
	tools := doctorTools(string(data))
	result := Result{OK: true, Tools: make([]Tool, 0, len(tools))}
	for _, name := range tools {
		tool := Tool{Name: name}
		path, err := exec.LookPath(name)
		if err == nil {
			tool.Found = true
			tool.Path = path
			tool.Version = commandVersion(name)
		} else {
			result.OK = false
		}
		result.Tools = append(result.Tools, tool)
	}
	return result, nil
}

func doctorTools(makefile string) []string {
	for _, line := range strings.Split(makefile, "\n") {
		line = strings.TrimSpace(line)
		for _, operator := range []string{":=", "?=", "+=", "="} {
			prefix := "DOCTOR_TOOLS " + operator
			if strings.HasPrefix(line, prefix) {
				return strings.Fields(strings.TrimSpace(strings.TrimPrefix(line, prefix)))
			}
		}
	}
	return nil
}

func commandVersion(name string) string {
	out, err := exec.Command(name, "--version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
}
