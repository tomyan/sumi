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

// Part is the interface for parts of text content.
type Part interface {
	partType() string
}

// StringPart represents literal text content.
type StringPart struct {
	Value string
}

func (s *StringPart) partType() string { return "string" }

// ExprPart represents an {expression} in text content.
type ExprPart struct {
	Expr string
}

func (e *ExprPart) partType() string { return "expr" }

// TextElement represents a <text>content</text> element.
type TextElement struct {
	Attributes map[string]string
	Parts      []Part
}

func (t *TextElement) nodeType() string { return "text" }

// BoxElement represents a <box>...</box> container element with attributes.
type BoxElement struct {
	Attributes map[string]string
	Children   []Node
}

func (b *BoxElement) nodeType() string { return "box" }

// ComponentElement represents a user-defined component reference like <counter />.
type ComponentElement struct {
	Name       string
	Attributes map[string]string
}

func (c *ComponentElement) nodeType() string { return "component" }

// TitleElement represents a <title>content</title> element for setting the terminal title.
type TitleElement struct {
	Parts []Part
}

func (t *TitleElement) nodeType() string { return "title" }

// IfNode represents an {if condition}...{else}...{/if} block.
type IfNode struct {
	Condition string // raw Go expression: "count > 0"
	Then      []Node // body when true
	Else      []Node // body when false (nil if no {else})
}

func (n *IfNode) nodeType() string { return "if" }

// ForNode represents a {for clause}...{/for} loop block.
type ForNode struct {
	Clause   string // raw Go clause: "i, item := range items"
	Children []Node
}

func (n *ForNode) nodeType() string { return "for" }
