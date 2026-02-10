package layout

import (
	"testing"
)

func TestLayoutColumnGapBetweenChildren(t *testing.T) {
	// Given a column box with gap=1 and three children
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Gap:       1,
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bbb"},
			{Kind: KindText, Content: "ccc"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are spaced with 1 cell gap between them
	wantY := []int{0, 2, 4} // 0, 1+1gap, 2+1gap+1+1gap
	for i, want := range wantY {
		if box.Children[i].Y != want {
			t.Errorf("child[%d].Y = %d, want %d", i, box.Children[i].Y, want)
		}
	}

	// Height includes gaps: 3 children of height 1 + 2 gaps of 1 = 5
	if box.Height != 5 {
		t.Errorf("Height = %d, want 5 (3 children + 2 gaps)", box.Height)
	}
}

func TestLayoutRowGapBetweenChildren(t *testing.T) {
	// Given a row box with gap=2
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Gap:       2,
		Children: []*Input{
			{Kind: KindText, Content: "aa"},
			{Kind: KindText, Content: "bb"},
			{Kind: KindText, Content: "cc"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are spaced with 2 cell gaps
	wantX := []int{0, 4, 8} // 0, 2+2gap, 2+2gap+2+2gap
	for i, want := range wantX {
		if box.Children[i].X != want {
			t.Errorf("child[%d].X = %d, want %d", i, box.Children[i].X, want)
		}
	}

	// Width = 3*2 + 2*2(gaps) = 10
	if box.Width != 10 {
		t.Errorf("Width = %d, want 10 (3 children + 2 gaps)", box.Width)
	}
}

func TestLayoutGapWithSingleChild(t *testing.T) {
	// Given gap=5 but only one child, no gap should be added
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Gap:       5,
		Children: []*Input{
			{Kind: KindText, Content: "only"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then no gap, just the child height
	if box.Height != 1 {
		t.Errorf("Height = %d, want 1 (no gap with single child)", box.Height)
	}
}

func TestLayoutGapZeroIsDefault(t *testing.T) {
	// Given gap=0 (default), children should be adjacent
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Gap:       0,
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bbb"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are adjacent (same as before gap feature)
	if box.Children[1].Y != 1 {
		t.Errorf("child[1].Y = %d, want 1 (no gap)", box.Children[1].Y)
	}
	if box.Height != 2 {
		t.Errorf("Height = %d, want 2", box.Height)
	}
}

func TestLayoutColumnGapWithPaddingAndBorder(t *testing.T) {
	// Given a column box with gap, padding, and border
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Gap:       1,
		Padding:   Padding{1, 0, 1, 0},
		Border:    "single",
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bbb"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then offsetY = border(1) + padTop(1) = 2
	first := box.Children[0]
	if first.Y != 2 {
		t.Errorf("first child Y = %d, want 2 (border+padTop)", first.Y)
	}
	second := box.Children[1]
	if second.Y != 4 {
		t.Errorf("second child Y = %d, want 4 (2+1height+1gap)", second.Y)
	}

	// Height = 2 children(1 each) + 1 gap + padTop(1) + padBottom(1) + border(2) = 7
	if box.Height != 7 {
		t.Errorf("Height = %d, want 7", box.Height)
	}
}

func TestLayoutRowGapWithPaddingAndBorder(t *testing.T) {
	// Given a row box with gap, padding, and border
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Gap:       2,
		Padding:   Padding{0, 1, 0, 1},
		Border:    "single",
		Children: []*Input{
			{Kind: KindText, Content: "aa"},
			{Kind: KindText, Content: "bb"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then offsetX = border(1) + padLeft(1) = 2
	first := box.Children[0]
	if first.X != 2 {
		t.Errorf("first child X = %d, want 2 (border+padLeft)", first.X)
	}
	second := box.Children[1]
	if second.X != 6 {
		t.Errorf("second child X = %d, want 6 (2+2width+2gap)", second.X)
	}

	// Width = 2 children(2 each) + 1 gap(2) + padLeft(1) + padRight(1) + border(2) = 10
	if box.Width != 10 {
		t.Errorf("Width = %d, want 10", box.Width)
	}
}
