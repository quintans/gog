package generator

import "strings"

func UncapFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func UncapFirstSingle(s string) string {
	return strings.ToLower(s[:1])
}

func MergeMaps(m map[string]string, m2 map[string]string) {
	for k, v := range m2 {
		m[k] = v
	}
}
