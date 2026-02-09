package layout

import (
	"strconv"
	"strings"
)

// NodeKind distinguishes text nodes from box containers.
type NodeKind int

const (
	KindBox  NodeKind = iota
	KindText
)

// Input describes a node in the layout tree before layout is computed.
type Input struct {
	Kind        NodeKind
	Content     string   // text content (KindText only)
	Direction   string   // "column" (default) or "row" (future)
	FixedWidth  int      // 0 = auto
	FixedHeight int      // 0 = auto
	Padding     Padding
	Border      string   // "single", "none", or ""
	Children    []*Input
}

// Padding holds inset values for each side.
type Padding struct {
	Top, Right, Bottom, Left int
}

// Box is a laid-out node with computed position and size.
type Box struct {
	X, Y, Width, Height int
	Children            []*Box
	Content             string // text content if text node
	Border              string // border style
}

// ParsePadding parses a CSS-like padding shorthand string.
// Supported formats:
//   - ""        → all zero
//   - "1"       → all sides 1
//   - "1 2"     → top/bottom=1, left/right=2
//   - "1 2 3 4" → top=1, right=2, bottom=3, left=4
func ParsePadding(s string) Padding {
	s = strings.TrimSpace(s)
	if s == "" {
		return Padding{}
	}

	parts := strings.Fields(s)
	vals := make([]int, len(parts))
	for i, p := range parts {
		v, _ := strconv.Atoi(p)
		vals[i] = v
	}

	switch len(vals) {
	case 1:
		return Padding{vals[0], vals[0], vals[0], vals[0]}
	case 2:
		return Padding{vals[0], vals[1], vals[0], vals[1]}
	case 4:
		return Padding{vals[0], vals[1], vals[2], vals[3]}
	default:
		return Padding{}
	}
}

// hasBorder returns true if the input has a visible border.
func hasBorder(border string) bool {
	return border != "" && border != "none"
}

// borderSize returns the number of cells consumed by the border on one side.
func borderSize(border string) int {
	if hasBorder(border) {
		return 1
	}
	return 0
}

// Layout computes positions and sizes for a tree of Input nodes.
func Layout(input *Input, availWidth, availHeight int) *Box {
	return layoutNode(input)
}

func layoutNode(input *Input) *Box {
	box := &Box{
		Border: input.Border,
	}

	if input.Kind == KindText {
		box.Content = input.Content
		box.Width = len(input.Content)
		box.Height = 1
		if input.FixedWidth > 0 {
			box.Width = input.FixedWidth
		}
		if input.FixedHeight > 0 {
			box.Height = input.FixedHeight
		}
		return box
	}

	// KindBox: lay out children in column direction
	b := borderSize(input.Border)
	pad := input.Padding

	// Inset from the edge to the content area
	offsetX := b + pad.Left
	offsetY := b + pad.Top

	// Lay out children vertically within the content area
	cursorY := 0
	maxChildWidth := 0
	for _, child := range input.Children {
		childBox := layoutNode(child)
		childBox.X = offsetX
		childBox.Y = offsetY + cursorY
		cursorY += childBox.Height
		if childBox.Width > maxChildWidth {
			maxChildWidth = childBox.Width
		}
		box.Children = append(box.Children, childBox)
	}

	// Compute size
	if input.FixedWidth > 0 {
		box.Width = input.FixedWidth
	} else {
		box.Width = maxChildWidth + pad.Left + pad.Right + 2*b
	}

	if input.FixedHeight > 0 {
		box.Height = input.FixedHeight
	} else {
		box.Height = cursorY + pad.Top + pad.Bottom + 2*b
	}

	return box
}
