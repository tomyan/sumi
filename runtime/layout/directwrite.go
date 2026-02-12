package layout

import (
	"io"

	"github.com/tomyan/sumi/runtime/render"
)

// MapInputToBox walks Input and Box trees in lockstep and returns a map
// from each Input node to its corresponding Box node. Nil children
// (display:none placeholders) are skipped.
func MapInputToBox(input *Input, box *Box) map[*Input]*Box {
	m := make(map[*Input]*Box)
	mapWalk(input, box, m)
	return m
}

func mapWalk(input *Input, box *Box, m map[*Input]*Box) {
	if input == nil || box == nil {
		return
	}
	m[input] = box
	for i := 0; i < len(input.Children) && i < len(box.Children); i++ {
		if box.Children[i] == nil {
			continue
		}
		mapWalk(input.Children[i], box.Children[i], m)
	}
}

// DirectWriteText writes new content at the box's position when the content
// length is unchanged and the box is a single-line (unwrapped) text node.
// Returns true on success, false if conditions aren't met (caller should fall
// back to full Layout+Diff).
func DirectWriteText(w io.Writer, box *Box, newContent, oldContent string) bool {
	if box == nil {
		return false
	}
	if box.Lines != nil {
		return false // wrapped text needs relayout
	}
	if len(newContent) != len(oldContent) {
		return false // dimension change needs relayout
	}

	render.WriteAt(w, box.Y, box.X, newContent, box.Style)
	return true
}
