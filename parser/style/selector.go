package style

import (
	"fmt"
	"strings"
)

// SimpleSelector is one compound selector: an optional tag plus any number of
// class, id, attribute, and pseudo-class qualifiers
// (e.g. `box.panel#main[focusable=true]:hover`).
// The universal selector `*` parses to an empty SimpleSelector.
type SimpleSelector struct {
	Tag        string          // "" matches any tag
	ID         string          // "" matches any id
	Classes    []string        // all must be present
	Attrs      []AttrMatcher   // all must match
	Pseudo     string          // state pseudo-class ("hover", "focus", ...); "" for none
	Structural []string        // structural pseudo-classes, raw ("first-child", "nth-child(2n+1)")
	Logical    []LogicalPseudo // :not(), :is(), :where()
}

// LogicalPseudo is a :not()/:is()/:where() pseudo-class with its selector
// arguments. Arguments are compound selectors; combinators inside the
// arguments are not supported (such an argument never matches).
type LogicalPseudo struct {
	Name string // "not", "is", "where"
	Args []ComplexSelector
}

// statePseudos are pseudo-classes driven by runtime state; they bucket the
// rule (Rule.Pseudo) rather than participating in structural matching.
var statePseudos = map[string]bool{
	"hover": true, "focus": true, "active": true,
	"checked": true, "disabled": true, "enabled": true,
}

// AttrMatcher is one attribute selector: [name], [name=v], [name^=v], etc.
type AttrMatcher struct {
	Name  string
	Op    string // "", "=", "^=", "$=", "*=", "~=", "|="
	Value string
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
		sp.Classes += len(p.Classes) + len(p.Attrs) + len(p.Structural)
		if p.Pseudo != "" {
			sp.Classes++
		}
		for _, lp := range p.Logical {
			sp = sp.add(lp.specificity())
		}
		if p.Tag != "" {
			sp.Types++
		}
	}
	return sp
}

func (s Specificity) add(o Specificity) Specificity {
	return Specificity{s.IDs + o.IDs, s.Classes + o.Classes, s.Types + o.Types}
}

// specificity of a logical pseudo: :where() contributes nothing; :not() and
// :is() contribute their most specific argument, per spec.
func (l LogicalPseudo) specificity() Specificity {
	if l.Name == "where" {
		return Specificity{}
	}
	var max Specificity
	for _, arg := range l.Args {
		if sp := arg.Specificity(); max.Less(sp) {
			max = sp
		}
	}
	return max
}

// ParseSelectorList parses a comma-separated selector list. Commas inside
// parentheses or brackets (`:is(.a, .b)`) do not split.
func ParseSelectorList(text string) ([]ComplexSelector, error) {
	var out []ComplexSelector
	for _, item := range SplitSelectorList(text) {
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

// SplitSelectorList splits selector-list text on top-level commas only.
func SplitSelectorList(text string) []string {
	var items []string
	depth, start := 0, 0
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ',':
			if depth == 0 {
				items = append(items, text[start:i])
				start = i + 1
			}
		}
	}
	return append(items, text[start:])
}

// parseComplexSelector parses one combinator chain like `.panel > text`.
func parseComplexSelector(text string) (ComplexSelector, error) {
	var sel ComplexSelector
	pending := byte(' ')
	for _, tok := range tokenizeSelector(text) {
		if tok == ">" || tok == "+" || tok == "~" {
			pending = tok[0]
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

// tokenizeSelector splits a selector into compound and combinator tokens.
// Combinators (>, +, ~) become their own tokens even without surrounding
// spaces; characters inside () or [] never split (`:nth-child(2n+1)`).
func tokenizeSelector(text string) []string {
	var tokens []string
	var cur strings.Builder
	depth := 0
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(text); i++ {
		ch := text[i]
		switch {
		case ch == '(' || ch == '[':
			depth++
			cur.WriteByte(ch)
		case ch == ')' || ch == ']':
			depth--
			cur.WriteByte(ch)
		case depth == 0 && (ch == ' ' || ch == '\t'):
			flush()
		case depth == 0 && (ch == '>' || ch == '+' || ch == '~'):
			flush()
			tokens = append(tokens, string(ch))
		default:
			cur.WriteByte(ch)
		}
	}
	flush()
	return tokens
}

// parseSimpleSelector parses one compound like `box.panel#main[a=v]:hover`
// or `*`.
func parseSimpleSelector(tok string) (SimpleSelector, error) {
	var s SimpleSelector
	if tok == "*" {
		return s, nil
	}
	rest := tok
	// Leading tag name (up to the first qualifier).
	if i := strings.IndexAny(rest, ".#:["); i != 0 {
		if i < 0 {
			s.Tag = rest
			return s, nil
		}
		s.Tag = rest[:i]
		rest = rest[i:]
	}
	for rest != "" {
		kind := rest[0]
		if kind == '[' {
			end := strings.IndexByte(rest, ']')
			if end < 0 {
				return SimpleSelector{}, fmt.Errorf("unterminated attribute selector in %q", tok)
			}
			attr, err := parseAttrMatcher(rest[1:end])
			if err != nil {
				return SimpleSelector{}, err
			}
			s.Attrs = append(s.Attrs, attr)
			rest = rest[end+1:]
			continue
		}
		rest = rest[1:]
		end := qualifierEnd(rest)
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
			base, arg, hasArg := splitFunctionalPseudo(name)
			switch {
			case hasArg && (base == "not" || base == "is" || base == "where"):
				args, err := ParseSelectorList(arg)
				if err != nil {
					return SimpleSelector{}, fmt.Errorf("in :%s(): %w", base, err)
				}
				s.Logical = append(s.Logical, LogicalPseudo{Name: base, Args: args})
			case statePseudos[name]:
				s.Pseudo = name
			default:
				s.Structural = append(s.Structural, name)
			}
		}
	}
	return s, nil
}

// splitFunctionalPseudo splits "not(.a)" into ("not", ".a", true).
func splitFunctionalPseudo(name string) (string, string, bool) {
	open := strings.IndexByte(name, '(')
	if open < 0 || !strings.HasSuffix(name, ")") {
		return name, "", false
	}
	return name[:open], name[open+1 : len(name)-1], true
}

// qualifierEnd finds where the current qualifier name ends: the next
// qualifier marker at paren depth zero, so `nth-child(2n+1)` stays intact.
func qualifierEnd(rest string) int {
	depth := 0
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case '(':
			depth++
		case ')':
			depth--
		case '.', '#', ':', '[':
			if depth == 0 {
				return i
			}
		}
	}
	return len(rest)
}

// parseAttrMatcher parses the inside of an attribute selector: `name`,
// `name=value`, `name^=value`, ... Values may be single- or double-quoted.
func parseAttrMatcher(body string) (AttrMatcher, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return AttrMatcher{}, fmt.Errorf("empty attribute selector")
	}
	for _, op := range []string{"^=", "$=", "*=", "~=", "|=", "="} {
		if i := strings.Index(body, op); i >= 0 {
			name := strings.TrimSpace(body[:i])
			value := strings.TrimSpace(body[i+len(op):])
			value = strings.Trim(value, `"'`)
			if name == "" {
				return AttrMatcher{}, fmt.Errorf("attribute selector missing name: [%s]", body)
			}
			return AttrMatcher{Name: name, Op: op, Value: value}, nil
		}
	}
	return AttrMatcher{Name: body}, nil
}
