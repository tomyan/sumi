package script

import "fmt"

// tryParseSignalDecl tries to parse: name := sumi.New(expr) or name := signal.New(expr)
func (p *parser) tryParseSignalDecl() (SignalDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return SignalDecl{}, false, nil
	}

	// Try sumi.New( or signal.New(
	if !p.matchInlineSequence(":=", "sumi.New(") && !p.matchInlineSequenceFrom(saved+len(name), ":=", "signal.New(") {
		p.pos = saved
		return SignalDecl{}, false, nil
	}

	expr, err := p.readBalancedParenContents()
	if err != nil {
		return SignalDecl{}, false, fmt.Errorf("unterminated sumi.New expression for %q: %w", name, err)
	}

	return SignalDecl{Name: name, InitExpr: expr}, true, nil
}

// tryParseComputedDecl tries to parse: name := sumi.From(expr) or name := signal.From(expr)
func (p *parser) tryParseComputedDecl() (ComputedDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return ComputedDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "sumi.From(") && !p.matchInlineSequenceFrom(saved+len(name), ":=", "signal.From(") {
		p.pos = saved
		return ComputedDecl{}, false, nil
	}

	expr, err := p.readBalancedParenContents()
	if err != nil {
		return ComputedDecl{}, false, fmt.Errorf("unterminated sumi.From expression for %q: %w", name, err)
	}

	return ComputedDecl{Name: name, Expr: expr}, true, nil
}

// matchInlineSequenceFrom resets to a given position and tries matching.
func (p *parser) matchInlineSequenceFrom(pos int, tokens ...string) bool {
	p.pos = pos
	return p.matchInlineSequence(tokens...)
}
