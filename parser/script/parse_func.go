package script

import (
	"fmt"
	"strings"
)

// tryParseFuncDecl tries to parse: func name(params) { body }
func (p *parser) tryParseFuncDecl() (FuncDecl, bool, error) {
	saved := p.pos

	if !p.matchFuncKeyword() {
		return p.resetAndNoMatch(saved)
	}
	p.skipInlineWhitespace()

	name := p.readIdentifier()
	if name == "" {
		return p.resetAndNoMatch(saved)
	}

	return p.completeFuncDecl(saved, name)
}

// completeFuncDecl finishes parsing a function declaration after keyword and name.
func (p *parser) completeFuncDecl(saved int, name string) (FuncDecl, bool, error) {
	params, ok, err := p.readFuncParams(name)
	if err != nil {
		return FuncDecl{}, false, err
	}
	if !ok {
		return p.resetAndNoMatch(saved)
	}

	returnType, body, ok, err := p.readReturnTypeAndBody(name)
	if err != nil {
		return FuncDecl{}, false, err
	}
	if !ok {
		return p.resetAndNoMatch(saved)
	}

	return FuncDecl{Name: name, Params: params, ReturnType: returnType, Body: body}, true, nil
}

// resetAndNoMatch resets the parser position and returns a "no match" result.
func (p *parser) resetAndNoMatch(saved int) (FuncDecl, bool, error) {
	p.pos = saved
	return FuncDecl{}, false, nil
}

// matchFuncKeyword matches "func" followed by a non-identifier character.
func (p *parser) matchFuncKeyword() bool {
	if !p.matchString("func") {
		return false
	}
	return p.pos >= len(p.input) || !isIdentChar(p.input[p.pos])
}

// readFuncParams reads the parameter list between parens.
func (p *parser) readFuncParams(name string) (string, bool, error) {
	p.skipInlineWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		return "", false, nil
	}
	p.pos++ // skip (

	params, err := p.readUntilByte(')')
	if err != nil {
		return "", false, fmt.Errorf("unterminated function parameter list for %q", name)
	}
	p.pos++ // skip )
	return params, true, nil
}

// readReturnTypeAndBody reads the optional return type and body.
// Returns (returnType, body, matched, error).
func (p *parser) readReturnTypeAndBody(name string) (string, string, bool, error) {
	p.skipInlineWhitespace()

	// Capture any return type between ) and {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '{' && p.input[p.pos] != '\n' {
		p.pos++
	}
	returnType := strings.TrimSpace(p.input[start:p.pos])

	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return "", "", false, nil
	}

	body, err := p.readBalancedBraceContents()
	if err != nil {
		return "", "", false, fmt.Errorf("unterminated function body for %q: %w", name, err)
	}
	return returnType, body, true, nil
}
