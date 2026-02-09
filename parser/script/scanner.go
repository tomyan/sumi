package script

import (
	"fmt"
	"unicode"
)

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.pos++
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (p *parser) skipLine() {
	for p.pos < len(p.input) && p.input[p.pos] != '\n' {
		p.pos++
	}
	if p.pos < len(p.input) {
		p.pos++ // skip the newline
	}
}

func (p *parser) skipInlineWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t') {
		p.pos++
	}
}

func (p *parser) readIdentifier() string {
	start := p.pos
	if p.pos >= len(p.input) {
		return ""
	}
	if !unicode.IsLetter(rune(p.input[p.pos])) && p.input[p.pos] != '_' {
		return ""
	}
	p.pos++
	for p.pos < len(p.input) && isIdentChar(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func (p *parser) matchString(s string) bool {
	if p.pos+len(s) <= len(p.input) && p.input[p.pos:p.pos+len(s)] == s {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *parser) readUntilByte(b byte) (string, error) {
	start := p.pos
	for p.pos < len(p.input) {
		if p.input[p.pos] == b {
			return p.input[start:p.pos], nil
		}
		p.pos++
	}
	return "", fmt.Errorf("unexpected end of input, expected %q", string(b))
}
