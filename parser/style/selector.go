package style

import (
	"fmt"
	"strings"
)

// SimpleSelector is one compound selector: an optional tag plus any number of
// class, id, and pseudo-class qualifiers (e.g. `box.panel#main:hover`).
// The universal selector `*` parses to an empty SimpleSelector.
type SimpleSelector struct {
	Tag     string   // "" matches any tag
	ID      string   // "" matches any id
	Classes []string // all must be present
	Pseudo  string   // pseudo-class name ("hover", ...); "" for none
}

// ComplexSelector is a combinator chain; the subject (rightmost compound)
// is Parts[len(Parts)-1]. Combinators[i] joins Parts[i] and Parts[i+1]:
// ' ' for descendant, '>' for child.
type ComplexSelector struct {
	Parts       []SimpleSelector
	Combinators []byte
}

// Specificity is the CSS (id, class, type) triple.
type Specificity struct {
	IDs, Classes, Types int
}

// Less reports whether s is lower specificity than o.
func (s Specificity) Less(o Specificity) bool {
	if s.IDs != o.IDs {
		return s.IDs < o.IDs
	}
	if s.Classes != o.Classes {
		return s.Classes < o.Classes
	}
	return s.Types < o.Types
}

// Specificity computes the selector's (id, class, type) counts.
// Pseudo-classes count as classes, per spec.
func (c ComplexSelector) Specificity() Specificity {
	var sp Specificity
	for _, p := range c.Parts {
		if p.ID != "" {
			sp.IDs++
		}
		sp.Classes += len(p.Classes)
		if p.Pseudo != "" {
			sp.Classes++
		}
		if p.Tag != "" {
			sp.Types++
		}
	}
	return sp
}

// ParseSelectorList parses a comma-separated selector list.
func ParseSelectorList(text string) ([]ComplexSelector, error) {
	var out []ComplexSelector
	for _, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("empty selector in list %q", text)
		}
		sel, err := parseComplexSelector(item)
		if err != nil {
			return nil, err
		}
		out = append(out, sel)
	}
	return out, nil
}

// parseComplexSelector parses one combinator chain like `.panel > text`.
func parseComplexSelector(text string) (ComplexSelector, error) {
	var sel ComplexSelector
	pending := byte(' ')
	for _, tok := range tokenizeSelector(text) {
		if tok == ">" {
			pending = '>'
			continue
		}
		part, err := parseSimpleSelector(tok)
		if err != nil {
			return ComplexSelector{}, err
		}
		if len(sel.Parts) > 0 {
			sel.Combinators = append(sel.Combinators, pending)
		}
		sel.Parts = append(sel.Parts, part)
		pending = ' '
	}
	if len(sel.Parts) == 0 {
		return ComplexSelector{}, fmt.Errorf("empty selector %q", text)
	}
	return sel, nil
}

// tokenizeSelector splits on whitespace, surfacing `>` as its own token even
// when written without surrounding spaces (`.a>.b`).
func tokenizeSelector(text string) []string {
	text = strings.ReplaceAll(text, ">", " > ")
	return strings.Fields(text)
}

// parseSimpleSelector parses one compound like `box.panel#main:hover` or `*`.
func parseSimpleSelector(tok string) (SimpleSelector, error) {
	var s SimpleSelector
	if tok == "*" {
		return s, nil
	}
	rest := tok
	// Leading tag name (up to the first qualifier).
	if i := strings.IndexAny(rest, ".#:"); i != 0 {
		if i < 0 {
			s.Tag = rest
			return s, nil
		}
		s.Tag = rest[:i]
		rest = rest[i:]
	}
	for rest != "" {
		kind := rest[0]
		rest = rest[1:]
		end := strings.IndexAny(rest, ".#:")
		if end < 0 {
			end = len(rest)
		}
		name := rest[:end]
		rest = rest[end:]
		if name == "" {
			return SimpleSelector{}, fmt.Errorf("empty qualifier in selector %q", tok)
		}
		switch kind {
		case '.':
			s.Classes = append(s.Classes, name)
		case '#':
			s.ID = name
		case ':':
			s.Pseudo = name
		}
	}
	return s, nil
}
