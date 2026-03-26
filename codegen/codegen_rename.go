package codegen

import "strings"

// replaceIdentifier replaces all standalone occurrences of oldName with newName.
// A standalone identifier is one not preceded or followed by a Go identifier character
// (letter, digit, or underscore). This ensures "count" is not replaced inside "discount".
func replaceIdentifier(s, oldName, newName string) string {
	if oldName == newName {
		return s
	}
	var result strings.Builder
	i := 0
	for i < len(s) {
		idx := strings.Index(s[i:], oldName)
		if idx == -1 {
			result.WriteString(s[i:])
			break
		}
		absIdx := i + idx
		if isWordBoundary(s, absIdx, len(oldName)) {
			result.WriteString(s[i:absIdx])
			result.WriteString(newName)
			i = absIdx + len(oldName)
		} else {
			result.WriteString(s[i : absIdx+len(oldName)])
			i = absIdx + len(oldName)
		}
	}
	return result.String()
}

// isWordBoundary checks that the substring at pos with the given length is a standalone identifier.
func isWordBoundary(s string, pos, length int) bool {
	if pos > 0 && isIdentChar(s[pos-1]) {
		return false
	}
	after := pos + length
	if after < len(s) && isIdentChar(s[after]) {
		return false
	}
	return true
}

// isIdentChar returns true for Go identifier characters (letter, digit, underscore).
func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

