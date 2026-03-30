package layout

import (
	"sort"

	"github.com/tomyan/sumi/runtime/render"
)

// RenderTree renders a layout tree to a buffer, applying clip regions and scroll offsets.
func RenderTree(buf *render.Buffer, box *Box, clip *render.Clip) {
	renderTreeWithInherit(buf, box, clip, render.Style{})
}

func renderTreeWithInherit(buf *render.Buffer, box *Box, clip *render.Clip, inherited render.Style) {
	// Apply hover: when hovered and a hover style exists, use it instead.
	if box.Hovered && !box.HoverStyle.IsZero() {
		box.Style = box.HoverStyle
	}

	// Apply style inheritance: merge parent's inheritable properties into this node.
	box.Style = box.Style.Inherit(inherited)

	renderBackground(buf, box, clip)
	renderBorder(buf, box)
	renderContent(buf, box, clip)
	renderScrollbarsAndChildrenWithInherit(buf, box, clip, box.Style)
}

// renderBackground fills the box area with spaces using the box's BG color.
// Only fills when a background color is set. Fills inside border if present.
func renderBackground(buf *render.Buffer, box *Box, clip *render.Clip) {
	if box.Style.BG.Name == "" {
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
	if box.Lines != nil {
		for i, line := range box.Lines {
			buf.WriteStyledTextClipped(box.Y+i, box.X, line, box.Style, clip)
		}
	} else if box.Content != "" {
		buf.WriteStyledTextClipped(box.Y, box.X, box.Content, box.Style, clip)
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

// renderScrollbarsAndChildren draws scrollbars (if needed) then renders children
// with the content clip narrowed to avoid overlap with scrollbars.
func renderScrollbarsAndChildren(buf *render.Buffer, box *Box, clip *render.Clip) {
	renderScrollbarsAndChildrenWithInherit(buf, box, clip, render.Style{})
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
