package style

import (
	"fmt"
	"strings"
)

// Stylesheet represents the parsed contents of a <style> block.
type Stylesheet struct {
	Rules []Rule
}

// Rule represents a single CSS-like rule with a selector and properties.
type Rule struct {
	Selector   string            // e.g. ".title", "text", ".container"
	Properties map[string]string // e.g. {"color": "green", "bold": "true"}
}

// Parse parses CSS-like style rules from the content of a <style> block.
func Parse(input string) (*Stylesheet, error) {
	p := &parser{input: input, pos: 0}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() (*Stylesheet, error) {
	s := &Stylesheet{}

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}

		// Check for comment that wasn't fully consumed (unterminated)
		if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '*' {
			return nil, fmt.Errorf("unterminated comment")
		}

		rule, err := p.parseRule()
		if err != nil {
			return nil, err
		}
		s.Rules = append(s.Rules, rule)
	}

	return s, nil
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.pos++
	}
}

func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '*' {
			if err := p.skipComment(); err != nil {
				return // leave position at the comment start for error reporting
			}
		} else {
			return
		}
	}
}

func (p *parser) skipComment() error {
	p.pos += 2 // skip /*
	for p.pos+1 < len(p.input) {
		if p.input[p.pos] == '*' && p.input[p.pos+1] == '/' {
			p.pos += 2
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated comment")
}

func (p *parser) parseRule() (Rule, error) {
	// Parse selector
	selector := p.parseSelector()
	if selector == "" {
		return Rule{}, fmt.Errorf("expected selector at position %d", p.pos)
	}

	p.skipWhitespace()

	// Expect opening brace
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return Rule{}, fmt.Errorf("expected '{' after selector %q", selector)
	}
	p.pos++ // skip {

	// Parse properties
	props, err := p.parseProperties()
	if err != nil {
		return Rule{}, err
	}

	return Rule{Selector: selector, Properties: props}, nil
}

func (p *parser) parseSelector() string {
	p.skipWhitespace()
	start := p.pos

	if p.pos < len(p.input) && p.input[p.pos] == '.' {
		// Class selector: .classname
		p.pos++
		for p.pos < len(p.input) && isSelectorChar(p.input[p.pos]) {
			p.pos++
		}
	} else {
		// Element selector: text, box
		for p.pos < len(p.input) && isSelectorChar(p.input[p.pos]) {
			p.pos++
		}
	}

	return p.input[start:p.pos]
}

func (p *parser) parseProperties() (map[string]string, error) {
	props := make(map[string]string)

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unterminated rule block, expected '}'")
		}

		// Check for closing brace
		if p.input[p.pos] == '}' {
			p.pos++ // skip }
			return props, nil
		}

		// Parse property name
		name := p.readPropertyName()
		if name == "" {
			return nil, fmt.Errorf("expected property name at position %d", p.pos)
		}

		p.skipWhitespace()

		// Expect colon
		if p.pos >= len(p.input) || p.input[p.pos] != ':' {
			return nil, fmt.Errorf("expected ':' after property %q", name)
		}
		p.pos++ // skip :

		p.skipWhitespace()

		// Read value (up to ; or })
		value := p.readPropertyValue()

		props[name] = value
	}

	return nil, fmt.Errorf("unterminated rule block, expected '}'")
}

func (p *parser) readPropertyName() string {
	start := p.pos
	for p.pos < len(p.input) && isPropertyNameChar(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) readPropertyValue() string {
	var b strings.Builder
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ';' {
			p.pos++ // consume semicolon
			break
		}
		if ch == '}' {
			// Don't consume }, leave it for the caller
			break
		}
		b.WriteByte(ch)
		p.pos++
	}
	return strings.TrimSpace(b.String())
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isSelectorChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' || b == '-'
}

func isPropertyNameChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == '_'
}
