package layout

import (
	"fmt"
	"sort"

	"github.com/tomyan/sumi/runtime/anim"
	"github.com/tomyan/sumi/runtime/render"
)

// RenderTree renders a layout tree to a buffer, applying clip regions and scroll offsets.
func RenderTree(buf *render.Buffer, box *Box, clip *render.Clip) {
	counter := 0
	renderTreeFull(buf, box, clip, render.Style{}, nil, &counter)
}

// RenderTreeWithEngine renders with animation engine support.
func RenderTreeWithEngine(buf *render.Buffer, box *Box, clip *render.Clip, engine *anim.Engine) {
	counter := 0
	renderTreeFull(buf, box, clip, render.Style{}, engine, &counter)
}

func renderTreeFull(buf *render.Buffer, box *Box, clip *render.Clip, inherited render.Style, engine *anim.Engine, counter *int) {
	nodeID := fmt.Sprintf("n%d", *counter)
	*counter++

	if box.Visibility == "hidden" {
		return // occupies layout space but paints nothing (children included)
	}

	// Apply hover: when hovered and a hover style exists, use it instead.
	if box.Hovered && !box.HoverStyle.IsZero() {
		box.Style = box.HoverStyle
	}
	if box.Focused && !box.FocusStyle.IsZero() {
		box.Style = box.FocusStyle
	}

	// Apply animation engine.
	if engine != nil {
		if len(box.Transitions) > 0 {
			box.Style = engine.BeforeRender(nodeID, box.Style, box.Transitions)
		}
		if box.AnimationSpec != nil {
			box.Style = engine.BeforeRenderAnim(nodeID, box.Style, box.AnimationSpec)
		}
	}

	// Apply style inheritance: merge parent's inheritable properties into this node.
	box.Style = box.Style.Inherit(inherited)

	renderBackground(buf, box, clip)
	renderBorder(buf, box)
	renderContent(buf, box, clip)
	renderCells(buf, box, clip)
	renderScrollbarsAndChildrenFull(buf, box, clip, box.Style, engine, counter)
}

// renderCells blits per-cell styled content (ansi/region elements) at the
// box's content origin, clipped to the content area.
func renderCells(buf *render.Buffer, box *Box, clip *render.Clip) {
	if box.Cells == nil {
		return
	}
	b := borderSize(box.Border)
	originX := box.X + b + box.Padding.Left
	originY := box.Y + b + box.Padding.Top
	maxW := box.Width - 2*b - box.Padding.Left - box.Padding.Right
	maxH := box.Height - 2*b - box.Padding.Top - box.Padding.Bottom
	for row := 0; row < box.Cells.Height() && row < maxH; row++ {
		for col := 0; col < box.Cells.Width() && col < maxW; col++ {
			c := box.Cells.Cell(row, col)
			buf.SetStyledCellClipped(originY+row, originX+col, c.Ch, c.Style, clip)
		}
	}
}

// renderBackground fills the box area with spaces using the box's BG color.
// Only fills when a background color is set. Fills inside border if present.
func renderBackground(buf *render.Buffer, box *Box, clip *render.Clip) {
	if box.Style.BG.Name == "" && !box.Style.BG.IsRGB {
		return
	}
	b := borderSize(box.Border)
	for row := box.Y + b; row < box.Y+box.Height-b; row++ {
		for col := box.X + b; col < box.X+box.Width-b; col++ {
			buf.SetStyledCellClipped(row, col, ' ', box.Style, clip)
		}
	}
}

func renderBorder(buf *render.Buffer, box *Box) {
	if box.Border != "" && box.Border != "none" {
		if !box.Collapsed.IsZero() {
			buf.DrawCollapsedBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style, box.Collapsed)
		} else {
			buf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)
		}
		if box.BorderTitle != "" {
			buf.DrawStyledBorderTitle(box.Y, box.X, box.Width, box.BorderTitle, box.Style)
		}
		return
	}
	// Partial borders: top and/or bottom only (no corners, no side edges).
	if hasBorder(box.BorderTop) {
		for c := box.X; c < box.X+box.Width; c++ {
			buf.SetStyledCell(box.Y, c, '─', box.Style)
		}
	}
	if hasBorder(box.BorderBottom) {
		bottomRow := box.Y + box.Height - 1
		for c := box.X; c < box.X+box.Width; c++ {
			buf.SetStyledCell(bottomRow, c, '─', box.Style)
		}
	}
}

