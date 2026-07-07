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
		if p.input[p.pos] == '{' {
			node, err := p.parseControlFlow()
			if err != nil {
				return nil, err
			}
			doc.Children = append(doc.Children, node)
			continue
		}
		if p.input[p.pos] != '<' {
			return nil, p.errorf("unexpected character %q at position %d", p.input[p.pos], p.pos)
		}
		node, err := p.parseElement()
		if err != nil {
			return nil, err
		}
		doc.Children = append(doc.Children, node)
	}
	if err := validateSingleTitle(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// validateSingleTitle ensures at most one <title> element exists in the document.
func validateSingleTitle(doc *Document) error {
	count := 0
	for _, child := range doc.Children {
		if _, ok := child.(*TitleElement); ok {
			count++
		}
	}
	if count > 1 {
		return fmt.Errorf("only one <title> element is allowed, found %d", count)
	}
	return nil
}

func (p *parser) parseElement() (Node, error) {
	p.pos++ // consume opening '<'
	tagName := p.readTagName()
	if tagName == "" {
		return nil, p.errorf("empty tag name at position %d", p.pos)
	}
	switch {
	case tagName == "text":
		return nil, p.errorf("the <text> tag was removed; use <span> (or another HTML text element)")
	case tagName == "box":
		return nil, p.errorf("the <box> tag was removed; use <div> (or another HTML container element)")
	case tagName == "title":
		return p.parseTitleElement()
	case strings.HasPrefix(tagName, "slot:"):
		return p.parseSlotElement(tagName)
	case isHTMLTagName(tagName):
		return p.parseHTMLElement(tagName)
	default:
		return p.parseComponentElement(tagName)
	}
}

// htmlTags is the recognised HTML element vocabulary. Anything else
// lowercase remains a component reference (legacy components like
// <textedit> keep working until the C1c migration).
var htmlTags = map[string]bool{
	"div": true, "span": true, "p": true, "h1": true, "h2": true,
	"h3": true, "h4": true, "h5": true, "h6": true, "ul": true,
	"ol": true, "li": true, "blockquote": true, "pre": true,
	"code": true, "hr": true, "section": true, "main": true,
	"header": true, "footer": true, "nav": true, "article": true,
	"aside": true, "strong": true, "b": true, "em": true, "i": true,
	"u": true, "s": true, "del": true, "mark": true, "kbd": true,
	"abbr": true, "samp": true, "var": true, "a": true, "label": true,
	"table": true, "thead": true, "tbody": true, "tfoot": true,
	"tr": true, "td": true, "th": true, "caption": true,
	"button": true, "input": true, "textarea": true, "select": true,
	"option": true, "progress": true, "meter": true, "details": true,
	"summary": true, "dialog": true, "img": true,
	"ansi": true, "region": true, "colgroup": true, "col": true,
	"optgroup": true,
}

// isHTMLTagName reports whether a tag is a recognised HTML element name.
func isHTMLTagName(tag string) bool {
	return htmlTags[tag]
}

// containerTags are HTML elements that default to container (box) form
// even when empty. Text bodies become implicit untagged children so the
// element itself keeps borders, padding, and pseudo-element markers.
var containerTags = map[string]bool{
	"div": true, "section": true, "main": true, "header": true,
	"footer": true, "nav": true, "article": true, "aside": true,
	"ul": true, "ol": true, "li": true, "blockquote": true, "table": true,
	"thead": true, "tbody": true, "tfoot": true, "tr": true,
	"button": true, "textarea": true, "dialog": true, "details": true,
	"select": true, "ansi": true, "region": true, "colgroup": true,
	"optgroup": true,
}

// parseHTMLElement parses any HTML element. Bodies containing child
// elements or control flow become containers (BoxElement); plain
// text/expression bodies become text elements. Self-closing is allowed.
func (p *parser) parseHTMLElement(tag string) (Node, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}
	if attrs == nil {
		attrs = make(map[string]string)
	}
	if p.isSelfClosing() {
		return &BoxElement{Tag: tag, Attributes: attrs}, nil
	}
	if err := p.expectClose(tag); err != nil {
		return nil, err
	}
	if p.htmlBodyIsText(tag) {
		node, err := p.parseTextBody(tag, attrs)
		if err != nil {
			return nil, err
		}
		if containerTags[tag] {
			// Container tags keep their box form so pseudo-element markers
			// and borders work; the text body becomes an implicit child.
			text := node.(*TextElement)
			text.Attributes = map[string]string{}
			text.Tag = ""
			return &BoxElement{Tag: tag, Attributes: attrs, Children: []Node{text}}, nil
		}
		return retagText(node, tag), nil
	}
	children, err := p.parseChildren(tag)
	if err != nil {
		return nil, err
	}
	return &BoxElement{Tag: tag, Attributes: attrs, Children: children}, nil
}

// htmlBodyIsText reports whether the element body up to the closing tag is
// plain text/expressions (no child elements or control-flow blocks).
func (p *parser) htmlBodyIsText(tag string) bool {
	rest := p.input[p.pos:]
	lt := strings.Index(rest, "<")
	if lt < 0 {
		return false
	}
	if !strings.HasPrefix(rest[lt:], "</"+tag+">") {
		return false
	}
	body := rest[:lt]
	for _, marker := range []string{"{if ", "{for ", "{render ", "{snippet ", "{slot"} {
		if strings.Contains(body, marker) {
			return false
		}
	}
	if strings.TrimSpace(body) == "" {
		return !containerTags[tag]
	}
	return true
}

