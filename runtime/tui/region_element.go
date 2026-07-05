package tui

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/layout"
)

// resyncRegionElement dispatches a "resize" DOMEvent on a region whose
// laid-out content size changed since the last dispatch. The consumer's
// handler feeds evt.Target.Cells with content for the new size. Returns
// true when a resize fired (requesting another converge pass to paint
// the freshly fed cells).
func resyncRegionElement(comp *Component, n *layout.Input) bool {
	width, height := regionContentSize(n)
	if width <= 0 || height <= 0 {
		return false
	}
	size := fmt.Sprintf("%dx%d", width, height)
	if n.Attrs == nil {
		n.Attrs = map[string]string{}
	}
	if n.Attrs["sumi:region-size"] == size {
		return false
	}
	n.Attrs["sumi:region-size"] = size
	path := layout.PathTo(comp.Tree, n)
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "resize",
		Data: map[string]any{"width": width, "height": height}})
	return true
}

// regionContentSize is the area inside the region's border and padding.
func regionContentSize(n *layout.Input) (int, int) {
	w, h := n.LastW, n.LastH
	if w <= 0 {
		w = n.FixedWidth
	}
	if h <= 0 {
		h = n.FixedHeight
	}
	if n.Border != "" && n.Border != "none" {
		w -= 2
		h -= 2
	}
	return w - n.Padding.Left - n.Padding.Right, h - n.Padding.Top - n.Padding.Bottom
}
