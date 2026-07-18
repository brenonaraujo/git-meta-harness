package ui

import "strings"

// Repeat is a thin wrapper around strings.Repeat to avoid importing
// strings in callers.
func Repeat(s string, n int) string {
	return strings.Repeat(s, n)
}
