package css

import "strings"

// mediaMatches evaluates a media query at style-resolution (compile) time.
// Only conditions decidable then are honoured: display-mode is terminal.
// Conditions that depend on runtime state (prefers-color-scheme,
// min/max-width/height) are not yet supported — the block is skipped so the
// stylesheet stays valid dual-target CSS.
func mediaMatches(query string) bool {
	for _, cond := range strings.Split(query, " and ") {
		if !conditionMatches(strings.TrimSpace(cond)) {
			return false
		}
	}
	return true
}

func conditionMatches(cond string) bool {
	cond = strings.TrimPrefix(cond, "(")
	cond = strings.TrimSuffix(cond, ")")
	name, value, found := strings.Cut(cond, ":")
	if !found {
		return false
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "display-mode" {
		return value == "terminal"
	}
	return false
}
