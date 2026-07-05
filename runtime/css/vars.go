package css

import "strings"

// maxVarDepth bounds recursive var() expansion so reference cycles collapse
// to empty rather than hanging.
const maxVarDepth = 8

// ExpandVarRefs substitutes var(--name) and var(--name, fallback) references
// in a property value. Unresolvable references without a fallback expand to
// "" (the property is then dropped by the graceful-drop rule). Variable
// values may themselves contain var() references.
func ExpandVarRefs(value string, vars map[string]string) string {
	return expandVars(value, vars, 0)
}

func expandVars(value string, vars map[string]string, depth int) string {
	if depth > maxVarDepth || !strings.Contains(value, "var(") {
		return valueOrEmptyOnBudget(value, depth)
	}
	var b strings.Builder
	rest := value
	for {
		i := strings.Index(rest, "var(")
		if i < 0 {
			b.WriteString(rest)
			return b.String()
		}
		b.WriteString(rest[:i])
		body, after, ok := matchParen(rest[i+len("var("):])
		if !ok {
			// Unterminated var() — drop the remainder.
			return b.String()
		}
		b.WriteString(resolveVar(body, vars, depth))
		rest = after
	}
}

// valueOrEmptyOnBudget returns "" when the depth budget is blown while a
// var() reference is still present (cycle), else the value unchanged.
func valueOrEmptyOnBudget(value string, depth int) string {
	if depth > maxVarDepth && strings.Contains(value, "var(") {
		return ""
	}
	return value
}

// matchParen scans to the closing paren at depth zero, returning the body
// and the text after it.
func matchParen(s string) (string, string, bool) {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return s[:i], s[i+1:], true
			}
			depth--
		}
	}
	return "", "", false
}

// resolveVar resolves one var() body: "--name" or "--name, fallback".
func resolveVar(body string, vars map[string]string, depth int) string {
	name := body
	fallback := ""
	hasFallback := false
	if i := topLevelComma(body); i >= 0 {
		name, fallback = body[:i], strings.TrimSpace(body[i+1:])
		hasFallback = true
	}
	name = strings.TrimSpace(name)
	if v, ok := vars[name]; ok {
		return expandVars(v, vars, depth+1)
	}
	if hasFallback {
		return expandVars(fallback, vars, depth+1)
	}
	return ""
}

func topLevelComma(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
