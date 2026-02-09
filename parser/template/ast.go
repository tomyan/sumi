package template

// Node is the interface all AST nodes implement.
type Node interface {
	nodeType() string
}

// Document is the root node containing top-level elements.
type Document struct {
	Children []Node
}

func (d *Document) nodeType() string { return "document" }

// TextElement represents a <text>content</text> element.
type TextElement struct {
	Content string
}

func (t *TextElement) nodeType() string { return "text" }

// BoxElement represents a <box>...</box> container element with attributes.
type BoxElement struct {
	Attributes map[string]string
	Children   []Node
}

func (b *BoxElement) nodeType() string { return "box" }
