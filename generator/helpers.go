package generator

import (
	"strings"
)

func UncapFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func UncapFirstSingle(s string) string {
	return strings.ToLower(s[:1])
}

func MergeMaps(m, m2 map[string]string) {
	for k, v := range m2 {
		m[k] = v
	}
}

func JoinAround(strs []string, left, right, separator string) string {
	s := Scribler{}
	for k, v := range strs {
		if k > 0 {
			s.BPrint(separator)
		}
		s.BPrint(left, v, right)
	}
	return s.String()
}

func Contains[T comparable](arr []T, a T) bool {
	for _, v := range arr {
		if v == a {
			return true
		}
	}
	return false
}
