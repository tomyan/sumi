package layout

import (
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
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
	Content     string       // text content (KindText only)
	Direction   string       // "column" (default) or "row"
	FixedWidth  int          // 0 = auto
	FixedHeight int          // 0 = auto
	Gap         int          // space between children (cells)
	FlexGrow    int          // flex-grow factor (0 = no grow)
	Justify     string       // main-axis alignment: start, end, center, space-between
	Align       string       // cross-axis alignment: start, end, center, stretch
	Padding     Padding
	Border      string       // "single", "none", or ""
	Style       render.Style // resolved style for this node
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
	Content             string       // text content if text node
	Border              string       // border style
	Style               render.Style // visual style
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
	return layoutNode(input, availWidth, availHeight)
}

func layoutNode(input *Input, availW, availH int) *Box {
	box := &Box{
		Border: input.Border,
		Style:  input.Style,
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

	b := borderSize(input.Border)
	pad := input.Padding

	// Inset from the edge to the content area
	offsetX := b + pad.Left
	offsetY := b + pad.Top
	insetW := pad.Left + pad.Right + 2*b
	insetH := pad.Top + pad.Bottom + 2*b

	// Determine the content area dimensions available
	contentAvailW := availW - insetW
	contentAvailH := availH - insetH
	if input.FixedWidth > 0 {
		contentAvailW = input.FixedWidth - insetW
	}
	if input.FixedHeight > 0 {
		contentAvailH = input.FixedHeight - insetH
	}

	hasFlexChildren := hasFlexGrow(input.Children)

	if input.Direction == "row" {
		if hasFlexChildren {
			box.Children = layoutRowFlex(input.Children, offsetX, offsetY, input.Gap, contentAvailW, contentAvailH)
		} else {
			box.Children = layoutRow(input.Children, offsetX, offsetY, input.Gap, contentAvailW, contentAvailH)
		}
	} else {
		if hasFlexChildren {
			box.Children = layoutColumnFlex(input.Children, offsetX, offsetY, input.Gap, contentAvailW, contentAvailH)
		} else {
			box.Children = layoutColumn(input.Children, offsetX, offsetY, input.Gap, contentAvailW, contentAvailH)
		}
	}

	// Apply justify (shift children along main axis)
	if input.Justify != "" && input.Justify != "start" {
		if input.Direction == "row" {
			applyJustifyRow(box.Children, offsetX, contentAvailW, input.Justify)
		} else {
			applyJustifyColumn(box.Children, offsetY, contentAvailH, input.Justify)
		}
	}

	// Apply align (shift/stretch children along cross axis)
	// Default is "stretch" to match CSS flexbox behavior
	align := input.Align
	if align == "" {
		align = "stretch"
	}
	if align != "start" {
		if input.Direction == "row" {
			applyAlignRow(box.Children, input.Children, offsetY, contentAvailH, align)
		} else {
			applyAlignColumn(box.Children, input.Children, offsetX, contentAvailW, align)
		}
	}

	// Compute size
	contentW, contentH := childrenExtent(box.Children, offsetX, offsetY)

	if input.FixedWidth > 0 {
		box.Width = input.FixedWidth
	} else if hasFlexChildren && contentAvailW > 0 {
		// If children use flex-grow, parent expands to available width
		box.Width = availW
	} else {
		box.Width = contentW + insetW
	}
	if input.FixedHeight > 0 {
		box.Height = input.FixedHeight
	} else if hasFlexChildren && input.Direction != "row" && contentAvailH > 0 {
		box.Height = availH
	} else {
		box.Height = contentH + insetH
	}

	return box
}

// hasFlexGrow returns true if any child has FlexGrow > 0.
func hasFlexGrow(children []*Input) bool {
	for _, c := range children {
		if c.FlexGrow > 0 {
			return true
		}
	}
	return false
}

// layoutColumn places children vertically, advancing Y after each child.
func layoutColumn(children []*Input, offsetX, offsetY, gap, availW, availH int) []*Box {
	var boxes []*Box
	cursorY := 0
	for i, child := range children {
		if i > 0 && gap > 0 {
			cursorY += gap
		}
		childBox := layoutNode(child, availW, availH)
		childBox.X = offsetX
		childBox.Y = offsetY + cursorY
		cursorY += childBox.Height
		boxes = append(boxes, childBox)
	}
	return boxes
}

// layoutRow places children horizontally, advancing X after each child.
func layoutRow(children []*Input, offsetX, offsetY, gap, availW, availH int) []*Box {
	var boxes []*Box
	cursorX := 0
	for i, child := range children {
		if i > 0 && gap > 0 {
			cursorX += gap
		}
		childBox := layoutNode(child, availW, availH)
		childBox.X = offsetX + cursorX
		childBox.Y = offsetY
		cursorX += childBox.Width
		boxes = append(boxes, childBox)
	}
	return boxes
}

// layoutRowFlex is a two-pass row layout that distributes extra space to flex-grow children.
func layoutRowFlex(children []*Input, offsetX, offsetY, gap, availW, availH int) []*Box {
	// Pass 1: lay out non-flex children to get their natural width
	naturalWidths := make([]int, len(children))
	totalFixed := 0
	totalGaps := 0
	totalFlex := 0
	for i, child := range children {
		if i > 0 {
			totalGaps += gap
		}
		if child.FlexGrow > 0 {
			totalFlex += child.FlexGrow
		} else {
			childBox := layoutNode(child, availW, availH)
			naturalWidths[i] = childBox.Width
			totalFixed += childBox.Width
		}
	}

	// Pass 2: distribute remaining space among flex children
	remaining := availW - totalFixed - totalGaps
	if remaining < 0 {
		remaining = 0
	}

	boxes := make([]*Box, len(children))
	cursorX := 0
	for i, child := range children {
		if i > 0 {
			cursorX += gap
		}
		var childBox *Box
		if child.FlexGrow > 0 {
			flexWidth := remaining * child.FlexGrow / totalFlex
			childBox = layoutNode(child, flexWidth, availH)
			childBox.Width = flexWidth
		} else {
			childBox = layoutNode(child, naturalWidths[i], availH)
		}
		childBox.X = offsetX + cursorX
		childBox.Y = offsetY
		cursorX += childBox.Width
		boxes[i] = childBox
	}
	return boxes
}

// layoutColumnFlex is a two-pass column layout that distributes extra space to flex-grow children.
func layoutColumnFlex(children []*Input, offsetX, offsetY, gap, availW, availH int) []*Box {
	// Pass 1: lay out non-flex children to get their natural height
	naturalHeights := make([]int, len(children))
	totalFixed := 0
	totalGaps := 0
	totalFlex := 0
	for i, child := range children {
		if i > 0 {
			totalGaps += gap
		}
		if child.FlexGrow > 0 {
			totalFlex += child.FlexGrow
		} else {
			childBox := layoutNode(child, availW, availH)
			naturalHeights[i] = childBox.Height
			totalFixed += childBox.Height
		}
	}

	// Pass 2: distribute remaining space among flex children
	remaining := availH - totalFixed - totalGaps
	if remaining < 0 {
		remaining = 0
	}

	boxes := make([]*Box, len(children))
	cursorY := 0
	for i, child := range children {
		if i > 0 {
			cursorY += gap
		}
		var childBox *Box
		if child.FlexGrow > 0 {
			flexHeight := remaining * child.FlexGrow / totalFlex
			childBox = layoutNode(child, availW, flexHeight)
			childBox.Height = flexHeight
		} else {
			childBox = layoutNode(child, availW, availH)
		}
		childBox.X = offsetX
		childBox.Y = offsetY + cursorY
		cursorY += childBox.Height
		boxes[i] = childBox
	}
	return boxes
}

// applyJustifyRow shifts children along the X axis within the available width.
func applyJustifyRow(boxes []*Box, offsetX, availW int, justify string) {
	if len(boxes) == 0 {
		return
	}
	totalChildWidth := 0
	for _, b := range boxes {
		totalChildWidth += b.Width
	}
	// Include gaps that are already in the positions
	lastChild := boxes[len(boxes)-1]
	usedWidth := (lastChild.X - offsetX) + lastChild.Width
	remaining := availW - usedWidth
	if remaining <= 0 {
		return
	}
	applyJustify(boxes, remaining, justify, true, offsetX)
}

// applyJustifyColumn shifts children along the Y axis within the available height.
func applyJustifyColumn(boxes []*Box, offsetY, availH int, justify string) {
	if len(boxes) == 0 {
		return
	}
	lastChild := boxes[len(boxes)-1]
	usedHeight := (lastChild.Y - offsetY) + lastChild.Height
	remaining := availH - usedHeight
	if remaining <= 0 {
		return
	}
	applyJustify(boxes, remaining, justify, false, offsetY)
}

// applyJustify shifts boxes along an axis. isRow=true shifts X, isRow=false shifts Y.
func applyJustify(boxes []*Box, remaining int, justify string, isRow bool, offset int) {
	n := len(boxes)
	switch justify {
	case "end":
		for _, b := range boxes {
			if isRow {
				b.X += remaining
			} else {
				b.Y += remaining
			}
		}
	case "center":
		shift := remaining / 2
		for _, b := range boxes {
			if isRow {
				b.X += shift
			} else {
				b.Y += shift
			}
		}
	case "space-between":
		if n <= 1 {
			return
		}
		gaps := n - 1
		for i, b := range boxes {
			// Distribute remaining space evenly among gaps
			shift := remaining * i / gaps
			if isRow {
				b.X += shift
			} else {
				b.Y += shift
			}
		}
	}
}

// applyAlignRow aligns children along the Y axis (cross axis for row layout).
func applyAlignRow(boxes []*Box, inputs []*Input, offsetY, availH int, align string) {
	for i, b := range boxes {
		switch align {
		case "end":
			b.Y = offsetY + availH - b.Height
		case "center":
			b.Y = offsetY + (availH-b.Height)/2
		case "stretch":
			if i < len(inputs) && !canStretch(inputs[i], false) {
				continue
			}
			b.Y = offsetY
			b.Height = availH
		}
	}
}

// applyAlignColumn aligns children along the X axis (cross axis for column layout).
func applyAlignColumn(boxes []*Box, inputs []*Input, offsetX, availW int, align string) {
	for i, b := range boxes {
		switch align {
		case "end":
			b.X = offsetX + availW - b.Width
		case "center":
			b.X = offsetX + (availW-b.Width)/2
		case "stretch":
			if i < len(inputs) && !canStretch(inputs[i], true) {
				continue
			}
			b.X = offsetX
			b.Width = availW
		}
	}
}

// canStretch returns whether a child can be stretched on the cross axis.
// Text nodes have intrinsic size and are never stretched.
// Children with explicit fixed cross-axis size are not stretched.
func canStretch(input *Input, isWidth bool) bool {
	if input.Kind == KindText {
		return false
	}
	if isWidth && input.FixedWidth > 0 {
		return false
	}
	if !isWidth && input.FixedHeight > 0 {
		return false
	}
	return true
}

// childrenExtent returns the content width and height occupied by children.
// Works generically by computing the bounding box of children relative to offset.
func childrenExtent(boxes []*Box, offsetX, offsetY int) (int, int) {
	maxW := 0
	maxH := 0
	for _, b := range boxes {
		right := (b.X - offsetX) + b.Width
		bottom := (b.Y - offsetY) + b.Height
		if right > maxW {
			maxW = right
		}
		if bottom > maxH {
			maxH = bottom
		}
	}
	return maxW, maxH
}
