package script

import "fmt"

// tryParseSelfDecl tries to parse: name := $self(key)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParseSelfDecl() (SelfDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return SelfDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "$self(") {
		p.pos = saved
		return SelfDecl{}, false, nil
	}

	key, err := p.readBalancedParenContents()
	if err != nil {
		return SelfDecl{}, false, fmt.Errorf("unterminated $self expression for %q: %w", name, err)
	}

	return SelfDecl{Name: name, Key: key}, true, nil
}
