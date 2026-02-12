package layout

import "testing"

func TestTextNodesHaveNoCursor(t *testing.T) {
	// Given — a tree with a text node and a box with an explicit cursor.
	// The root box sets CursorCol=-1 (as codegen does); the text node has
	// Go zero value CursorCol=0, which should NOT be treated as a cursor.
	input := &Input{
		Kind: KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindBox, CursorCol: 3, CursorRow: 0,
				Children: []*Input{{Kind: KindText, Content: "field"}}},
		},
	}

	// When
	tree := Layout(input, 40, 10)

	// Then — FindCursor should find the box with explicit cursor, not the text node
	found := FindCursor(tree)
	if found == nil {
		t.Fatal("expected to find cursor box")
	}
	if found.CursorCol != 3 {
		t.Errorf("CursorCol = %d, want 3 (found text node instead of cursor box)", found.CursorCol)
	}
}

func TestTextNodeDoesNotMatchFindCursor(t *testing.T) {
	// Given — only text nodes, no explicit cursor
	input := &Input{
		Kind:      KindBox,
		CursorCol: -1,
		CursorRow: -1,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	tree := Layout(input, 40, 10)

	// Then — no cursor should be found
	found := FindCursor(tree)
	if found != nil {
		t.Errorf("expected no cursor, but found box with CursorCol=%d CursorRow=%d content=%q",
			found.CursorCol, found.CursorRow, found.Content)
	}
}
