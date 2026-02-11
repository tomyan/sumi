package layout

import (
	"sort"

	"github.com/tomyan/sumi/runtime/render"
)

// RenderTree renders a layout tree to a buffer, applying clip regions and scroll offsets.
func RenderTree(buf *render.Buffer, box *Box, clip *render.Clip) {
	renderBackground(buf, box, clip)
	renderBorder(buf, box)
	renderContent(buf, box, clip)
	renderScrollbarsAndChildren(buf, box, clip)
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
}

// renderScrollbarsAndChildren draws scrollbars (if needed) then renders children
// with the content clip narrowed to avoid overlap with scrollbars.
func renderScrollbarsAndChildren(buf *render.Buffer, box *Box, clip *render.Clip) {
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
			RenderTree(buf, child, nil)
			continue
		}
		renderChildWithScroll(buf, child, box.ScrollX, box.ScrollY, childClip)
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
	if scrollX == 0 && scrollY == 0 {
		RenderTree(buf, child, clip)
		return
	}
	shiftTree(child, -scrollX, -scrollY)
	stickyDY := applyStickyClamp(child, clip)
	RenderTree(buf, child, clip)
	shiftTree(child, scrollX, scrollY-stickyDY) // restore (undo scroll + clamp)
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
