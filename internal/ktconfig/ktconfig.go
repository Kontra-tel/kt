package ktconfig

import (
	"fmt"
	"os"
	"strings"
)

const path = ".kt/project.yaml"

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
	lines, err := readLines()
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s not found — run kt init first", path)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return lines, nil
}
