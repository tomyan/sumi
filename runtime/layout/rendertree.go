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
	for _, child := range box.Children {
		renderChildWithScroll(buf, child, box.ScrollY, childClip)
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

// renderChildWithScroll renders a child box, translating by the parent's scroll offset.
func renderChildWithScroll(buf *render.Buffer, child *Box, scrollY int, clip *render.Clip) {
	if scrollY == 0 {
		RenderTree(buf, child, clip)
		return
	}
	// Translate child position by -scrollY for rendering
	child.Y -= scrollY
	RenderTree(buf, child, clip)
	child.Y += scrollY // restore
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
