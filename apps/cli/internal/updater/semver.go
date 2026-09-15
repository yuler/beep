package updater

import (
	"strconv"
	"strings"
)

// CompareVersions compares two semantic version strings.
// Returns:
//
//	 1 if v1 > v2
//	-1 if v1 < v2
//	 0 if v1 == v2
//
// Special handling:
// - Leading 'v' is stripped (e.g. "v1.2.3" -> "1.2.3").
// - "dev" is treated as older than any released version (except another "dev").
func CompareVersions(v1, v2 string) int {
	s1 := strings.TrimSpace(v1)
	s2 := strings.TrimSpace(v2)

	if s1 == s2 {
		return 0
	}
	if s1 == "dev" {
		return -1
	}
	if s2 == "dev" {
		return 1
	}

	s1 = strings.TrimPrefix(s1, "v")
	s2 = strings.TrimPrefix(s2, "v")

	if s1 == s2 {
		return 0
	}

	parts1, pre1 := splitVersionAndPrerelease(s1)
	parts2, pre2 := splitVersionAndPrerelease(s2)

	// Compare numeric components: major, minor, patch...
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		n1 := 0
		n2 := 0
		if i < len(parts1) {
			n1 = parts1[i]
		}
		if i < len(parts2) {
			n2 = parts2[i]
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	// When numeric components are equal:
	// A version with a prerelease (e.g. -rc1) is lower than normal release without prerelease.
	// e.g. 1.0.0-rc1 < 1.0.0
	if pre1 != "" && pre2 == "" {
		return -1
	}
	if pre1 == "" && pre2 != "" {
		return 1
	}
	if pre1 < pre2 {
		return -1
	}
	if pre1 > pre2 {
		return 1
	}

	return 0
}

func splitVersionAndPrerelease(v string) ([]int, string) {
	var prerelease string
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		prerelease = v[idx+1:]
		v = v[:idx]
	}

	chunks := strings.Split(v, ".")
	nums := make([]int, 0, len(chunks))
	for _, chunk := range chunks {
		n, err := strconv.Atoi(chunk)
		if err != nil {
			break
		}
		nums = append(nums, n)
	}
	return nums, prerelease
}