func renderContent(buf *render.Buffer, box *Box, clip *render.Clip) {
	if box.Fragments != nil {
		for _, f := range box.Fragments {
			buf.WriteStyledTextClipped(box.Y+f.Y, box.X+f.X, f.Text, box.Style, clip)
		}
	} else if box.Lines != nil {
		for i, line := range box.Lines {
			line = fitLine(line, box.Width, box.TextOverflow)
			buf.WriteStyledTextClipped(box.Y+i, box.X+alignShift(line, box.Width, box.TextAlign), line, box.Style, clip)
		}
	} else if box.Content != "" {
		line := fitLine(box.Content, box.Width, box.TextOverflow)
		buf.WriteStyledTextClipped(box.Y, box.X+alignShift(line, box.Width, box.TextAlign), line, box.Style, clip)
	}
	// Render inverse cursor for contenteditable elements.
	if box.ContentEditable && box.CursorCol >= 0 && box.CursorRow >= 0 {
		cursorY := box.Y + box.CursorRow
		cursorX := box.X + box.CursorCol
		cell := buf.Cell(cursorY, cursorX)
		ch := cell.Ch
		if ch == 0 {
			ch = ' '
		}
		cursorStyle := box.Style
		cursorStyle.Inverse = true
		buf.SetStyledCell(cursorY, cursorX, ch, cursorStyle)
	}
}

// alignShift computes the X offset that aligns a line within the box width.
func alignShift(line string, width int, align string) int {
	if align != "center" && align != "right" {
		return 0
	}
	pad := width - runeLen(line)
	if pad <= 0 {
		return 0
	}
	if align == "center" {
		return pad / 2
	}
	return pad
}

// fitLine truncates an overflowing line per text-overflow.
func fitLine(line string, width int, overflow string) string {
	if overflow == "" || width <= 0 {
		return line
	}
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	switch overflow {
	case "ellipsis":
		if width == 1 {
			return "…"
		}
		return string(runes[:width-1]) + "…"
	case "ellipsis-middle":
		if width <= 2 {
			return string(runes[:width])
		}
		head := (width - 1) / 2
		tail := width - 1 - head
		return string(runes[:head]) + "…" + string(runes[len(runes)-tail:])
	}
	return string(runes[:width])
}

func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// renderScrollbarsAndChildren draws scrollbars (if needed) then renders children
// with the content clip narrowed to avoid overlap with scrollbars.
func renderScrollbarsAndChildren(buf *render.Buffer, box *Box, clip *render.Clip) {
	renderScrollbarsAndChildrenWithInherit(buf, box, clip, render.Style{})
}

func renderScrollbarsAndChildrenFull(buf *render.Buffer, box *Box, clip *render.Clip, inherited render.Style, engine *anim.Engine, counter *int) {
	childClip := mergeClip(clip, box.Clip)
	if box.NeedsScrollbar && box.Clip != nil {
		drawVerticalScrollbar(buf, box)
		childClip = narrowClipForVerticalScrollbar(childClip)
	}
	if box.NeedsHorizontalScrollbar && box.Clip != nil {
		drawHorizontalScrollbar(buf, box)
		childClip = narrowClipForHorizontalScrollbar(childClip)
	}
	sorted := zSortChildren(box.Children)
	for _, child := range sorted {
		if child.Position == "fixed" {
			renderTreeFull(buf, child, nil, inherited, engine, counter)
			continue
		}
		renderChildWithScrollFull(buf, child, box.ScrollX, box.ScrollY, childClip, inherited, engine, counter)
	}
}

func renderChildWithScrollFull(buf *render.Buffer, child *Box, scrollX, scrollY int, clip *render.Clip, inherited render.Style, engine *anim.Engine, counter *int) {
	if scrollX == 0 && scrollY == 0 {
		renderTreeFull(buf, child, clip, inherited, engine, counter)
		return
	}
	shiftTree(child, -scrollX, -scrollY)
	stickyDY := applyStickyClamp(child, clip)
	renderTreeFull(buf, child, clip, inherited, engine, counter)
	shiftTree(child, scrollX, scrollY-stickyDY)
}

// renderTreeWithInherit is a backward-compat wrapper for code that doesn't use the engine.
func renderTreeWithInherit(buf *render.Buffer, box *Box, clip *render.Clip, inherited render.Style) {
	counter := 0
	renderTreeFull(buf, box, clip, inherited, nil, &counter)
}

