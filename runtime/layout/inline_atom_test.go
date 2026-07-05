package layout

import "testing"

// B4d: inline-block atoms — boxes join the IFC as unbreakable units,
// top-aligned on the line; line height grows to the tallest item.

func TestInlineBlockAtomFlowsWithText(t *testing.T) {
	// Given: text, a 4x2 inline-block, text.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "ab "},
		{Kind: KindBox, Display: "inline-block", FixedWidth: 4, FixedHeight: 2},
		{Kind: KindText, Content: " cd"},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then: one line — atom at x3, tail text at x8.
	atom := box.Children[1]
	if atom.X != 3 || atom.Y != 0 {
		t.Errorf("atom at (%d,%d), want (3,0)", atom.X, atom.Y)
	}
	if atom.Width != 4 || atom.Height != 2 {
		t.Errorf("atom size = %dx%d, want 4x2", atom.Width, atom.Height)
	}
	// The tail box starts at its leading space fragment " cd".
	tail := box.Children[2]
	if tail.X != 7 || tail.Y != 0 {
		t.Errorf("tail at (%d,%d), want (7,0)", tail.X, tail.Y)
	}
	// Line height is the atom's 2 rows.
	if box.Height != 2 {
		t.Errorf("container height = %d, want 2", box.Height)
	}
}

func TestInlineBlockAtomWrapsToNextLine(t *testing.T) {
	// Given: the atom does not fit after the text at width 6.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "abcd "},
		{Kind: KindBox, Display: "inline-block", FixedWidth: 4, FixedHeight: 1},
	}}

	// When
	box := Layout(p, 6, 24)

	// Then
	atom := box.Children[1]
	if atom.X != 0 || atom.Y != 1 {
		t.Errorf("atom at (%d,%d), want (0,1)", atom.X, atom.Y)
	}
}

func TestTallAtomPushesNextLineDown(t *testing.T) {
	// Given: a 3-row atom on line 0, then a wrap: line 1 starts at y3.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindBox, Display: "inline-block", FixedWidth: 4, FixedHeight: 3},
		{Kind: KindText, Content: " aa bb"},
	}}

	// When: width 8 fits atom+" aa", wraps "bb" to line 1.
	box := Layout(p, 8, 24)

	// Then
	text := box.Children[1]
	if len(text.Fragments) != 2 {
		t.Fatalf("fragments = %+v, want 2", text.Fragments)
	}
	// Second fragment starts below the 3-row first line.
	if got := text.Y + text.Fragments[1].Y; got != 3 {
		t.Errorf("second line Y = %d, want 3", got)
	}
	if box.Height != 4 {
		t.Errorf("container height = %d, want 4", box.Height)
	}
}
