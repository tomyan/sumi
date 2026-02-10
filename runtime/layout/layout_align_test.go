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

func TestLayoutAlignStretchRowAutoHeight(t *testing.T) {
	// Given a row with NO fixed height and children of different heights,
	// stretch should make all children match the tallest child's height,
	// NOT the parent's available height (terminal height).
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 10, FixedHeight: 3},
			{Kind: KindBox, FixedWidth: 10},
		},
	}

	// When laid out in an 80x24 terminal
	box := Layout(input, 80, 24)

	// Then the stretchable child should match the tallest sibling (3),
	// not the terminal height (24)
	if box.Children[0].Height != 3 {
		t.Errorf("child[0].Height = %d, want 3 (fixed height)", box.Children[0].Height)
	}
	if box.Children[1].Height != 3 {
		t.Errorf("child[1].Height = %d, want 3 (stretched to tallest sibling, not terminal height)", box.Children[1].Height)
	}
	// The row container itself should be 3 tall, not 24
	if box.Height != 3 {
		t.Errorf("row.Height = %d, want 3 (auto-sized to tallest child)", box.Height)
	}
}

func TestLayoutAlignStretchRowAutoHeightWithFlexGrow(t *testing.T) {
	// Given a row with flex-grow children but NO fixed height,
	// stretch should use tallest child height, not terminal height.
	// This is the flexbox dashboard bug.
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Gap:       1,
		Children: []*Input{
			{
				Kind:     KindBox,
				FlexGrow: 1,
				Border:   "single",
				Padding:  Padding{0, 1, 0, 1},
				Children: []*Input{
					{Kind: KindText, Content: "Left Panel"},
					{Kind: KindText, Content: "This panel uses flex-grow"},
					{Kind: KindText, Content: "to fill available space."},
				},
			},
			{
				Kind:     KindBox,
				FlexGrow: 1,
				Border:   "single",
				Padding:  Padding{0, 1, 0, 1},
				Children: []*Input{
					{Kind: KindText, Content: "Right Panel"},
					{Kind: KindText, Content: "Both panels share the"},
					{Kind: KindText, Content: "width equally."},
				},
			},
		},
	}

	// When laid out in 80x50 terminal
	box := Layout(input, 80, 50)

	// Then both panels should have matching height based on content,
	// not stretch to 50 (terminal height)
	// Each panel: 3 text lines + 2 border = 5 rows
	leftH := box.Children[0].Height
	rightH := box.Children[1].Height
	if leftH != rightH {
		t.Errorf("panels have different heights: left=%d, right=%d", leftH, rightH)
	}
	if leftH > 10 {
		t.Errorf("left panel height = %d, want <=10 (content-sized, not terminal-sized)", leftH)
	}
	if box.Height > 10 {
		t.Errorf("row height = %d, want <=10 (auto-sized to content)", box.Height)
	}
}

func TestLayoutAlignStretchRowAutoHeightMixedChildren(t *testing.T) {
	// Given a row with children of varying natural heights and no fixed height,
	// stretch should make all stretchable children match the tallest.
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{
				Kind:       KindBox,
				FixedWidth: 10,
				Children: []*Input{
					{Kind: KindText, Content: "Line 1"},
				},
			},
			{
				Kind:       KindBox,
				FixedWidth: 10,
				Children: []*Input{
					{Kind: KindText, Content: "Line 1"},
					{Kind: KindText, Content: "Line 2"},
					{Kind: KindText, Content: "Line 3"},
					{Kind: KindText, Content: "Line 4"},
					{Kind: KindText, Content: "Line 5"},
				},
			},
			{
				Kind:       KindBox,
				FixedWidth: 10,
				Children: []*Input{
					{Kind: KindText, Content: "Line 1"},
					{Kind: KindText, Content: "Line 2"},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 40)

	// Then all children stretch to the tallest (5), not terminal height (40)
	for i, child := range box.Children {
		if child.Height != 5 {
			t.Errorf("child[%d].Height = %d, want 5 (tallest sibling)", i, child.Height)
		}
	}
	if box.Height != 5 {
		t.Errorf("row.Height = %d, want 5", box.Height)
	}
}
