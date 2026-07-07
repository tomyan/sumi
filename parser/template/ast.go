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

// TextElement represents a text-bearing element: <text>, or any HTML
// element whose body is plain text/expressions (<h1>Hi {name}</h1>).
type TextElement struct {
	Tag        string // HTML tag name; "" means legacy <text>
	Attributes map[string]string
	Parts      []Part
}

func (t *TextElement) nodeType() string { return "text" }

// BoxElement represents a container element: <box>, or any HTML element
// with element children (<div>, <ul>, ...).
type BoxElement struct {
	Tag        string // HTML tag name; "" means legacy <box>
	Attributes map[string]string
	Children   []Node
}

func (b *BoxElement) nodeType() string { return "box" }

// ComponentElement represents a user-defined component reference like <counter />.
// Children holds the tag body: {snippet} blocks become snippet props and the
// remaining content becomes the implicit "children" snippet.
type ComponentElement struct {
	Name       string
	Attributes map[string]string
	Children   []Node
}

func (c *ComponentElement) nodeType() string { return "component" }

// TitleElement represents a <title>content</title> element for setting the terminal title.
type TitleElement struct {
	Parts []Part
}

func (t *TitleElement) nodeType() string { return "title" }

// SnippetNode represents a {snippet name(params)}...{/snippet} template function.
type SnippetNode struct {
	Name     string // function name
	Params   string // "(name string)" parameter list
	Children []Node // template body
}

func (s *SnippetNode) nodeType() string { return "snippet" }

// RenderNode represents a {render name(args)} invocation of a snippet.
type RenderNode struct {
	Name string // snippet name
	Args string // argument expression
}

func (r *RenderNode) nodeType() string { return "render" }

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
