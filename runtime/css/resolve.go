package css

import (
	"strings"

	"github.com/tomyan/sumi/parser/style"
)

// Resolve finds all matching rules for an element and merges their properties.
// tag is the element type ("text", "box"), classes is the list of class names.
// Rules are matched in order; later properties override earlier ones.
// Only matches rules with no pseudo-class.
func Resolve(stylesheet *style.Stylesheet, tag string, classes []string) map[string]string {
	props := make(map[string]string)
	for _, rule := range stylesheet.Rules {
		if rule.Pseudo == "" && matchesSelector(rule.Selector, tag, classes) {
			for k, v := range rule.Properties {
				props[k] = v
			}
		}
	}
	return props
}

// ResolveHover finds all matching :hover rules for an element.
// Returns nil if no hover rules match.
func ResolveHover(stylesheet *style.Stylesheet, tag string, classes []string) map[string]string {
	var props map[string]string
	for _, rule := range stylesheet.Rules {
		if rule.Pseudo == "hover" && matchesSelector(rule.Selector, tag, classes) {
			if props == nil {
				props = make(map[string]string)
			}
			for k, v := range rule.Properties {
				props[k] = v
			}
		}
	}
	return props
}

func matchesSelector(selector, tag string, classes []string) bool {
	if strings.HasPrefix(selector, ".") {
		className := selector[1:]
		for _, c := range classes {
			if c == className {
				return true
			}
		}
		return false
	}
	return selector == tag
}
