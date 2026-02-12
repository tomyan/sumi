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
	Key         string       // identity key for diffing (set by keyed {for} loops)
	Direction   string       // "column" (default) or "row"
	FixedWidth  int          // 0 = auto
	FixedHeight int          // 0 = auto
	Gap         int          // space between children (cells)
	FlexGrow    int          // flex-grow factor (0 = no grow)
	Justify     string       // main-axis alignment: start, end, center, space-between
	Align       string       // cross-axis alignment: start, end, center, stretch
	Overflow    string       // "hidden", "scroll", "auto", or "" (visible)
	MinWidth    int          // minimum content width (0 = no minimum)
	Display     string       // "" (default) or "none" (hidden from layout)
	Position    string       // "" (static), "relative", "absolute", "fixed", "sticky"
	ZIndex      int          // paint order (higher renders on top)
	Top         int          // offset from top (for positioned elements)
	Left        int          // offset from left (for positioned elements)
	Right       int          // offset from right (for positioned elements)
	Bottom      int          // offset from bottom (for positioned elements)
	SelfW       *int    // if non-nil, receives computed width after layout
	SelfH       *int    // if non-nil, receives computed height after layout
	SelfX       *int    // if non-nil, receives absolute X position after layout
	SelfY       *int    // if non-nil, receives absolute Y position after layout
	CursorCol   int     // cursor column within content (-1 = no cursor)
	CursorRow   int     // cursor row within content (-1 = no cursor)
	Focusable   bool    // true if this element can receive focus
	FocusIndex  int     // assigned focus index (0-based) for Tab cycling
	Padding     Padding
	Border         string       // "single", "none", or ""
	BorderTitle    string       // text to display in the top border edge
	BorderCollapse bool        // when true, children share borders
	Style          render.Style // resolved style for this node
	Children       []*Input
}

// Padding holds inset values for each side.
type Padding struct {
	Top, Right, Bottom, Left int
}

// Box is a laid-out node with computed position and size.
type Box struct {
	X, Y, Width, Height int
	ContentWidth        int          // full content width (set when overflow != "")
	ContentHeight       int          // full content height (set when overflow != "")
	ScrollY             int          // vertical scroll offset (applied during render)
	ScrollX             int          // horizontal scroll offset (applied during render)
	Key                 string       // identity key for diffing (propagated from Input)
	Position            string       // positioning mode (propagated from Input)
	ZIndex              int          // paint order (propagated from Input)
	Top                 int          // offset from top (propagated from Input)
	Left                int          // offset from left (propagated from Input)
	Right               int          // offset from right (propagated from Input)
	Bottom              int          // offset from bottom (propagated from Input)
	Children            []*Box
	Content             string       // text content if text node
	Lines               []string     // wrapped lines (nil = single line, use Content)
	Border              string              // border style
	BorderTitle         string              // text to display in the top border edge
	Collapsed           render.CollapsedEdges // edges shared with adjacent borders
	Style               render.Style        // visual style
	CursorCol                int          // cursor column within content (-1 = no cursor)
	CursorRow                int          // cursor row within content (-1 = no cursor)
	NeedsScrollbar           bool         // true when a vertical scrollbar should be drawn
	NeedsHorizontalScrollbar bool         // true when a horizontal scrollbar should be drawn
	Clip                     *render.Clip // clipping region (set when overflow is non-empty)
	HasOverlap               bool         // true when any descendant has absolute/fixed or non-zero z-index
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
// All positions in the returned tree are absolute (buffer coordinates).
func Layout(input *Input, availWidth, availHeight int) *Box {
	box := layoutNode(input, availWidth, availHeight)
	absolutePositions(box)
	repositionFixed(box, availWidth, availHeight)
	writeSelfPositions(input, box)
	return box
}

// writeSelfPositions writes absolute X/Y through SelfX/SelfY pointers
// after absolutePositions has converted to screen coordinates.
func writeSelfPositions(input *Input, box *Box) {
	if input == nil || box == nil {
		return
	}
	if input.SelfX != nil {
		*input.SelfX = box.X
	}
	if input.SelfY != nil {
		*input.SelfY = box.Y
	}
	for i, child := range input.Children {
		if i < len(box.Children) {
			writeSelfPositions(child, box.Children[i])
		}
	}
}

// absolutePositions converts relative positions to absolute by accumulating
// parent offsets down the tree. Also adjusts clip regions to absolute coordinates.
func absolutePositions(box *Box) {
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		// Fixed children are viewport-relative; skip parent offset accumulation
		if child.Position == "fixed" {
			absolutePositions(child)
			continue
		}
		child.X += box.X
		child.Y += box.Y
		if child.Clip != nil {
			child.Clip.Left += child.X
			child.Clip.Right += child.X
			child.Clip.Top += child.Y
			child.Clip.Bottom += child.Y
		}
		absolutePositions(child)
	}
}