func renderScrollbarsAndChildrenWithInherit(buf *render.Buffer, box *Box, clip *render.Clip, inherited render.Style) {
	childClip := mergeClip(clip, box.Clip)
	if box.NeedsScrollbar && box.Clip != nil {
		drawVerticalScrollbar(buf, box)
		childClip = narrowClipForVerticalScrollbar(childClip)
	}
	if box.NeedsHorizontalScrollbar && box.Clip != nil {
		drawHorizontalScrollbar(buf, box)
		childClip = narrowClipForHorizontalScrollbar(childClip)
	}
	// Sort children by z-index for paint order (stable sort preserves document order)
	sorted := zSortChildren(box.Children)
	for _, child := range sorted {
		// Fixed children escape parent scroll offsets and clipping
		if child.Position == "fixed" {
			renderTreeWithInherit(buf, child, nil, inherited)
			continue
		}
		renderChildWithScrollInherit(buf, child, box.ScrollX, box.ScrollY, childClip, inherited)
	}
}

// drawVerticalScrollbar draws the vertical scrollbar at the right edge of the clip.
func drawVerticalScrollbar(buf *render.Buffer, box *Box) {
	viewportH := box.Clip.Bottom - box.Clip.Top + 1
	render.DrawScrollbar(buf, box.Clip.Right, box.Clip.Top, viewportH, box.ContentHeight, box.ScrollY, box.Style)
}

// narrowClipForVerticalScrollbar reduces the right edge by 1 to make room for the scrollbar.
func narrowClipForVerticalScrollbar(clip *render.Clip) *render.Clip {
	if clip == nil {
		return nil
	}
	return &render.Clip{
		Top:    clip.Top,
		Left:   clip.Left,
		Bottom: clip.Bottom,
		Right:  clip.Right - 1,
	}
}

// drawHorizontalScrollbar draws the horizontal scrollbar at the bottom edge of the clip.
func drawHorizontalScrollbar(buf *render.Buffer, box *Box) {
	viewportW := box.Clip.Right - box.Clip.Left + 1
	render.DrawHorizontalScrollbar(buf, box.Clip.Left, box.Clip.Bottom, viewportW, box.ContentWidth, box.ScrollX, box.Style)
}

// narrowClipForHorizontalScrollbar reduces the bottom edge by 1 to make room for the scrollbar.
func narrowClipForHorizontalScrollbar(clip *render.Clip) *render.Clip {
	if clip == nil {
		return nil
	}
	return &render.Clip{
		Top:    clip.Top,
		Left:   clip.Left,
		Bottom: clip.Bottom - 1,
		Right:  clip.Right,
	}
}

// renderChildWithScroll renders a child box, translating by the parent's scroll offsets.
// Shifts the entire subtree so all descendants render at the correct scrolled position.
// Sticky children are clamped so they stay visible within the clip region.
func renderChildWithScroll(buf *render.Buffer, child *Box, scrollX, scrollY int, clip *render.Clip) {
	renderChildWithScrollInherit(buf, child, scrollX, scrollY, clip, render.Style{})
}

func renderChildWithScrollInherit(buf *render.Buffer, child *Box, scrollX, scrollY int, clip *render.Clip, inherited render.Style) {
	if scrollX == 0 && scrollY == 0 {
		renderTreeWithInherit(buf, child, clip, inherited)
		return
	}
	shiftTree(child, -scrollX, -scrollY)
	stickyDY := applyStickyClamp(child, clip)
	renderTreeWithInherit(buf, child, clip, inherited)
	shiftTree(child, scrollX, scrollY-stickyDY)
}

// applyStickyClamp adjusts a sticky child's position so it stays visible.
// Returns the extra Y shift applied (to be undone during restore).
func applyStickyClamp(child *Box, clip *render.Clip) int {
	if child.Position != "sticky" || clip == nil {
		return 0
	}
	stickyMinY := clip.Top + child.Top
	if child.Y < stickyMinY {
		dy := stickyMinY - child.Y
		shiftTree(child, 0, dy)
		return dy
	}
	return 0
}

// shiftTree recursively shifts a box and all its descendants by dx, dy.
func shiftTree(box *Box, dx, dy int) {
	box.X += dx
	box.Y += dy
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		shiftTree(child, dx, dy)
	}
}

// zSortChildren returns a copy of children sorted by ZIndex ascending.
// Nil children are filtered out. Stable sort preserves document order for equal z-index.
func zSortChildren(children []*Box) []*Box {
	var sorted []*Box
	for _, c := range children {
		if c != nil {
			sorted = append(sorted, c)
		}
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].ZIndex < sorted[j].ZIndex
	})
	return sorted
}

// mergeClip combines a parent clip with a box's own clip using intersection.
func mergeClip(parent *render.Clip, boxClip *render.Clip) *render.Clip {
	if boxClip == nil {
		return parent
	}
	if parent == nil {
		return boxClip
	}
	return parent.Intersect(boxClip)
}
