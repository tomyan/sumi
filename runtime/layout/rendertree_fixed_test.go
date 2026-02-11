package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// bufRowString extracts a string from a buffer row starting at col 0.
func bufRowString(buf *render.Buffer, row, length int) string {
	var runes []rune
	for col := 0; col < length; col++ {
		ch := buf.Cell(row, col).Ch
		if ch == 0 {
			ch = ' '
		}
		runes = append(runes, ch)
	}
	return string(runes)
}

func TestRenderFixedEscapesParentClip(t *testing.T) {
	// Given — parent has overflow:hidden clipping at rows 2-5,
	// fixed child at viewport row 0 should still be visible
	parent := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Clip: &render.Clip{Top: 2, Left: 0, Bottom: 5, Right: 19},
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 10, Height: 1,
				Content:  "fixed!",
				Position: "fixed",
			},
		},
	}

	// When
	buf := render.NewBuffer(20, 10)
	RenderTree(buf, parent, nil)

	// Then — fixed child rendered at row 0, not clipped by parent
	row := bufRowString(buf, 0, 6)
	if row != "fixed!" {
		t.Errorf("expected 'fixed!' at row 0, got %q", row)
	}
}

func TestRenderFixedIgnoresParentScroll(t *testing.T) {
	// Given — parent has scroll offset, fixed child should not be shifted
	parent := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		ScrollY:       5,
		ContentHeight: 100,
		Children: []*Box{
			{
				X: 0, Y: 2, Width: 12, Height: 1,
				Content: "scroll-child",
			},
			{
				X: 0, Y: 1, Width: 6, Height: 1,
				Content:  "fixed!",
				Position: "fixed",
			},
		},
	}

	// When
	buf := render.NewBuffer(20, 10)
	RenderTree(buf, parent, nil)

	// Then — fixed child stays at Y=1 (not shifted by ScrollY=5)
	row := bufRowString(buf, 1, 6)
	if row != "fixed!" {
		t.Errorf("expected 'fixed!' at row 1, got %q", row)
	}
}
