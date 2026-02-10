package layout

import "testing"

func TestCollapseColumnTwoChildren(t *testing.T) {
	// Given — column container with border-collapse and two bordered children
	input := &Input{
		Kind:           KindBox,
		BorderCollapse: true,
		FixedWidth:     20,
		FixedHeight:    10,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FlexGrow: 1},
			{Kind: KindBox, Border: "single", FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 20, 10)

	// Then — child[1].Y should equal child[0].Y + child[0].Height - 1 (overlap by 1)
	c0 := box.Children[0]
	c1 := box.Children[1]
	wantY := c0.Y + c0.Height - 1
	if c1.Y != wantY {
		t.Errorf("child[1].Y = %d, want %d (child[0].Y=%d + Height=%d - 1)", c1.Y, wantY, c0.Y, c0.Height)
	}
}

func TestCollapseColumnParentSize(t *testing.T) {
	// Given — two bordered children, each 5 tall
	input := &Input{
		Kind:           KindBox,
		BorderCollapse: true,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FixedHeight: 5, FixedWidth: 10},
			{Kind: KindBox, Border: "single", FixedHeight: 5, FixedWidth: 10},
		},
	}

	// When
	box := Layout(input, 20, 20)

	// Then — parent height = 5 + 5 - 1 (one overlap) = 9
	if box.Height != 9 {
		t.Errorf("parent Height = %d, want 9", box.Height)
	}
}

func TestCollapseColumnEdgeFlags(t *testing.T) {
	// Given
	input := &Input{
		Kind:           KindBox,
		BorderCollapse: true,
		FixedWidth:     20,
		FixedHeight:    10,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FlexGrow: 1},
			{Kind: KindBox, Border: "single", FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 20, 10)

	// Then — child[0] bottom collapsed, child[1] top collapsed
	c0 := box.Children[0]
	c1 := box.Children[1]
	if !c0.Collapsed.Bottom {
		t.Error("child[0].Collapsed.Bottom should be true")
	}
	if !c1.Collapsed.Top {
		t.Error("child[1].Collapsed.Top should be true")
	}
	// Top of first child and bottom of last child should NOT be collapsed
	if c0.Collapsed.Top {
		t.Error("child[0].Collapsed.Top should be false")
	}
	if c1.Collapsed.Bottom {
		t.Error("child[1].Collapsed.Bottom should be false")
	}
}

func TestCollapseColumnNoParentBorderInset(t *testing.T) {
	// Given — parent with border + collapse: children should use the full parent area (inset=0)
	input := &Input{
		Kind:           KindBox,
		Border:         "single",
		BorderCollapse: true,
		FixedWidth:     20,
		FixedHeight:    10,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FlexGrow: 1},
			{Kind: KindBox, Border: "single", FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 20, 10)

	// Then — children start at position 0 (no border inset)
	c0 := box.Children[0]
	if c0.X != 0 || c0.Y != 0 {
		t.Errorf("child[0] position = (%d,%d), want (0,0)", c0.X, c0.Y)
	}
	// Children should span the full width
	if c0.Width != 20 {
		t.Errorf("child[0].Width = %d, want 20", c0.Width)
	}
}
