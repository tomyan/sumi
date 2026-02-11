package layout

import "testing"

func TestLayoutFixedIsViewportRelative(t *testing.T) {
	// Given — fixed child inside a nested box
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  80,
		FixedHeight: 24,
		Padding:     Padding{Top: 5, Left: 5, Bottom: 5, Right: 5},
		Children: []*Input{
			{
				Kind:        KindBox,
				FixedWidth:  40,
				FixedHeight: 10,
				Children: []*Input{
					{Kind: KindText, Content: "fix", Position: "fixed", Top: 1, Left: 2},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — fixed child at viewport-relative (2,1), ignoring parent nesting
	fixed := box.Children[0].Children[0]
	if fixed.X != 2 || fixed.Y != 1 {
		t.Errorf("expected fixed at (2,1), got (%d,%d)", fixed.X, fixed.Y)
	}
}

func TestLayoutFixedDoesNotAffectFlowSiblings(t *testing.T) {
	// Given
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  80,
		FixedHeight: 24,
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindBox, Position: "fixed", Top: 0, Left: 0, FixedWidth: 80, FixedHeight: 24},
			{Kind: KindText, Content: "second"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — second flow child at Y=1
	if box.Children[2].Y != 1 {
		t.Errorf("expected Y=1, got %d", box.Children[2].Y)
	}
}

func TestLayoutFixedStretchWithOffsets(t *testing.T) {
	// Given — fixed with opposing offsets stretches to fill between them
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  80,
		FixedHeight: 24,
		Children: []*Input{
			{Kind: KindBox, Position: "fixed", Top: 2, Left: 5, Right: 5, Bottom: 2},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — stretches: width=70 (80-5-5), height=20 (24-2-2)
	fixed := box.Children[0]
	if fixed.Width != 70 {
		t.Errorf("expected width=70, got %d", fixed.Width)
	}
	if fixed.Height != 20 {
		t.Errorf("expected height=20, got %d", fixed.Height)
	}
	if fixed.X != 5 || fixed.Y != 2 {
		t.Errorf("expected at (5,2), got (%d,%d)", fixed.X, fixed.Y)
	}
}
