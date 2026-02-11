package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestRenderBoxFillsBackgroundWhenBGSet(t *testing.T) {
	// Given — a box with a background color and no content
	box := &Box{
		X: 2, Y: 1, Width: 6, Height: 3,
		Style: render.Style{BG: render.Color{Name: "blue"}},
	}

	// When
	buf := render.NewBuffer(20, 10)
	RenderTree(buf, box, nil)

	// Then — all cells in the box area should have the BG style and space char
	for row := 1; row < 4; row++ {
		for col := 2; col < 8; col++ {
			cell := buf.Cell(row, col)
			if cell.Style.BG.Name != "blue" {
				t.Errorf("cell (%d,%d) BG=%q, want 'blue'", row, col, cell.Style.BG.Name)
			}
			if cell.Ch != ' ' {
				t.Errorf("cell (%d,%d) Ch=%q, want space", row, col, cell.Ch)
			}
		}
	}
}

func TestRenderBoxBackgroundBehindContent(t *testing.T) {
	// Given — a box with BG and text content
	box := &Box{
		X: 0, Y: 0, Width: 10, Height: 3,
		Style: render.Style{BG: render.Color{Name: "red"}},
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 5, Height: 1,
				Content: "Hello",
				Style:   render.Style{BG: render.Color{Name: "red"}},
			},
		},
	}

	// When
	buf := render.NewBuffer(10, 3)
	RenderTree(buf, box, nil)

	// Then — text cells have content, empty cells still have BG
	cell := buf.Cell(0, 0)
	if cell.Ch != 'H' {
		t.Errorf("expected 'H' at (0,0), got %q", cell.Ch)
	}
	if cell.Style.BG.Name != "red" {
		t.Errorf("text cell BG=%q, want 'red'", cell.Style.BG.Name)
	}
	// Empty area below text should have BG fill
	emptyCell := buf.Cell(1, 5)
	if emptyCell.Style.BG.Name != "red" {
		t.Errorf("empty cell (1,5) BG=%q, want 'red'", emptyCell.Style.BG.Name)
	}
}

func TestRenderBoxNoFillWithoutBG(t *testing.T) {
	// Given — a box with no background color
	box := &Box{
		X: 0, Y: 0, Width: 5, Height: 2,
	}

	// When
	buf := render.NewBuffer(10, 5)
	RenderTree(buf, box, nil)

	// Then — cells should be untouched (zero rune, zero style)
	cell := buf.Cell(0, 0)
	if cell.Ch != 0 {
		t.Errorf("expected zero char, got %q", cell.Ch)
	}
	if cell.Style.BG.Name != "" {
		t.Errorf("expected no BG, got %q", cell.Style.BG.Name)
	}
}

func TestRenderBoxBackgroundWithBorder(t *testing.T) {
	// Given — a bordered box with BG fills interior only
	box := &Box{
		X: 0, Y: 0, Width: 6, Height: 4,
		Border: "single",
		Style:  render.Style{BG: render.Color{Name: "green"}},
	}

	// When
	buf := render.NewBuffer(10, 5)
	RenderTree(buf, box, nil)

	// Then — interior cells (1,1) to (2,4) have BG fill
	interior := buf.Cell(1, 1)
	if interior.Style.BG.Name != "green" {
		t.Errorf("interior cell (1,1) BG=%q, want 'green'", interior.Style.BG.Name)
	}
	if interior.Ch != ' ' {
		t.Errorf("interior cell (1,1) Ch=%q, want space", interior.Ch)
	}
}
