package layout

import (
	"testing"
)

func TestLayoutFlexGrowOneChildFillsRow(t *testing.T) {
	// Given a fixed-width row with one flex-grow child
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Children: []*Input{
			{Kind: KindText, Content: "fixed"},
			{Kind: KindBox, FlexGrow: 1, Children: []*Input{
				{Kind: KindText, Content: "grow"},
			}},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then the flex-grow child fills the remaining space
	fixed := box.Children[0]
	if fixed.Width != 5 {
		t.Errorf("fixed child Width = %d, want 5", fixed.Width)
	}
	grow := box.Children[1]
	// 40 total - 5 fixed = 35 remaining
	if grow.Width != 35 {
		t.Errorf("flex-grow child Width = %d, want 35", grow.Width)
	}
}

func TestLayoutFlexGrowTwoEqualChildren(t *testing.T) {
	// Given a fixed-width row with two equal flex-grow children
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Children: []*Input{
			{Kind: KindBox, FlexGrow: 1, Children: []*Input{
				{Kind: KindText, Content: "a"},
			}},
			{Kind: KindBox, FlexGrow: 1, Children: []*Input{
				{Kind: KindText, Content: "b"},
			}},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then each child gets half the space
	if box.Children[0].Width != 20 {
		t.Errorf("first child Width = %d, want 20", box.Children[0].Width)
	}
	if box.Children[1].Width != 20 {
		t.Errorf("second child Width = %d, want 20", box.Children[1].Width)
	}
}

func TestLayoutFlexGrowUnequalRatios(t *testing.T) {
	// Given a fixed-width row with unequal flex-grow ratios
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 30,
		Children: []*Input{
			{Kind: KindBox, FlexGrow: 1},
			{Kind: KindBox, FlexGrow: 2},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then space is distributed 1:2 (10 and 20)
	if box.Children[0].Width != 10 {
		t.Errorf("first child Width = %d, want 10", box.Children[0].Width)
	}
	if box.Children[1].Width != 20 {
		t.Errorf("second child Width = %d, want 20", box.Children[1].Width)
	}
}

func TestLayoutFlexGrowInColumn(t *testing.T) {
	// Given a fixed-height column with flex-grow children
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedHeight: 20,
		Children: []*Input{
			{Kind: KindText, Content: "header"},
			{Kind: KindBox, FlexGrow: 1, Children: []*Input{
				{Kind: KindText, Content: "body"},
			}},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then the flex-grow child fills the remaining height
	header := box.Children[0]
	if header.Height != 1 {
		t.Errorf("header Height = %d, want 1", header.Height)
	}
	body := box.Children[1]
	// 20 total - 1 header = 19 remaining
	if body.Height != 19 {
		t.Errorf("body Height = %d, want 19", body.Height)
	}
}

func TestLayoutFlexGrowWithPaddingAndBorder(t *testing.T) {
	// Given a fixed-width row with padding, border, and flex-grow
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Padding:    Padding{0, 2, 0, 2},
		Border:     "single",
		Children: []*Input{
			{Kind: KindBox, FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then flex-grow child fills content area
	// Content width = 40 - 2*border - padLeft - padRight = 40 - 2 - 2 - 2 = 34
	child := box.Children[0]
	if child.Width != 34 {
		t.Errorf("flex-grow child Width = %d, want 34", child.Width)
	}
}

func TestLayoutFlexGrowWithGap(t *testing.T) {
	// Given flex-grow with gap between children
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 30,
		Gap:        2,
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
			{Kind: KindBox, FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then remaining space accounts for gap
	// 30 total - 2 fixed - 2 gap = 26 for flex-grow
	fixed := box.Children[0]
	if fixed.Width != 2 {
		t.Errorf("fixed child Width = %d, want 2", fixed.Width)
	}
	grow := box.Children[1]
	if grow.Width != 26 {
		t.Errorf("flex-grow child Width = %d, want 26", grow.Width)
	}
}

func TestLayoutFlexGrowPositionsCorrectly(t *testing.T) {
	// Given flex-grow children, their X positions should be correct
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 30,
		Children: []*Input{
			{Kind: KindText, Content: "ab"},
			{Kind: KindBox, FlexGrow: 1},
			{Kind: KindText, Content: "cd"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then positions are sequential
	if box.Children[0].X != 0 {
		t.Errorf("child[0].X = %d, want 0", box.Children[0].X)
	}
	if box.Children[0].Width != 2 {
		t.Errorf("child[0].Width = %d, want 2", box.Children[0].Width)
	}
	// Flex child: 30 - 2 - 2 = 26
	if box.Children[1].X != 2 {
		t.Errorf("child[1].X = %d, want 2", box.Children[1].X)
	}
	if box.Children[1].Width != 26 {
		t.Errorf("child[1].Width = %d, want 26", box.Children[1].Width)
	}
	if box.Children[2].X != 28 {
		t.Errorf("child[2].X = %d, want 28", box.Children[2].X)
	}
}

func TestLayoutFlexGrowNoFixedParentUsesAvailWidth(t *testing.T) {
	// Given a row without fixed width, flex-grow should use available width
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
			{Kind: KindBox, FlexGrow: 1},
		},
	}

	// When laid out with 80 available
	box := Layout(input, 80, 24)

	// Then the row fills available width and flex child gets the rest
	if box.Width != 80 {
		t.Errorf("parent Width = %d, want 80 (available)", box.Width)
	}
	grow := box.Children[1]
	if grow.Width != 78 {
		t.Errorf("flex-grow child Width = %d, want 78 (80-2)", grow.Width)
	}
}

func TestLayoutFlexGrowZeroDoesNotGrow(t *testing.T) {
	// Given FlexGrow=0 (default), child should not grow
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children keep their natural size
	if box.Children[0].Width != 5 {
		t.Errorf("child[0].Width = %d, want 5", box.Children[0].Width)
	}
	if box.Children[1].Width != 5 {
		t.Errorf("child[1].Width = %d, want 5", box.Children[1].Width)
	}
}
