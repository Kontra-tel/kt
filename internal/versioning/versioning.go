package versioning

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	PreN  int
}

func Parse(s string) (Version, error) {
	var v Version
	main := s
	if pre, ok := strings.CutPrefix(s, "v"); ok {
		main = pre
	}
	parts := strings.SplitN(main, "-", 2)
	core := strings.Split(parts[0], ".")
	if len(core) != 3 {
		return Version{}, fmt.Errorf("version must be x.y.z or x.y.z-label.n")
	}
	var err error
	if v.Major, err = strconv.Atoi(core[0]); err != nil {
		return Version{}, err
	}
	if v.Minor, err = strconv.Atoi(core[1]); err != nil {
		return Version{}, err
	}
	if v.Patch, err = strconv.Atoi(core[2]); err != nil {
		return Version{}, err
	}
	if len(parts) == 2 {
		pre := strings.Split(parts[1], ".")
		if len(pre) != 2 || strings.TrimSpace(pre[0]) == "" {
			return Version{}, fmt.Errorf("prerelease must be label.number")
		}
		v.Pre = pre[0]
		if v.PreN, err = strconv.Atoi(pre[1]); err != nil {
			return Version{}, err
		}
		if v.PreN < 1 {
			return Version{}, fmt.Errorf("prerelease number must be >= 1")
		}
	}
	return v, nil
}

func (v Version) String() string {
	if v.Pre != "" {
		return fmt.Sprintf("%d.%d.%d-%s.%d", v.Major, v.Minor, v.Patch, v.Pre, v.PreN)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) Compare(other Version) int {
	switch {
	case v.Major != other.Major:
		return cmp(v.Major, other.Major)
	case v.Minor != other.Minor:
		return cmp(v.Minor, other.Minor)
	case v.Patch != other.Patch:
		return cmp(v.Patch, other.Patch)
	}
	if v.Pre == "" && other.Pre == "" {
		return 0
	}
	if v.Pre == "" {
		return 1
	}
	if other.Pre == "" {
		return -1
	}
	if rank := cmp(preRank(v.Pre), preRank(other.Pre)); rank != 0 {
		return rank
	}
	if v.Pre != other.Pre {
		if v.Pre < other.Pre {
			return -1
		}
		return 1
	}
	return cmp(v.PreN, other.PreN)
}

func Bump(file, kind string) (string, error) {
	v, err := readVersion(file)
	if err != nil {
		return "", err
	}
	// Stable bumps always discard prerelease suffixes.
	v.Pre = ""
	v.PreN = 0
	switch kind {
	case "patch":
		v.Patch++
	case "minor":
		v.Minor++
		v.Patch = 0
	case "major":
		v.Major++
		v.Minor = 0
		v.Patch = 0
	default:
		return "", fmt.Errorf("unknown bump %q", kind)
	}
	return v.String(), writeVersion(file, v)
}

func Set(file, version string) (string, error) {
	v, err := Parse(strings.TrimSpace(version))
	if err != nil {
		return "", err
	}
	return v.String(), writeVersion(file, v)
}

func readVersion(file string) (Version, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return Version{}, err
	}
	return Parse(strings.TrimSpace(string(b)))
}

func writeVersion(file string, v Version) error {
	return os.WriteFile(file, []byte(v.String()+"\n"), 0644)
}

func cmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func preRank(label string) int {
	switch label {
	case "alpha":
		return 0
	case "beta":
		return 1
	case "rc":
		return 2
	default:
		return 3
	}
}
