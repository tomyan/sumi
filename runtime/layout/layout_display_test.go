package layout

import "testing"

func TestLayoutDisplayNoneChildExcludedFromColumn(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "visible"},
			{Kind: KindText, Content: "hidden", Display: "none"},
			{Kind: KindText, Content: "also visible"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — hidden child takes no space, third child at Y=1 not Y=2
	if box.Children[2].Y != 1 {
		t.Errorf("expected third child Y=1, got %d", box.Children[2].Y)
	}
}

func TestLayoutDisplayNoneInRow(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "row",
		Children: []*Input{
			{Kind: KindText, Content: "AAA"},
			{Kind: KindText, Content: "BBB", Display: "none"},
			{Kind: KindText, Content: "CCC"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — hidden child takes no space, third child at X=3 not X=6
	if box.Children[2].X != 3 {
		t.Errorf("expected third child X=3, got %d", box.Children[2].X)
	}
}

func TestLayoutDisplayNoneWithFlexGrow(t *testing.T) {
	// Given — hidden child has flex-grow but should be ignored
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 30,
		Children: []*Input{
			{Kind: KindBox, FlexGrow: 1, Children: []*Input{{Kind: KindText, Content: "A"}}},
			{Kind: KindBox, FlexGrow: 1, Display: "none", Children: []*Input{{Kind: KindText, Content: "B"}}},
			{Kind: KindBox, FlexGrow: 1, Children: []*Input{{Kind: KindText, Content: "C"}}},
		},
	}

	// When
	box := Layout(input, 30, 24)

	// Then — space split between two visible children (15 each), not three (10 each)
	if box.Children[0].Width != 15 {
		t.Errorf("expected first child width=15, got %d", box.Children[0].Width)
	}
	if box.Children[2].Width != 15 {
		t.Errorf("expected third child width=15, got %d", box.Children[2].Width)
	}
}

func TestLayoutDisplayNoneChildrenCount(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "A"},
			{Kind: KindText, Content: "B", Display: "none"},
			{Kind: KindText, Content: "C"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — box.Children length matches input.Children length with nil placeholder
	if len(box.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(box.Children))
	}
	if box.Children[1] != nil {
		t.Errorf("expected nil placeholder for hidden child, got %+v", box.Children[1])
	}
}

func TestLayoutDisplayNoneDoesNotAffectParentSize(t *testing.T) {
	// Given — hidden child is taller than visible children
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "short"},
			{Kind: KindBox, Display: "none", FixedWidth: 50, FixedHeight: 50},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — parent sizes to visible content only
	if box.Height != 1 {
		t.Errorf("expected parent height=1, got %d", box.Height)
	}
	if box.Width > 10 {
		t.Errorf("expected parent width<=10, got %d", box.Width)
	}
}