func layoutNode(input *Input, availW, availH int) *Box {
	border := input.Border
	if input.BorderCollapse {
		border = "" // children's borders form the parent frame
	}
	box := &Box{
		Border:      border,
		BorderTitle: input.BorderTitle,
		Key:         input.Key,
		Position:    input.Position,
		ZIndex:      input.ZIndex,
		Top:         input.Top,
		Left:        input.Left,
		Right:       input.Right,
		Bottom:      input.Bottom,
		CursorCol:   input.CursorCol,
		CursorRow:   input.CursorRow,
		Style:       input.Style,
	}

	if input.Kind == KindText {
		box.CursorCol = -1
		box.CursorRow = -1
		box.Content = input.Content
		if availW > 0 && len(input.Content) > availW {
			lines := wrapText(input.Content, availW)
			box.Lines = lines
			box.Width = availW
			box.Height = len(lines)
		} else {
			box.Width = len(input.Content)
			box.Height = 1
		}
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

	// When border-collapse is active, children fill the entire parent area
	if input.BorderCollapse {
		b = 0
	}

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

	// For scroll/auto overflow, give children unbounded space
	childAvailH := contentAvailH
	if isScrollOverflow(input.Overflow) {
		childAvailH = 1000000
	}

	// Apply min-width: if available width is below min-width, use min-width for content
	if input.MinWidth > 0 && contentAvailW < input.MinWidth {
		contentAvailW = input.MinWidth
	}

	// Filter out display:none children
	visibleChildren, visibleIndices := filterVisible(input.Children)

	// Partition visible children into flow and positioned (absolute/fixed)
	flowChildren, flowIndices, posChildren, posIndices := partitionPositioned(visibleChildren)

	hasFlexChildren := hasFlexGrow(flowChildren)

	// Border-collapse forces gap to 0 and inflates available space to compensate for overlaps
	gap := input.Gap
	flexAvailW := contentAvailW
	flexAvailH := childAvailH
	if input.BorderCollapse {
		gap = 0
		overlaps := countOverlaps(flowChildren)
		if input.Direction == "row" {
			flexAvailW += overlaps
		} else {
			flexAvailH += overlaps
		}
	}

	var flowBoxes []*Box
	if input.Direction == "row" {
		if hasFlexChildren {
			flowBoxes = layoutRowFlex(flowChildren, offsetX, offsetY, gap, flexAvailW, flexAvailH)
		} else {
			flowBoxes = layoutRow(flowChildren, offsetX, offsetY, gap, flexAvailW, flexAvailH)
		}
	} else {
		if hasFlexChildren {
			flowBoxes = layoutColumnFlex(flowChildren, offsetX, offsetY, gap, flexAvailW, flexAvailH)
		} else {
			flowBoxes = layoutColumn(flowChildren, offsetX, offsetY, gap, flexAvailW, flexAvailH)
		}
	}

	// Apply border-collapse: overlap adjacent bordered children
	if input.BorderCollapse {
		if input.Direction == "row" {
			targetW := 0
			if input.FixedWidth > 0 || hasFlexChildren {
				targetW = contentAvailW
			}
			applyRowCollapse(flowBoxes, flowChildren, targetW)
		} else {
			targetH := 0
			if input.FixedHeight > 0 || hasFlexChildren {
				targetH = contentAvailH
			}
			applyColumnCollapse(flowBoxes, flowChildren, targetH)
		}
	}

	// Apply justify (shift children along main axis).
	// Skip when the container auto-sizes on the main axis — no free space to distribute.
	if input.Justify != "" && input.Justify != "start" {
		if input.Direction == "row" {
			applyJustifyRow(flowBoxes, offsetX, contentAvailW, input.Justify)
		} else if input.FixedHeight > 0 {
			applyJustifyColumn(flowBoxes, offsetY, contentAvailH, input.Justify)
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
			crossSize := rowCrossSize(contentAvailH, input.FixedHeight, flowBoxes)
			applyAlignRow(flowBoxes, flowChildren, offsetY, crossSize, align)
		} else {
			applyAlignColumn(flowBoxes, flowChildren, offsetX, contentAvailW, align)
		}
	}

	// Update self-measurement pointers after alignment may have changed dimensions
	updateSelfPointers(flowChildren, flowBoxes)

	// Layout positioned (absolute/fixed) children
	posBoxes := layoutPositionedChildren(posChildren, offsetX, offsetY, contentAvailW, contentAvailH)

	// Merge flow and positioned boxes back into visible order, then splice into full array
	visibleBoxes := mergePartitioned(flowBoxes, flowIndices, posBoxes, posIndices, len(visibleChildren))
	box.Children = spliceChildren(len(input.Children), visibleBoxes, visibleIndices)

	// Propagate HasOverlap from children
	box.HasOverlap = computeHasOverlap(box.Children)

	// Compute size from flow children only (positioned elements don't affect parent size)
	contentW, contentH := childrenExtent(flowBoxes, offsetX, offsetY)

	if input.FixedWidth > 0 {
		box.Width = input.FixedWidth
	} else if hasFlexChildren && contentAvailW > 0 {
		// If children use flex-grow, parent expands to available width
		box.Width = availW
	} else {
		box.Width = contentW + insetW
	}
	// When scroll overflow is active, cap width to available (content scrolls horizontally)
	if isScrollOverflow(input.Overflow) && box.Width > availW {
		box.Width = availW
	}
	if input.FixedHeight > 0 {
		box.Height = input.FixedHeight
	} else if hasFlexChildren && input.Direction != "row" && contentAvailH > 0 {
		box.Height = availH
	} else {
		box.Height = contentH + insetH
	}
	// Scroll viewport fills available height when no fixed height is set
	if isScrollOverflow(input.Overflow) && input.FixedHeight == 0 {
		box.Height = availH
	}

	if input.Overflow != "" {
		box.Clip = computeClip(box, b, pad)
	}
	if isScrollOverflow(input.Overflow) {
		box.ContentWidth = contentW
		box.ContentHeight = contentH
		box.NeedsScrollbar = needsScrollbar(input.Overflow, contentH, contentAvailH)
		viewportW := box.Width - insetW
		box.NeedsHorizontalScrollbar = needsHorizontalScrollbar(input.Overflow, contentW, viewportW)
	}

	// Apply relative offsets after size is computed (visual shift only)
	applyRelativeOffsets(box.Children)

	// Write back self-measurement pointers
	if input.SelfW != nil {
		*input.SelfW = box.Width
	}
	if input.SelfH != nil {
		*input.SelfH = box.Height
	}

	return box
}

// updateSelfPointers re-writes SelfW/SelfH pointers after alignment may have
// changed a child's dimensions (e.g., stretch in a column stretches width).
func updateSelfPointers(inputs []*Input, boxes []*Box) {
	for i, inp := range inputs {
		if inp == nil || i >= len(boxes) || boxes[i] == nil {
			continue
		}
		if inp.SelfW != nil {
			*inp.SelfW = boxes[i].Width
		}
		if inp.SelfH != nil {
			*inp.SelfH = boxes[i].Height
		}
	}
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

// rowCrossSize returns the cross-axis height for row alignment.
// When the row has a fixed height, the content area height is used.
// When auto-height, the tallest child determines the stretch target,
// matching CSS flexbox behavior where auto-height rows size to content.
func rowCrossSize(contentAvailH, fixedHeight int, children []*Box) int {
	if fixedHeight > 0 {
		return contentAvailH
	}
	maxH := 0
	for _, child := range children {
		if child.Height > maxH {
			maxH = child.Height
		}
	}
	return maxH
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

// computeHasOverlap returns true if any child has absolute/fixed positioning,
// non-zero z-index, or itself has HasOverlap set.
func computeHasOverlap(children []*Box) bool {
	for _, child := range children {
		if child == nil {
			continue
		}
		if child.Position == "absolute" || child.Position == "fixed" || child.ZIndex != 0 {
			return true
		}
		if child.HasOverlap {
			return true
		}
	}
	return false
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
