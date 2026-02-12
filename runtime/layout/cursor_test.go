package layout

import "testing"

func TestFindCursorReturnsBoxWithCursor(t *testing.T) {
	// Given — a tree with one child that has a cursor set
	tree := &Box{
		X: 0, Y: 0, Width: 40, Height: 10,
		CursorCol: -1, CursorRow: -1,
		Children: []*Box{
			{X: 1, Y: 1, Width: 10, Height: 1, Content: "hello", CursorCol: 3, CursorRow: 0},
		},
	}

	// When
	found := FindCursor(tree)

	// Then
	if found == nil {
		t.Fatal("expected to find a cursor box")
	}
	if found.CursorCol != 3 {
		t.Errorf("CursorCol = %d, want 3", found.CursorCol)
	}
}

func TestFindCursorReturnsNilWhenNoCursor(t *testing.T) {
	// Given — a tree with no cursor (all CursorCol = -1)
	tree := &Box{
		X: 0, Y: 0, Width: 40, Height: 10,
		CursorCol: -1, CursorRow: -1,
		Children: []*Box{
			{X: 1, Y: 1, Width: 10, Height: 1, Content: "hello", CursorCol: -1, CursorRow: -1},
		},
	}

	// When
	found := FindCursor(tree)

	// Then
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestFindCursorComputesAbsolutePosition(t *testing.T) {
	// Given — cursor at column 2 inside a box at X=5, Y=3
	tree := &Box{
		X: 0, Y: 0, Width: 40, Height: 10,
		CursorCol: -1, CursorRow: -1,
		Children: []*Box{
			{X: 5, Y: 3, Width: 10, Height: 1, Content: "hello", CursorCol: 2, CursorRow: 0},
		},
	}

	// When
	found := FindCursor(tree)

	// Then — cursor's absolute position is box.X + CursorCol, box.Y + CursorRow
	if found == nil {
		t.Fatal("expected to find a cursor box")
	}
	absCol := found.X + found.CursorCol
	absRow := found.Y + found.CursorRow
	if absCol != 7 {
		t.Errorf("absolute cursor col = %d, want 7", absCol)
	}
	if absRow != 3 {
		t.Errorf("absolute cursor row = %d, want 3", absRow)
	}
}

func TestFindCursorAtOriginIsValid(t *testing.T) {
	// Given — cursor at position (0,0) is a valid cursor position
	tree := &Box{
		X: 0, Y: 0, Width: 40, Height: 10,
		CursorCol: 0, CursorRow: 0,
	}

	// When
	found := FindCursor(tree)

	// Then
	if found == nil {
		t.Fatal("expected to find a cursor at origin")
	}
	if found.CursorCol != 0 || found.CursorRow != 0 {
		t.Errorf("expected cursor at (0,0), got (%d,%d)", found.CursorCol, found.CursorRow)
	}
}
