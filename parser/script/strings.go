package script

import "fmt"

// skipStringLiteral dispatches to the appropriate string-skipping function
// based on the quote character. Returns true if a string literal was skipped.
func (p *parser) skipStringLiteral() (bool, error) {
	if p.pos >= len(p.input) {
		return false, nil
	}
	switch p.input[p.pos] {
	case '"':
		p.pos++
		return true, p.skipDoubleQuotedString()
	case '`':
		p.pos++
		return true, p.skipBacktickString()
	case '\'':
		p.pos++
		return true, p.skipSingleQuotedChar()
	default:
		return false, nil
	}
}

func (p *parser) skipDoubleQuotedString() error {
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' {
			p.pos += 2 // skip escape sequence
			continue
		}
		if p.input[p.pos] == '"' {
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated string literal")
}

func (p *parser) skipBacktickString() error {
	for p.pos < len(p.input) {
		if p.input[p.pos] == '`' {
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated raw string literal")
}

func (p *parser) skipSingleQuotedChar() error {
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' {
			p.pos += 2
			continue
		}
		if p.input[p.pos] == '\'' {
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated character literal")
}
