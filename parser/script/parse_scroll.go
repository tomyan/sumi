package script

import "fmt"

// tryParseScrollDecl tries to parse: name := $scroll(boxId)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParseScrollDecl() (ScrollDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return ScrollDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "$scroll(") {
		p.pos = saved
		return ScrollDecl{}, false, nil
	}

	boxID, err := p.readBalancedParenContents()
	if err != nil {
		return ScrollDecl{}, false, fmt.Errorf("unterminated $scroll expression for %q: %w", name, err)
	}

	return ScrollDecl{Name: name, BoxID: boxID}, true, nil
}
