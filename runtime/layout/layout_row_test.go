package layout

import (
	"testing"
)

func TestLayoutRowTwoTextNodes(t *testing.T) {
	// Given two text children in a row container
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are placed side by side
	if len(box.Children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(box.Children))
	}

	first := box.Children[0]
	if first.X != 0 || first.Y != 0 {
		t.Errorf("first child position = (%d, %d), want (0, 0)", first.X, first.Y)
	}
	if first.Width != 5 {
		t.Errorf("first child Width = %d, want 5", first.Width)
	}

	second := box.Children[1]
	if second.X != 5 || second.Y != 0 {
		t.Errorf("second child position = (%d, %d), want (5, 0)", second.X, second.Y)
	}
	if second.Width != 5 {
		t.Errorf("second child Width = %d, want 5", second.Width)
	}

	// Parent auto-width = sum of child widths, auto-height = max child height
	if box.Width != 10 {
		t.Errorf("parent Width = %d, want 10 (sum of children)", box.Width)
	}
	if box.Height != 1 {
		t.Errorf("parent Height = %d, want 1 (max of children)", box.Height)
	}
}

func TestLayoutRowSingleChild(t *testing.T) {
	// Given a row container with one child
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindText, Content: "only"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	child := box.Children[0]
	if child.X != 0 || child.Y != 0 {
		t.Errorf("child position = (%d, %d), want (0, 0)", child.X, child.Y)
	}
	if box.Width != 4 {
		t.Errorf("parent Width = %d, want 4", box.Width)
	}
	if box.Height != 1 {
		t.Errorf("parent Height = %d, want 1", box.Height)
	}
}

func TestLayoutRowThreeChildren(t *testing.T) {
	// Given three children in a row
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bb"},
			{Kind: KindText, Content: "c"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are placed left to right at correct X positions
	wantX := []int{0, 3, 5}
	for i, want := range wantX {
		if box.Children[i].X != want {
			t.Errorf("child[%d].X = %d, want %d", i, box.Children[i].X, want)
		}
		if box.Children[i].Y != 0 {
			t.Errorf("child[%d].Y = %d, want 0", i, box.Children[i].Y)
		}
	}

	// Width = sum of children (3+2+1=6)
	if box.Width != 6 {
		t.Errorf("parent Width = %d, want 6", box.Width)
	}
}

func TestLayoutRowWithDifferentHeights(t *testing.T) {
	// Given row children with different heights, parent height = max
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 1},
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 3},
			{Kind: KindBox, FixedWidth: 5, FixedHeight: 2},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then parent height is the max child height
	if box.Height != 3 {
		t.Errorf("parent Height = %d, want 3 (max of children)", box.Height)
	}
	if box.Width != 15 {
		t.Errorf("parent Width = %d, want 15 (sum of children)", box.Width)
	}
}

func TestLayoutRowWithPaddingAndBorder(t *testing.T) {
	// Given a row container with padding and border
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Padding:   Padding{1, 2, 1, 2},
		Border:    "single",
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bbb"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are offset by border(1) + padding
	// offsetX = 1(border) + 2(padLeft) = 3
	// offsetY = 1(border) + 1(padTop) = 2
	first := box.Children[0]
	if first.X != 3 {
		t.Errorf("first child X = %d, want 3 (border+padLeft)", first.X)
	}
	if first.Y != 2 {
		t.Errorf("first child Y = %d, want 2 (border+padTop)", first.Y)
	}

	second := box.Children[1]
	if second.X != 6 {
		t.Errorf("second child X = %d, want 6 (3+3)", second.X)
	}
	if second.Y != 2 {
		t.Errorf("second child Y = %d, want 2 (border+padTop)", second.Y)
	}

	// Width = children(3+3) + padLeft(2) + padRight(2) + border(2) = 12
	if box.Width != 12 {
		t.Errorf("parent Width = %d, want 12", box.Width)
	}
	// Height = maxChildHeight(1) + padTop(1) + padBottom(1) + border(2) = 5
	if box.Height != 5 {
		t.Errorf("parent Height = %d, want 5", box.Height)
	}
}

func TestLayoutRowNestedInColumn(t *testing.T) {
	// Given a column container with a row container inside
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Children: []*Input{
			{Kind: KindText, Content: "header"},
			{
				Kind:      KindBox,
				Direction: "row",
				Children: []*Input{
					{Kind: KindText, Content: "left"},
					{Kind: KindText, Content: "right"},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then the row container is the second child of the column
	row := box.Children[1]
	if row.Y != 1 {
		t.Errorf("row Y = %d, want 1 (below header)", row.Y)
	}
	// Row children should be side by side within the row
	if len(row.Children) != 2 {
		t.Fatalf("row children = %d, want 2", len(row.Children))
	}
	left := row.Children[0]
	right := row.Children[1]
	if left.X != 0 {
		t.Errorf("left.X = %d, want 0", left.X)
	}
	if right.X != 4 {
		t.Errorf("right.X = %d, want 4 (after 'left')", right.X)
	}
	// Row stretches to fill parent width (default align=stretch)
	if row.Width != 80 {
		t.Errorf("row Width = %d, want 80 (stretched to fill parent)", row.Width)
	}
}

func TestLayoutRowEmpty(t *testing.T) {
	// Given an empty row container
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
	}

	// When
	box := Layout(input, 80, 24)

	// Then dimensions are zero
	if box.Width != 0 {
		t.Errorf("Width = %d, want 0", box.Width)
	}
	if box.Height != 0 {
		t.Errorf("Height = %d, want 0", box.Height)
	}
}
