package layout

import "github.com/tomyan/sumi/runtime/render"

// computeClip returns the clip region for a box with overflow set.
// The clip covers the content area inside border and padding.
func computeClip(box *Box, borderWidth int, pad Padding) *render.Clip {
	top := box.Y + borderWidth + pad.Top
	left := box.X + borderWidth + pad.Left
	bottom := box.Y + box.Height - 1 - borderWidth - pad.Bottom
	right := box.X + box.Width - 1 - borderWidth - pad.Right
	return &render.Clip{Top: top, Left: left, Bottom: bottom, Right: right}
}
