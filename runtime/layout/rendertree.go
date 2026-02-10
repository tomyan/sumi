package layout

import "github.com/tomyan/sumi/runtime/render"

// RenderTree renders a layout tree to a buffer, applying clip regions and scroll offsets.
func RenderTree(buf *render.Buffer, box *Box, clip *render.Clip) {
	renderBorder(buf, box)
	renderContent(buf, box, clip)
	renderScrollbarsAndChildren(buf, box, clip)
}

func renderBorder(buf *render.Buffer, box *Box) {
	if box.Border != "" && box.Border != "none" {
		buf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)
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
	for _, child := range box.Children {
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
func renderChildWithScroll(buf *render.Buffer, child *Box, scrollX, scrollY int, clip *render.Clip) {
	if scrollX == 0 && scrollY == 0 {
		RenderTree(buf, child, clip)
		return
	}
	shiftTree(child, -scrollX, -scrollY)
	RenderTree(buf, child, clip)
	shiftTree(child, scrollX, scrollY) // restore
}

// shiftTree recursively shifts a box and all its descendants by dx, dy.
func shiftTree(box *Box, dx, dy int) {
	box.X += dx
	box.Y += dy
	for _, child := range box.Children {
		shiftTree(child, dx, dy)
	}
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
