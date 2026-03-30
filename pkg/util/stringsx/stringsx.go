// Package stringsx provides string and string-slice utility functions
// (Diff, Unique, Contains, Filter, Reverse, etc.). See README.md for examples.
package stringsx

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"
)

// Diff returns elements in base that are not in exclude.
func Diff(base, exclude []string) []string {
	excludeMap := make(map[string]bool, len(exclude))
	for _, s := range exclude {
		excludeMap[s] = true
	}
	var result []string
	for _, s := range base {
		if !excludeMap[s] {
			result = append(result, s)
		}
	}
	return result
}

// Include returns elements that appear in both base and include.
func Include(base, include []string) []string {
	baseMap := make(map[string]bool, len(base))
	for _, s := range base {
		baseMap[s] = true
	}
	var result []string
	for _, s := range include {
		if baseMap[s] {
			result = append(result, s)
		}
	}
	return result
}

// Unique returns a deduplicated copy of ss. Order is not guaranteed.
func Unique(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	var result []string
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

// Contains returns true if list contains strToSearch.
func Contains(list []string, strToSearch string) bool {
	for _, item := range list {
		if item == strToSearch {
			return true
		}
	}
	return false
}

// ContainsEqualFold returns true if list contains s under Unicode case-folding.
func ContainsEqualFold(list []string, s string) bool {
	for _, item := range list {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

// Index returns the index of str in array, or -1 if not found.
func Index(array []string, str string) int {
	for i, s := range array {
		if s == str {
			return i
		}
	}
	return -1
}

// Filter returns a new slice with all occurrences of strToFilter removed.
func Filter(list []string, strToFilter string) []string {
	var result []string
	for _, item := range list {
		if item != strToFilter {
			result = append(result, item)
		}
	}
	return result
}

// Add appends str to list if it is not already present.
func Add(list []string, str string) []string {
	if Contains(list, str) {
		return list
	}
	return append(list, str)
}

// Reverse returns a new string with its UTF-8 runes in reverse order.
func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

// DecodeBase64 decodes a standard base64-encoded string.
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
