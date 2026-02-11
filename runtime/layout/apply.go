package layout

import (
	"io"

	"github.com/tomyan/sumi/runtime/render"
)

// ApplyChanges writes only the changed nodes to w using direct cursor positioning.
// For each change: clears the old region (if any), then renders the new node.
func ApplyChanges(w io.Writer, changes []Change) {
	for _, c := range changes {
		if c.Old != nil && c.New == nil {
			// Removal — clear old region
			render.ClearRegion(w, c.Old.Y, c.Old.X, c.Old.Width, c.Old.Height)
			continue
		}
		if c.Old != nil {
			// Clear old region before rendering new content
			render.ClearRegion(w, c.Old.Y, c.Old.X, c.Old.Width, c.Old.Height)
		}
		if c.New != nil {
			renderNodeDirect(w, c.New)
		}
	}
}

// renderNodeDirect renders a single node directly to a writer.
func renderNodeDirect(w io.Writer, box *Box) {
	if box.Border != "" && box.Border != "none" {
		if !box.Collapsed.IsZero() {
			render.DrawCollapsedBorderAt(w, box.Y, box.X, box.Width, box.Height, box.Border, box.Style, box.Collapsed)
		} else {
			render.DrawBorderAt(w, box.Y, box.X, box.Width, box.Height, box.Border, box.Style)
		}
		if box.BorderTitle != "" {
			render.DrawBorderTitleAt(w, box.Y, box.X, box.Width, box.BorderTitle, box.Style)
		}
	}
	if box.Lines != nil {
		for i, line := range box.Lines {
			render.WriteAt(w, box.Y+i, box.X, line, box.Style)
		}
	} else if box.Content != "" {
		render.WriteAt(w, box.Y, box.X, box.Content, box.Style)
	}
}
