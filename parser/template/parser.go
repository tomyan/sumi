package template

import (
	"fmt"
	"strings"
)

// Parse parses template markup into a Document AST.
func Parse(input string) (*Document, error) {
	p := &parser{input: input}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() (*Document, error) {
	doc := &Document{}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		if p.input[p.pos] != '<' {
			return nil, fmt.Errorf("unexpected character %q at position %d", p.input[p.pos], p.pos)
		}

		node, err := p.parseElement()
		if err != nil {
			return nil, err
		}
		doc.Children = append(doc.Children, node)
	}

	return doc, nil
}

func (p *parser) parseElement() (Node, error) {
	// Consume opening '<'
	p.pos++

	// Read tag name
	tagName := p.readUntil('>')
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of input in opening tag")
	}
	p.pos++ // consume '>'

	switch tagName {
	case "text":
		return p.parseTextElement()
	default:
		return nil, fmt.Errorf("unknown element <%s>", tagName)
	}
}

func (p *parser) parseTextElement() (Node, error) {
	closingTag := "</text>"
	closeIdx := strings.Index(p.input[p.pos:], closingTag)
	if closeIdx == -1 {
		return nil, fmt.Errorf("missing closing </text> tag")
	}

	content := p.input[p.pos : p.pos+closeIdx]
	p.pos += closeIdx + len(closingTag)

	return &TextElement{Content: content}, nil
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.pos++
	}
}

func (p *parser) readUntil(ch byte) string {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != ch {
		p.pos++
	}
	return p.input[start:p.pos]
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
