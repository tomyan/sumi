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
	p.pos++ // consume opening '<'
	tagName := p.readTagName()
	if tagName == "" {
		return nil, fmt.Errorf("empty tag name at position %d", p.pos)
	}
	switch tagName {
	case "text":
		return p.parseTextTag()
	case "box":
		return p.parseBoxElement()
	default:
		return p.parseComponentElement(tagName)
	}
}

// readTagName reads a tag name (until whitespace, '>', or '/').
func (p *parser) readTagName() string {
	return p.readWhile(func(b byte) bool {
		return b != '>' && b != '/' && !isWhitespace(b)
	})
}

// parseTextTag parses attributes and body of a <text> element.
func (p *parser) parseTextTag() (Node, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}
	if err := p.expectClose("text"); err != nil {
		return nil, err
	}
	return p.parseTextElement(attrs)
}

func (p *parser) parseBoxElement() (Node, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}
	if err := p.expectClose("box"); err != nil {
		return nil, err
	}
	children, err := p.parseChildren("box")
	if err != nil {
		return nil, err
	}
	if attrs == nil {
		attrs = make(map[string]string)
	}
	return &BoxElement{Attributes: attrs, Children: children}, nil
}

func (p *parser) parseComponentElement(name string) (Node, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}
	if attrs == nil {
		attrs = make(map[string]string)
	}
	if p.isSelfClosing() {
		return &ComponentElement{Name: name, Attributes: attrs}, nil
	}
	return p.parseComponentClosingTag(name, attrs)
}

func (p *parser) isSelfClosing() bool {
	if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '>' {
		p.pos += 2 // consume '/>'
		return true
	}
	return false
}

func (p *parser) parseComponentClosingTag(name string, attrs map[string]string) (Node, error) {
	if err := p.expectClose(name); err != nil {
		return nil, err
	}
	closingTag := "</" + name + ">"
	if !strings.HasPrefix(p.input[p.pos:], closingTag) {
		return nil, fmt.Errorf("expected closing </%s> tag", name)
	}
	p.pos += len(closingTag)
	return &ComponentElement{Name: name, Attributes: attrs}, nil
}

// expectClose expects and consumes a '>' to close an opening tag.
func (p *parser) expectClose(tagName string) error {
	if p.pos >= len(p.input) || p.input[p.pos] != '>' {
		return fmt.Errorf("expected '>' to close <%s> tag", tagName)
	}
	p.pos++ // consume '>'
	return nil
}

// parseChildren parses child elements until the closing tag </tagName> is found.
func (p *parser) parseChildren(tagName string) ([]Node, error) {
	closingTag := "</" + tagName + ">"
	var children []Node
	for {
		child, done, err := p.parseNextChild(closingTag, tagName)
		if err != nil {
			return nil, err
		}
		if done {
			return children, nil
		}
		children = append(children, child)
	}
}

// parseNextChild parses one child element or detects the closing tag.
// Returns (nil, true, nil) when the closing tag is consumed.
func (p *parser) parseNextChild(closingTag, tagName string) (Node, bool, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, false, fmt.Errorf("missing closing </%s> tag", tagName)
	}
	if strings.HasPrefix(p.input[p.pos:], closingTag) {
		p.pos += len(closingTag)
		return nil, true, nil
	}
	if p.input[p.pos] != '<' {
		return nil, false, fmt.Errorf("unexpected character %q inside <%s> at position %d", p.input[p.pos], tagName, p.pos)
	}
	child, err := p.parseElement()
	return child, false, err
}

func (p *parser) parseAttributes() (map[string]string, error) {
	attrs := make(map[string]string)
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unexpected end of input in tag attributes")
		}
		if p.input[p.pos] == '>' || p.input[p.pos] == '/' {
			break
		}
		name, value, err := p.readAttribute()
		if err != nil {
			return nil, err
		}
		if name == "" {
			break
		}
		attrs[name] = value
	}
	return attrs, nil
}

// readAttribute reads a single name="value" attribute pair.
func (p *parser) readAttribute() (string, string, error) {
	name := p.readWhile(func(b byte) bool {
		return b != '=' && b != '>' && !isWhitespace(b)
	})
	if name == "" {
		return "", "", nil
	}
	if err := p.expectByte('=', "after attribute name %q", name); err != nil {
		return "", "", err
	}
	value, err := p.readQuotedValue(name)
	if err != nil {
		return "", "", err
	}
	return name, value, nil
}

// readQuotedValue reads a double-quoted attribute value.
func (p *parser) readQuotedValue(attrName string) (string, error) {
	if err := p.expectByte('"', "for attribute %q value", attrName); err != nil {
		return "", err
	}
	value := p.readUntil('"')
	if p.pos >= len(p.input) {
		return "", fmt.Errorf("unterminated attribute value for %q", attrName)
	}
	p.pos++ // consume closing '"'
	return value, nil
}

// expectByte checks the current byte matches ch and consumes it.
func (p *parser) expectByte(ch byte, context string, args ...any) error {
	if p.pos >= len(p.input) || p.input[p.pos] != ch {
		msg := fmt.Sprintf(context, args...)
		return fmt.Errorf("expected %q %s", ch, msg)
	}
	p.pos++
	return nil
}

func (p *parser) parseTextElement(attrs map[string]string) (Node, error) {
	closingTag := "</text>"
	closeIdx := strings.Index(p.input[p.pos:], closingTag)
	if closeIdx == -1 {
		return nil, fmt.Errorf("missing closing </text> tag")
	}
	content := p.input[p.pos : p.pos+closeIdx]
	p.pos += closeIdx + len(closingTag)
	parts := parseTextParts(content)
	return &TextElement{Attributes: attrs, Parts: parts}, nil
}

// parseTextParts splits text content into StringPart and ExprPart segments.
func parseTextParts(content string) []Part {
	if content == "" {
		return nil
	}
	var parts []Part
	for len(content) > 0 {
		part, rest, done := extractNextPart(content)
		parts = append(parts, part...)
		content = rest
		if done {
			break
		}
	}
	return parts
}

// extractNextPart extracts the next part(s) from content.
// Returns the parts found, remaining content, and whether parsing is done.
func extractNextPart(content string) ([]Part, string, bool) {
	openIdx := strings.Index(content, "{")
	if openIdx == -1 {
		return []Part{&StringPart{Value: content}}, "", true
	}
	var parts []Part
	if openIdx > 0 {
		parts = append(parts, &StringPart{Value: content[:openIdx]})
	}
	closeIdx := strings.Index(content[openIdx:], "}")
	if closeIdx == -1 {
		parts = append(parts, &StringPart{Value: content[openIdx:]})
		return parts, "", true
	}
	expr := strings.TrimSpace(content[openIdx+1 : openIdx+closeIdx])
	parts = append(parts, &ExprPart{Expr: expr})
	return parts, content[openIdx+closeIdx+1:], false
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
