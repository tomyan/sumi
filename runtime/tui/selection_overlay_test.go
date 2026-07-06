package tui

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// D5b: selection overlay — toggles inverse over the ribbon.

func TestOverlayInvertsRibbon(t *testing.T) {
	// Given: 3 rows, selection (2,0)..(1,2).
	buf := selBuffer("aaaa", "bbbb", "cccc")

	// When
	ApplySelectionOverlay(buf, &SelectionRange{Start: CellPos{Col: 2, Row: 0}, End: CellPos{Col: 1, Row: 2}})

	// Then: row 0 from col 2; row 1 whole; row 2 through col 1.
	cases := []struct {
		row, col int
		want     bool
	}{
		{0, 1, false}, {0, 2, true}, {0, 3, true},
		{1, 0, true}, {1, 3, true},
		{2, 0, true}, {2, 1, true}, {2, 2, false},
	}
	for _, c := range cases {
		if got := buf.Cell(c.row, c.col).Style.Inverse; got != c.want {
			t.Errorf("cell (%d,%d) inverse = %v, want %v", c.row, c.col, got, c.want)
		}
	}
}

func TestOverlayTogglesAlreadyInverseCells(t *testing.T) {
	// Given: a cell already inverse (e.g. kbd styling).
	buf := selBuffer("ab")
	buf.SetStyledCell(0, 1, 'b', render.Style{Inverse: true})

	// When
	ApplySelectionOverlay(buf, &SelectionRange{Start: CellPos{Col: 0, Row: 0}, End: CellPos{Col: 1, Row: 0}})

	// Then: XOR — the inverse cell renders normal inside the selection.
	if !buf.Cell(0, 0).Style.Inverse {
		t.Error("plain cell should invert")
	}
	if buf.Cell(0, 1).Style.Inverse {
		t.Error("inverse cell should toggle back to normal")
	}
}

func TestOverlayNilRangeIsNoop(t *testing.T) {
	// Given
	buf := selBuffer("ab")

	// When
	ApplySelectionOverlay(buf, nil)

	// Then
	if buf.Cell(0, 0).Style.Inverse {
		t.Error("no range must not paint")
	}
}
