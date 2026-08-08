package versioning

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+)\.([1-9][0-9]*))?$`)

type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	PreN  int
}

func Parse(s string) (Version, error) {
	var v Version
	main := strings.TrimPrefix(strings.TrimSpace(s), "v")
	matches := versionPattern.FindStringSubmatch(main)
	if matches == nil {
		return Version{}, fmt.Errorf("version must be strict semver: x.y.z or x.y.z-label.n")
	}
	var err error
	if v.Major, err = strconv.Atoi(matches[1]); err != nil {
		return Version{}, err
	}
	if v.Minor, err = strconv.Atoi(matches[2]); err != nil {
		return Version{}, err
	}
	if v.Patch, err = strconv.Atoi(matches[3]); err != nil {
		return Version{}, err
	}
	if matches[4] != "" {
		v.Pre = matches[4]
		if v.PreN, err = strconv.Atoi(matches[5]); err != nil {
			return Version{}, err
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

func (v Version) Next(kind string) (Version, error) {
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
		return Version{}, fmt.Errorf("unknown release kind %q", kind)
	}
	return v, nil
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
