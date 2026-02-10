package layout

import (
	"testing"
)

func TestLayoutAlignStretchIsDefault(t *testing.T) {
	// Given a row with children of different heights, default align = stretch (CSS flexbox default)
	input := &Input{
		Kind:        KindBox,
		Direction:   "row",
		FixedWidth:  40,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 3},
			{Kind: KindBox, FixedWidth: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children stretch to fill cross axis (height=10)
	// Child with FixedHeight keeps its height (explicit size overrides stretch)
	if box.Children[0].Height != 3 {
		t.Errorf("child[0].Height = %d, want 3 (fixed height overrides stretch)", box.Children[0].Height)
	}
	if box.Children[1].Height != 10 {
		t.Errorf("child[1].Height = %d, want 10 (stretched)", box.Children[1].Height)
	}
}

func TestLayoutAlignStretchDefaultColumn(t *testing.T) {
	// Given a column (default), children should stretch to fill width
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedWidth:  40,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindBox, FixedHeight: 3},
			{Kind: KindBox, FixedHeight: 3, FixedWidth: 10},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then child without fixed width stretches, child with fixed width keeps it
	if box.Children[0].Width != 40 {
		t.Errorf("child[0].Width = %d, want 40 (stretched)", box.Children[0].Width)
	}
	if box.Children[1].Width != 10 {
		t.Errorf("child[1].Width = %d, want 10 (fixed width overrides stretch)", box.Children[1].Width)
	}
}

func TestLayoutAlignStartExplicit(t *testing.T) {
	// Given explicit align=start, children should NOT stretch
	input := &Input{
		Kind:        KindBox,
		Direction:   "row",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "start",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 3},
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children keep their natural height
	if box.Children[0].Height != 3 {
		t.Errorf("child[0].Height = %d, want 3", box.Children[0].Height)
	}
	if box.Children[1].Height != 5 {
		t.Errorf("child[1].Height = %d, want 5", box.Children[1].Height)
	}
}

func TestLayoutAlignEndRow(t *testing.T) {
	// Given a row with align=end, children aligned to bottom
	input := &Input{
		Kind:        KindBox,
		Direction:   "row",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "end",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 3},
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are bottom-aligned: 10 - height
	if box.Children[0].Y != 7 {
		t.Errorf("child[0].Y = %d, want 7 (10-3)", box.Children[0].Y)
	}
	if box.Children[1].Y != 5 {
		t.Errorf("child[1].Y = %d, want 5 (10-5)", box.Children[1].Y)
	}
}

func TestLayoutAlignCenterRow(t *testing.T) {
	// Given a row with align=center, children centered vertically
	input := &Input{
		Kind:        KindBox,
		Direction:   "row",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "center",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 2},
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 4},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are centered: (10-height)/2
	if box.Children[0].Y != 4 {
		t.Errorf("child[0].Y = %d, want 4 ((10-2)/2)", box.Children[0].Y)
	}
	if box.Children[1].Y != 3 {
		t.Errorf("child[1].Y = %d, want 3 ((10-4)/2)", box.Children[1].Y)
	}
}

func TestLayoutAlignStretchRow(t *testing.T) {
	// Given a row with align=stretch, children expand to fill cross axis
	input := &Input{
		Kind:        KindBox,
		Direction:   "row",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "stretch",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5},
			{Kind: KindBox, FixedWidth: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children stretch to fill height
	if box.Children[0].Height != 10 {
		t.Errorf("child[0].Height = %d, want 10 (stretched)", box.Children[0].Height)
	}
	if box.Children[1].Height != 10 {
		t.Errorf("child[1].Height = %d, want 10 (stretched)", box.Children[1].Height)
	}
}

func TestLayoutAlignEndColumn(t *testing.T) {
	// Given a column with align=end, children right-aligned
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "end",
		Children: []*Input{
			{Kind: KindText, Content: "short"},
			{Kind: KindText, Content: "longer text"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are right-aligned: 40 - width
	if box.Children[0].X != 35 {
		t.Errorf("child[0].X = %d, want 35 (40-5)", box.Children[0].X)
	}
	if box.Children[1].X != 29 {
		t.Errorf("child[1].X = %d, want 29 (40-11)", box.Children[1].X)
	}
}

func TestLayoutAlignCenterColumn(t *testing.T) {
	// Given a column with align=center, children centered horizontally
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "center",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then child is centered: (40-5)/2 = 17
	if box.Children[0].X != 17 {
		t.Errorf("child[0].X = %d, want 17 ((40-5)/2)", box.Children[0].X)
	}
}

func TestLayoutAlignStretchColumn(t *testing.T) {
	// Given a column with align=stretch, children expand to fill width
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedWidth:  40,
		FixedHeight: 10,
		Align:       "stretch",
		Children: []*Input{
			{Kind: KindBox, FixedHeight: 3},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then child stretches to fill width
	if box.Children[0].Width != 40 {
		t.Errorf("child[0].Width = %d, want 40 (stretched)", box.Children[0].Width)
	}
}

func TestLayoutAlignWithPaddingAndBorder(t *testing.T) {
	// Given align=center with padding and border
	input := &Input{
		Kind:        KindBox,
		Direction:   "row",
		FixedWidth:  40,
		FixedHeight: 10,
		Padding:     Padding{1, 1, 1, 1},
		Border:      "single",
		Align:       "center",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 2},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then: content area height = 10 - 2(border) - 2(pad) = 6
	// center = offsetY + (6-2)/2 = 2 + 2 = 4
	if box.Children[0].Y != 4 {
		t.Errorf("child Y = %d, want 4 (centered in content area)", box.Children[0].Y)
	}
}
