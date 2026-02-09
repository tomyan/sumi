package script

import "fmt"

// readBalancedParenContents reads content inside parens, handling nested parens and strings.
// Assumes the opening '(' has already been consumed. Consumes the closing ')'.
func (p *parser) readBalancedParenContents() (string, error) {
	return p.readBalancedContents('(', ')')
}

// readBalancedBraceContents reads content inside braces, handling nested braces and strings.
// Assumes the caller has NOT consumed the opening '{'. Consumes both '{' and '}'.
func (p *parser) readBalancedBraceContents() (string, error) {
	p.pos++ // skip past opening {
	return p.readBalancedContents('{', '}')
}

// readBalancedContents reads content between balanced open/close delimiters,
// handling nested delimiters and string literals.
// Assumes the opening delimiter has already been consumed.
func (p *parser) readBalancedContents(open, close byte) (string, error) {
	depth := 1
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		switch {
		case ch == open:
			depth++
		case ch == close:
			if depth == 1 {
				return p.consumeBalancedResult(start)
			}
			depth--
		default:
			if err := p.skipStringOrAdvance(); err != nil {
				return "", err
			}
			continue
		}
		p.pos++
	}
	return "", fmt.Errorf("unexpected end of input, expected %q", string(close))
}

// consumeBalancedResult extracts the content and consumes the closing delimiter.
func (p *parser) consumeBalancedResult(start int) (string, error) {
	result := p.input[start:p.pos]
	p.pos++
	return result, nil
}

// skipStringOrAdvance skips a string literal if at a quote character,
// otherwise advances by one byte.
func (p *parser) skipStringOrAdvance() error {
	if skipped, err := p.skipStringLiteral(); err != nil {
		return err
	} else if !skipped {
		p.pos++
	}
	return nil
}
