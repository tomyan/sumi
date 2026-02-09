package script

import "fmt"

// tryParseStateDecl tries to parse: name := $state(expr)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParseStateDecl() (StateDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return StateDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "$state(") {
		p.pos = saved
		return StateDecl{}, false, nil
	}

	expr, err := p.readBalancedParenContents()
	if err != nil {
		return StateDecl{}, false, fmt.Errorf("unterminated $state expression for %q: %w", name, err)
	}

	return StateDecl{Name: name, InitExpr: expr}, true, nil
}

// matchInlineSequence matches a sequence of whitespace-separated tokens.
func (p *parser) matchInlineSequence(tokens ...string) bool {
	for _, token := range tokens {
		p.skipInlineWhitespace()
		if !p.matchString(token) {
			return false
		}
	}
	return true
}
