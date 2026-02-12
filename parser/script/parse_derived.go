package script

import "fmt"

// tryParseDerivedDecl tries to parse: name := $derived(expr)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParseDerivedDecl() (DerivedDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return DerivedDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "$derived(") {
		p.pos = saved
		return DerivedDecl{}, false, nil
	}

	expr, err := p.readBalancedParenContents()
	if err != nil {
		return DerivedDecl{}, false, fmt.Errorf("unterminated $derived expression for %q: %w", name, err)
	}

	return DerivedDecl{Name: name, Expr: expr}, true, nil
}
