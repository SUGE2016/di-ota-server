package server

import (
	"strconv"
	"strings"
)

func normalizeVersionToken(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	v = strings.TrimPrefix(v, "v")
	return v
}

// CompareVersion returns -1 if a<b, 0 if equal/unordered, 1 if a>b.
// Unparseable segments compare lexicographically.
func CompareVersion(a, b string) int {
	a = normalizeVersionToken(a)
	b = normalizeVersionToken(b)
	if a == b {
		return 0
	}
	if a == "" {
		return -1
	}
	if b == "" {
		return 1
	}

	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	n := len(aParts)
	if len(bParts) > n {
		n = len(bParts)
	}
	for i := 0; i < n; i++ {
		var ap, bp string
		if i < len(aParts) {
			ap = aParts[i]
		} else {
			ap = "0"
		}
		if i < len(bParts) {
			bp = bParts[i]
		} else {
			bp = "0"
		}
		ai, aErr := strconv.Atoi(ap)
		bi, bErr := strconv.Atoi(bp)
		if aErr == nil && bErr == nil {
			if ai < bi {
				return -1
			}
			if ai > bi {
				return 1
			}
			continue
		}
		if ap < bp {
			return -1
		}
		if ap > bp {
			return 1
		}
	}
	return 0
}

func versionAtLeast(current, target string) bool {
	return CompareVersion(current, target) >= 0
}

func effectiveReportedVersion(reported, catalog string) string {
	if strings.TrimSpace(reported) != "" {
		return strings.TrimSpace(reported)
	}
	return strings.TrimSpace(catalog)
}
