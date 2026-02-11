package layout

import "testing"

func TestLayoutRelativeTopShiftsDown(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Top: 2},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — shifted 2 cells down from flow position (0)
	if box.Children[0].Y != 2 {
		t.Errorf("expected Y=2, got %d", box.Children[0].Y)
	}
}

func TestLayoutRelativeLeftShiftsRight(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Left: 3},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — shifted 3 cells right from flow position (0)
	if box.Children[0].X != 3 {
		t.Errorf("expected X=3, got %d", box.Children[0].X)
	}
}

func TestLayoutRelativeRightShiftsLeft(t *testing.T) {
	// Given — right:2 shifts left when left is not set
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Right: 2},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — shifted 2 cells left from flow position (0)
	if box.Children[0].X != -2 {
		t.Errorf("expected X=-2, got %d", box.Children[0].X)
	}
}

func TestLayoutRelativeBottomShiftsUp(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Bottom: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — shifted 1 cell up from flow position (0)
	if box.Children[0].Y != -1 {
		t.Errorf("expected Y=-1, got %d", box.Children[0].Y)
	}
}

func TestLayoutRelativeDoesNotAffectParentSize(t *testing.T) {
	// Given — child shifted down, parent should still size to flow position
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Top: 10},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — parent height based on flow position (1 line), not shifted position
	if box.Height != 1 {
		t.Errorf("expected parent height=1, got %d", box.Height)
	}
}

func TestLayoutRelativeDoesNotAffectSiblings(t *testing.T) {
	// Given — first child is relatively positioned, second should be unaffected
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "AAA", Position: "relative", Top: 5},
			{Kind: KindText, Content: "BBB"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — second child at Y=1 (normal flow after first)
	if box.Children[1].Y != 1 {
		t.Errorf("expected second child Y=1, got %d", box.Children[1].Y)
	}
}

func TestLayoutRelativeLeftWinsOverRight(t *testing.T) {
	// Given — both left and right set, left wins
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Left: 3, Right: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — left:3 applied, right ignored
	if box.Children[0].X != 3 {
		t.Errorf("expected X=3, got %d", box.Children[0].X)
	}
}

func TestLayoutRelativeTopWinsOverBottom(t *testing.T) {
	// Given — both top and bottom set, top wins
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello", Position: "relative", Top: 2, Bottom: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — top:2 applied, bottom ignored
	if box.Children[0].Y != 2 {
		t.Errorf("expected Y=2, got %d", box.Children[0].Y)
	}
}