// retagText stamps the HTML tag onto a parsed text element.
func retagText(node Node, tag string) Node {
	if t, ok := node.(*TextElement); ok {
		t.Tag = tag
	}
	return node
}

// readTagName reads a tag name (until whitespace, '>', or '/').
func (p *parser) readTagName() string {
	return p.readWhile(func(b byte) bool {
		return b != '>' && b != '/' && !isWhitespace(b)
	})
}

func (p *parser) parseTitleElement() (Node, error) {
	if err := p.expectClose("title"); err != nil {
		return nil, err
	}
	closingTag := "</title>"
	closeIdx := strings.Index(p.input[p.pos:], closingTag)
	if closeIdx == -1 {
		return nil, p.errorf("missing closing </title> tag")
	}
	content := p.input[p.pos : p.pos+closeIdx]
	p.pos += closeIdx + len(closingTag)
	parts := parseTextParts(content)
	return &TitleElement{Parts: parts}, nil
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
		return nil, p.errorf("expected closing </%s> tag", name)
	}
	p.pos += len(closingTag)
	return &ComponentElement{Name: name, Attributes: attrs}, nil
}

// expectClose expects and consumes a '>' to close an opening tag.
func (p *parser) expectClose(tagName string) error {
	if p.pos >= len(p.input) || p.input[p.pos] != '>' {
		return p.errorf("expected '>' to close <%s> tag", tagName)
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

// parseNextChild parses one child element, control-flow block, or loose
// text run, or detects the closing tag. Dropped formatting gaps loop
// until a real child or the closing tag is found.
// Returns (nil, true, nil) when the closing tag is consumed.
func (p *parser) parseNextChild(closingTag, tagName string) (Node, bool, error) {
	for {
		if p.pos >= len(p.input) {
			return nil, false, p.errorf("missing closing </%s> tag", tagName)
		}
		if strings.HasPrefix(p.input[p.pos:], closingTag) {
			p.pos += len(closingTag)
			return nil, true, nil
		}
		if p.input[p.pos] == '<' {
			child, err := p.parseElement()
			return child, false, err
		}
		if p.input[p.pos] == '{' && p.controlFlowStart() {
			child, err := p.parseControlFlow()
			return child, false, err
		}
		if node := p.parseLooseText(); node != nil {
			return node, false, nil
		}
	}
}

func (p *parser) parseAttributes() (map[string]string, error) {
	attrs := make(map[string]string)
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, p.errorf("unexpected end of input in tag attributes")
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

// readAttribute reads an attribute in one of three forms:
//   - name="value"  → literal string
//   - name={expr}   → expression (stored as "{expr}")
//   - {name}        → shorthand for name={name}
func (p *parser) readAttribute() (string, string, error) {
	// Shorthand: {name} → name={name}
	if p.pos < len(p.input) && p.input[p.pos] == '{' {
		expr, err := p.readBracedValue()
		if err != nil {
			return "", "", err
		}
		return expr, "{" + expr + "}", nil
	}

	name := p.readWhile(func(b byte) bool {
		return b != '=' && b != '>' && b != '/' && !isWhitespace(b)
	})
	if name == "" {
		return "", "", nil
	}
	if p.pos >= len(p.input) || p.input[p.pos] != '=' {
		return "", "", p.errorf("expected '=' after attribute name %q", name)
	}
	p.pos++ // consume '='

	// Expression value: name={expr}
	if p.pos < len(p.input) && p.input[p.pos] == '{' {
		expr, err := p.readBracedValue()
		if err != nil {
			return "", "", err
		}
		return name, "{" + expr + "}", nil
	}

	// Quoted value: name="value"
	value, err := p.readQuotedValue(name)
	if err != nil {
		return "", "", err
	}
	return name, value, nil
}

// readBracedValue reads content between { and }, consuming both delimiters.
func (p *parser) readBracedValue() (string, error) {
	if err := p.expectByte('{', "for expression value"); err != nil {
		return "", err
	}
	depth := 1
	start := p.pos
	for p.pos < len(p.input) && depth > 0 {
		if p.input[p.pos] == '{' {
			depth++
		} else if p.input[p.pos] == '}' {
			depth--
		}
		if depth > 0 {
			p.pos++
		}
	}
	if depth != 0 {
		return "", p.errorf("unterminated expression value")
	}
	expr := p.input[start:p.pos]
	p.pos++ // consume closing '}'
	return expr, nil
}

// readQuotedValue reads a double-quoted attribute value.
func (p *parser) readQuotedValue(attrName string) (string, error) {
	if err := p.expectByte('"', "for attribute %q value", attrName); err != nil {
		return "", err
	}
	value := p.readUntil('"')
	if p.pos >= len(p.input) {
		return "", p.errorf("unterminated attribute value for %q", attrName)
	}
	p.pos++ // consume closing '"'
	return value, nil
}

// expectByte checks the current byte matches ch and consumes it.
func (p *parser) expectByte(ch byte, context string, args ...any) error {
	if p.pos >= len(p.input) || p.input[p.pos] != ch {
		msg := fmt.Sprintf(context, args...)
		return p.errorf("expected %q %s", ch, msg)
	}
	p.pos++
	return nil
}

func (p *parser) parseTextElement(attrs map[string]string) (Node, error) {
	return p.parseTextBody("text", attrs)
}

// parseTextBody reads text/expression content up to the closing tag.
func (p *parser) parseTextBody(tag string, attrs map[string]string) (Node, error) {
	closingTag := "</" + tag + ">"
	closeIdx := strings.Index(p.input[p.pos:], closingTag)
	if closeIdx == -1 {
		return nil, p.errorf("missing closing %s tag", closingTag)
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
