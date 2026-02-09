package script

import "fmt"

// tryParsePropDecl tries to parse: name := $prop(expr)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParsePropDecl() (PropDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return PropDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "$prop(") {
		p.pos = saved
		return PropDecl{}, false, nil
	}

	expr, err := p.readBalancedParenContents()
	if err != nil {
		return PropDecl{}, false, fmt.Errorf("unterminated $prop expression for %q: %w", name, err)
	}

	return PropDecl{Name: name, DefaultExpr: expr}, true, nil
}
