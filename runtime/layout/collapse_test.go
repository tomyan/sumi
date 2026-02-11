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

func TestCollapseRowTwoChildren(t *testing.T) {
	// Given — row container with border-collapse and two bordered children
	input := &Input{
		Kind:           KindBox,
		Direction:      "row",
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

	// Then — child[1].X should equal child[0].X + child[0].Width - 1 (overlap by 1)
	c0 := box.Children[0]
	c1 := box.Children[1]
	wantX := c0.X + c0.Width - 1
	if c1.X != wantX {
		t.Errorf("child[1].X = %d, want %d (child[0].X=%d + Width=%d - 1)", c1.X, wantX, c0.X, c0.Width)
	}
}

func TestCollapseRowParentSize(t *testing.T) {
	// Given — two bordered children, each 10 wide
	input := &Input{
		Kind:           KindBox,
		Direction:      "row",
		BorderCollapse: true,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FixedWidth: 10, FixedHeight: 5},
			{Kind: KindBox, Border: "single", FixedWidth: 10, FixedHeight: 5},
		},
	}

	// When
	box := Layout(input, 40, 20)

	// Then — parent width = 10 + 10 - 1 (one overlap) = 19
	if box.Width != 19 {
		t.Errorf("parent Width = %d, want 19", box.Width)
	}
}

func TestCollapseRowEdgeFlags(t *testing.T) {
	// Given
	input := &Input{
		Kind:           KindBox,
		Direction:      "row",
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

	// Then — child[0] right collapsed, child[1] left collapsed
	c0 := box.Children[0]
	c1 := box.Children[1]
	if !c0.Collapsed.Right {
		t.Error("child[0].Collapsed.Right should be true")
	}
	if !c1.Collapsed.Left {
		t.Error("child[1].Collapsed.Left should be true")
	}
	// Left of first child and right of last child should NOT be collapsed
	if c0.Collapsed.Left {
		t.Error("child[0].Collapsed.Left should be false")
	}
	if c1.Collapsed.Right {
		t.Error("child[1].Collapsed.Right should be false")
	}
}

func TestCollapseNestedLayout(t *testing.T) {
	// Given — tmux-style 3-panel layout:
	// Row container (collapse) with:
	//   - Left column (collapse) with Panel 1 + Panel 2
	//   - Right Panel 3
	input := &Input{
		Kind:           KindBox,
		Direction:      "row",
		BorderCollapse: true,
		FixedWidth:     40,
		FixedHeight:    12,
		Children: []*Input{
			{
				Kind:           KindBox,
				BorderCollapse: true,
				Border:         "single",
				FlexGrow:       1,
				Children: []*Input{
					{Kind: KindBox, Border: "single", FlexGrow: 1},
					{Kind: KindBox, Border: "single", FlexGrow: 1},
				},
			},
			{Kind: KindBox, Border: "single", FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 40, 12)

	// Then — verify structure
	leftCol := box.Children[0]
	rightPanel := box.Children[1]

	// Row collapse: left column and right panel share a vertical border
	if !leftCol.Collapsed.Right {
		t.Error("left column Collapsed.Right should be true")
	}
	if !rightPanel.Collapsed.Left {
		t.Error("right panel Collapsed.Left should be true")
	}

	// Column collapse within left column: Panel 1 and Panel 2 share a horizontal border
	panel1 := leftCol.Children[0]
	panel2 := leftCol.Children[1]
	if !panel1.Collapsed.Bottom {
		t.Error("panel1 Collapsed.Bottom should be true")
	}
	if !panel2.Collapsed.Top {
		t.Error("panel2 Collapsed.Top should be true")
	}

	// Right panel should span the full height of the parent
	if rightPanel.Height != 12 {
		t.Errorf("right panel Height = %d, want 12", rightPanel.Height)
	}

	// Left column + right panel should share border (overlap by 1)
	wantX := leftCol.X + leftCol.Width - 1
	if rightPanel.X != wantX {
		t.Errorf("right panel X = %d, want %d", rightPanel.X, wantX)
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
