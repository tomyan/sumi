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

// SlotElement represents a <slot:name /> placeholder in a component template.
// Default contains fallback content rendered when the consumer doesn't provide content.
type SlotElement struct {
	Name       string            // slot name: "header", "children", etc.
	Attributes map[string]string // scoped slot params passed to consumer
	Default    []Node            // default content (nil if self-closing)
}

func (s *SlotElement) nodeType() string { return "slot" }

// SlotDefNode represents a {slot name}...{/slot} content definition from a consumer.
type SlotDefNode struct {
	Name     string // slot name to fill: "header", "children"
	Params   string // scoped slot params: "(item Item, i int)" or ""
	Children []Node // content to render in the slot
}

func (s *SlotDefNode) nodeType() string { return "slotdef" }

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
	Key      string // key expression for diffing: "item.ID" (empty if no key)
	Children []Node
}

func (n *ForNode) nodeType() string { return "for" }
