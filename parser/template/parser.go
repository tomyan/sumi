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

	// Read tag name (until whitespace or '>')
	tagName := p.readWhile(func(b byte) bool {
		return b != '>' && !isWhitespace(b)
	})
	if tagName == "" {
		return nil, fmt.Errorf("empty tag name at position %d", p.pos)
	}

	switch tagName {
	case "text":
		// Skip any remaining content up to '>'
		p.readUntil('>')
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unexpected end of input in opening tag")
		}
		p.pos++ // consume '>'
		return p.parseTextElement()
	case "box":
		return p.parseBoxElement()
	default:
		return nil, fmt.Errorf("unknown element <%s>", tagName)
	}
}

func (p *parser) parseBoxElement() (Node, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}

	// Expect '>' to close the opening tag
	if p.pos >= len(p.input) || p.input[p.pos] != '>' {
		return nil, fmt.Errorf("expected '>' to close <box> tag")
	}
	p.pos++ // consume '>'

	// Parse children until </box>
	var children []Node
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("missing closing </box> tag")
		}

		// Check for closing tag
		if strings.HasPrefix(p.input[p.pos:], "</box>") {
			p.pos += len("</box>")
			break
		}

		if p.input[p.pos] != '<' {
			return nil, fmt.Errorf("unexpected character %q inside <box> at position %d", p.input[p.pos], p.pos)
		}

		child, err := p.parseElement()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	if attrs == nil {
		attrs = make(map[string]string)
	}

	return &BoxElement{Attributes: attrs, Children: children}, nil
}

func (p *parser) parseAttributes() (map[string]string, error) {
	attrs := make(map[string]string)
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unexpected end of input in tag attributes")
		}
		if p.input[p.pos] == '>' {
			break
		}

		// Read attribute name
		name := p.readWhile(func(b byte) bool {
			return b != '=' && b != '>' && !isWhitespace(b)
		})
		if name == "" {
			break
		}

		// Expect '='
		if p.pos >= len(p.input) || p.input[p.pos] != '=' {
			return nil, fmt.Errorf("expected '=' after attribute name %q", name)
		}
		p.pos++ // consume '='

		// Expect '"'
		if p.pos >= len(p.input) || p.input[p.pos] != '"' {
			return nil, fmt.Errorf("expected '\"' for attribute %q value", name)
		}
		p.pos++ // consume opening '"'

		// Read until closing '"'
		value := p.readUntil('"')
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unterminated attribute value for %q", name)
		}
		p.pos++ // consume closing '"'

		attrs[name] = value
	}
	return attrs, nil
}

func (p *parser) parseTextElement() (Node, error) {
	closingTag := "</text>"
	closeIdx := strings.Index(p.input[p.pos:], closingTag)
	if closeIdx == -1 {
		return nil, fmt.Errorf("missing closing </text> tag")
	}

	content := p.input[p.pos : p.pos+closeIdx]
	p.pos += closeIdx + len(closingTag)

	parts := parseTextParts(content)
	return &TextElement{Parts: parts}, nil
}

// parseTextParts splits text content into StringPart and ExprPart segments.
func parseTextParts(content string) []Part {
	if content == "" {
		return nil
	}

	var parts []Part
	for len(content) > 0 {
		openIdx := strings.Index(content, "{")
		if openIdx == -1 {
			// No more expressions — rest is a string part
			parts = append(parts, &StringPart{Value: content})
			break
		}
		// Text before the '{'
		if openIdx > 0 {
			parts = append(parts, &StringPart{Value: content[:openIdx]})
		}
		// Find closing '}'
		closeIdx := strings.Index(content[openIdx:], "}")
		if closeIdx == -1 {
			// No closing brace — treat rest as literal text
			parts = append(parts, &StringPart{Value: content[openIdx:]})
			break
		}
		expr := strings.TrimSpace(content[openIdx+1 : openIdx+closeIdx])
		parts = append(parts, &ExprPart{Expr: expr})
		content = content[openIdx+closeIdx+1:]
	}
	return parts
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

func (p *parser) readWhile(pred func(byte) bool) string {
	start := p.pos
	for p.pos < len(p.input) && pred(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
