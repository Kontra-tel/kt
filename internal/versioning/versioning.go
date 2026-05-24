package versioning

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Bump(file, kind string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimSpace(string(b)), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("version must be x.y.z")
	}
	n := make([]int, 3)
	for i, p := range parts {
		n[i], err = strconv.Atoi(p)
		if err != nil {
			return "", err
		}
	}
	switch kind {
	case "patch":
		n[2]++
	case "minor":
		n[1]++
		n[2] = 0
	case "major":
		n[0]++
		n[1] = 0
		n[2] = 0
	default:
		return "", fmt.Errorf("unknown bump %q", kind)
	}
	v := fmt.Sprintf("%d.%d.%d", n[0], n[1], n[2])
	return v, os.WriteFile(file, []byte(v+"\n"), 0644)
}
