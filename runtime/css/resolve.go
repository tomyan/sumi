package css

import (
	"sort"
	"strings"

	"github.com/tomyan/sumi/parser/style"
)

// Element identifies one node on the path from the root to the element being
// styled: its tag name, id, class list, and attributes, plus the sibling
// context needed by structural pseudo-classes and sibling combinators.
type Element struct {
	Tag     string
	ID      string
	Classes []string
	Attrs   map[string]string

	Siblings []Element // all element siblings including self, in order (nil = unknown)
	Index    int       // position of self within Siblings
	Empty    bool      // element has no children/content

	ContainerW int // nearest laid-out ancestor width (container queries)
	ContainerH int // nearest laid-out ancestor height
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

// ResolveFocus computes the cascaded :focus properties for the element at
// the end of path. Returns nil if no focus rules match.
func ResolveFocus(stylesheet *style.Stylesheet, path []Element) map[string]string {
	props := resolveWithPseudo(stylesheet, path, "focus")
	if len(props) == 0 {
		return nil
	}
	return props
}

// ResolvePseudoElement computes the cascaded properties for a ::before or
// ::after pseudo-element of the element at the end of path. Returns nil when
// no rules match.
func ResolvePseudoElement(stylesheet *style.Stylesheet, path []Element, name string) map[string]string {
	props := resolvePseudoElement(stylesheet, path, name)
	if len(props) == 0 {
		return nil
	}
	return props
}

func resolvePseudoElement(stylesheet *style.Stylesheet, path []Element, name string) map[string]string {
	type match struct {
		spec  style.Specificity
		order int
		props map[string]string
	}
	var matches []match
	for i, rule := range stylesheet.Rules {
		if rule.PseudoElement != name || rule.Pseudo != "" {
			continue
		}
		if rule.Media != "" && !mediaMatches(rule.Media) {
			continue
		}
		if rule.Supports != "" && !supportsMatches(rule.Supports) {
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

func resolveWithPseudo(stylesheet *style.Stylesheet, path []Element, pseudo string) map[string]string {
	type match struct {
		spec  style.Specificity
		order int
		props map[string]string
	}
	var matches []match
	for i, rule := range stylesheet.Rules {
		if rule.Pseudo != pseudo || rule.PseudoElement != "" {
			continue
		}
		if rule.Media != "" && !mediaMatches(rule.Media) {
			continue
		}
		if rule.Supports != "" && !supportsMatches(rule.Supports) {
			continue
		}
		if rule.Container != "" {
			self := path[len(path)-1]
			if !containerMatches(rule.Container, self.ContainerW, self.ContainerH) {
				continue
			}
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
// element with its combinator chain satisfied.
func matchComplex(sel style.ComplexSelector, path []Element) bool {
	if len(sel.Parts) == 0 || len(path) == 0 {
		return false
	}
	return matchFrom(sel, len(sel.Parts)-1, path)
}

// matchFrom matches sel.Parts[pi] against the last element of path, then
// walks the combinator chain leftward. For sibling combinators the last path
// element is replaced by a preceding sibling (siblings share ancestors).
func matchFrom(sel style.ComplexSelector, pi int, path []Element) bool {
	// State pseudo-classes are supported on the subject compound only;
	// a state pseudo on an ancestor/sibling compound makes the rule inert.
	if pi != len(sel.Parts)-1 && sel.Parts[pi].Pseudo != "" {
		return false
	}
	self := path[len(path)-1]
	if !matchSimple(sel.Parts[pi], self) {
		return false
	}
	if pi == 0 {
		return true
	}
	switch sel.Combinators[pi-1] {
	case '>':
		if len(path) < 2 {
			return false
		}
		return matchFrom(sel, pi-1, path[:len(path)-1])
	case '+':
		if self.Siblings == nil || self.Index == 0 {
			return false
		}
		return matchFrom(sel, pi-1, siblingPath(path, self, self.Index-1))
	case '~':
		for k := self.Index - 1; k >= 0; k-- {
			if self.Siblings == nil {
				break
			}
			if matchFrom(sel, pi-1, siblingPath(path, self, k)) {
				return true
			}
		}
		return false
	default: // descendant
		for e := len(path) - 2; e >= 0; e-- {
			if matchFrom(sel, pi-1, path[:e+1]) {
				return true
			}
		}
		return false
	}
}

// siblingPath swaps the last path element for the sibling at index k,
// giving it the shared sibling context.
func siblingPath(path []Element, self Element, k int) []Element {
	sib := self.Siblings[k]
	sib.Siblings = self.Siblings
	sib.Index = k
	out := make([]Element, len(path))
	copy(out, path)
	out[len(out)-1] = sib
	return out
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
	for _, ps := range s.Structural {
		if !matchStructural(ps, e) {
			return false
		}
	}
	for _, lp := range s.Logical {
		if !matchLogical(lp, e) {
			return false
		}
	}
	return true
}

// matchLogical evaluates :not()/:is()/:where() against an element.
func matchLogical(lp style.LogicalPseudo, e Element) bool {
	anyMatch := false
	for _, arg := range lp.Args {
		if matchCompoundArg(arg, e) {
			anyMatch = true
			break
		}
	}
	if lp.Name == "not" {
		return !anyMatch
	}
	return anyMatch // is, where
}

// matchCompoundArg matches one logical-pseudo argument. Arguments with
// combinators are unsupported and never match.
func matchCompoundArg(arg style.ComplexSelector, e Element) bool {
	if len(arg.Parts) != 1 {
		return false
	}
	return matchSimple(arg.Parts[0], e)
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
