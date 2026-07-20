package text

import "strings"

// NormalizeUTF8 replaces each run of invalid UTF-8 byte sequences with one
// U+FFFD
func NormalizeUTF8(input string) string {
	return strings.ToValidUTF8(input, "\uFFFD")
}
