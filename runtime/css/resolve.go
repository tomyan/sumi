package css

import (
	"sort"
	"strings"

	"github.com/tomyan/sumi/parser/style"
)

// Element identifies one node on the path from the root to the element being
// styled: its tag name, id, class list, and attributes.
type Element struct {
	Tag     string
	ID      string
	Classes []string
	Attrs   map[string]string
}

// Resolve computes the cascaded properties for the element at the end of path.
// Matching rules merge in cascade order: ascending specificity, then source
// order — so higher-specificity and later declarations win.
func Resolve(stylesheet *style.Stylesheet, path []Element) map[string]string {
	return resolveWithPseudo(stylesheet, path, "")
}

// ResolveHover computes the cascaded :hover properties for the element at the
// end of path. Returns nil if no hover rules match.
func ResolveHover(stylesheet *style.Stylesheet, path []Element) map[string]string {
	props := resolveWithPseudo(stylesheet, path, "hover")
	if len(props) == 0 {
		return nil
	}
	return props
}

func resolveWithPseudo(stylesheet *style.Stylesheet, path []Element, pseudo string) map[string]string {
	type match struct {
		spec  style.Specificity
		order int
		props map[string]string
	}
	var matches []match
	for i, rule := range stylesheet.Rules {
		if rule.Pseudo != pseudo {
			continue
		}
		if matchComplex(rule.Parsed, path) {
			matches = append(matches, match{rule.Parsed.Specificity(), i, rule.Properties})
		}
	}
	sort.SliceStable(matches, func(a, b int) bool {
		return matches[a].spec.Less(matches[b].spec)
	})
	merged := make(map[string]string)
	for _, m := range matches {
		for k, v := range m.props {
			merged[k] = v
		}
	}
	return merged
}

// matchComplex reports whether the selector's subject matches the last path
// element with its combinator chain satisfied by the ancestors.
func matchComplex(sel style.ComplexSelector, path []Element) bool {
	n := len(sel.Parts)
	if n == 0 || len(path) == 0 {
		return false
	}
	if !matchSimple(sel.Parts[n-1], path[len(path)-1]) {
		return false
	}
	return matchAncestors(sel, n-2, path, len(path)-2)
}

// matchAncestors matches sel.Parts[0..pi] against path[0..ei] right to left.
func matchAncestors(sel style.ComplexSelector, pi int, path []Element, ei int) bool {
	if pi < 0 {
		return true
	}
	if sel.Combinators[pi] == '>' {
		if ei < 0 || !matchSimple(sel.Parts[pi], path[ei]) {
			return false
		}
		return matchAncestors(sel, pi-1, path, ei-1)
	}
	// Descendant: any ancestor position may match.
	for e := ei; e >= 0; e-- {
		if matchSimple(sel.Parts[pi], path[e]) && matchAncestors(sel, pi-1, path, e-1) {
			return true
		}
	}
	return false
}

func matchSimple(s style.SimpleSelector, e Element) bool {
	if s.Tag != "" && s.Tag != e.Tag {
		return false
	}
	if s.ID != "" && s.ID != e.ID {
		return false
	}
	for _, c := range s.Classes {
		if !hasClass(e.Classes, c) {
			return false
		}
	}
	for _, a := range s.Attrs {
		if !matchAttr(a, e.Attrs) {
			return false
		}
	}
	return true
}

func matchAttr(m style.AttrMatcher, attrs map[string]string) bool {
	v, ok := attrs[m.Name]
	if !ok {
		return false
	}
	switch m.Op {
	case "":
		return true
	case "=":
		return v == m.Value
	case "^=":
		return m.Value != "" && strings.HasPrefix(v, m.Value)
	case "$=":
		return m.Value != "" && strings.HasSuffix(v, m.Value)
	case "*=":
		return m.Value != "" && strings.Contains(v, m.Value)
	case "~=":
		for _, word := range strings.Fields(v) {
			if word == m.Value {
				return true
			}
		}
		return false
	case "|=":
		return v == m.Value || strings.HasPrefix(v, m.Value+"-")
	}
	return false
}

func hasClass(classes []string, want string) bool {
	for _, c := range classes {
		if c == want {
			return true
		}
	}
	return false
}
