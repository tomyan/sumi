package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestRenderTreeDrawsBorderTitle(t *testing.T) {
	// Given — a box with border and title
	box := &Box{
		X: 0, Y: 0, Width: 20, Height: 5,
		Border:      "single",
		BorderTitle: "Panel",
	}

	// When
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then — title should appear in the top edge
	// Pattern: ┌─ Panel ──────────┐
	expected := " Panel "
	for i, ch := range expected {
		if c := buf.Cell(0, 2+i); c.Ch != ch {
			t.Errorf("Cell(0,%d) = %c, want %c", 2+i, c.Ch, ch)
		}
	}
}
