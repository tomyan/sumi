package layout

import "testing"

func TestLayoutAbsoluteAtTopLeft(t *testing.T) {
	// Given — absolute child at top:0 left:0, parent has border
	input := &Input{
		Kind:        KindBox,
		Border:      "single",
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindText, Content: "flow"},
			{Kind: KindText, Content: "abs", Position: "absolute", Top: 0, Left: 0},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — absolute child at parent content origin (1,1 inside border)
	abs := box.Children[1]
	if abs.X != 1 || abs.Y != 1 {
		t.Errorf("expected abs at (1,1), got (%d,%d)", abs.X, abs.Y)
	}
}

func TestLayoutAbsoluteFromBottomRight(t *testing.T) {
	// Given — absolute child positioned from bottom-right
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindText, Content: "abs", Position: "absolute", Bottom: 1, Right: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — 3-char text at bottom-right with 1-cell inset
	abs := box.Children[0]
	expectedX := 20 - 1 - 3 // parentW - right - childW
	expectedY := 10 - 1 - 1 // parentH - bottom - childH
	if abs.X != expectedX {
		t.Errorf("expected X=%d, got %d", expectedX, abs.X)
	}
	if abs.Y != expectedY {
		t.Errorf("expected Y=%d, got %d", expectedY, abs.Y)
	}
}

func TestLayoutAbsoluteStretchBothAxes(t *testing.T) {
	// Given — top:1 bottom:1 left:1 right:1 without fixed size stretches
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindBox, Position: "absolute", Top: 1, Bottom: 1, Left: 1, Right: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — stretched: width=18, height=8
	abs := box.Children[0]
	if abs.Width != 18 {
		t.Errorf("expected width=18, got %d", abs.Width)
	}
	if abs.Height != 8 {
		t.Errorf("expected height=8, got %d", abs.Height)
	}
	if abs.X != 1 || abs.Y != 1 {
		t.Errorf("expected at (1,1), got (%d,%d)", abs.X, abs.Y)
	}
}

func TestLayoutAbsoluteDoesNotAffectFlowSiblings(t *testing.T) {
	// Given — absolute child should not push flow children
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindBox, Position: "absolute", Top: 0, Left: 0, FixedWidth: 50, FixedHeight: 50},
			{Kind: KindText, Content: "second"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — second flow child is at Y=1 (after first), not pushed by absolute
	if box.Children[2].Y != 1 {
		t.Errorf("expected second flow child Y=1, got %d", box.Children[2].Y)
	}
}

func TestLayoutAbsoluteDoesNotAffectParentSize(t *testing.T) {
	// Given — large absolute child should not grow parent
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
			{Kind: KindBox, Position: "absolute", Top: 0, Left: 0, FixedWidth: 100, FixedHeight: 100},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — parent sizes to flow children only
	if box.Height != 1 {
		t.Errorf("expected parent height=1, got %d", box.Height)
	}
}

func TestLayoutAbsoluteWithBorder(t *testing.T) {
	// Given — absolute child respects parent content area (inside border+padding)
	input := &Input{
		Kind:        KindBox,
		Border:      "single",
		Padding:     Padding{Top: 1, Right: 1, Bottom: 1, Left: 1},
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindText, Content: "abs", Position: "absolute", Top: 0, Left: 0},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — content area starts at (2,2) = border(1) + padding(1)
	abs := box.Children[0]
	if abs.X != 2 || abs.Y != 2 {
		t.Errorf("expected abs at (2,2), got (%d,%d)", abs.X, abs.Y)
	}
}

func TestCodegenPositionAbsolute(t *testing.T) {
	// Codegen test is in codegen/codegen_position_test.go
	// This is a layout-level test that absolute positioning is correctly applied
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  40,
		FixedHeight: 20,
		Children: []*Input{
			{Kind: KindBox, Position: "absolute", Top: 5, Left: 10, FixedWidth: 10, FixedHeight: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	abs := box.Children[0]
	if abs.X != 10 || abs.Y != 5 {
		t.Errorf("expected at (10,5), got (%d,%d)", abs.X, abs.Y)
	}
	if abs.Width != 10 || abs.Height != 5 {
		t.Errorf("expected 10x5, got %dx%d", abs.Width, abs.Height)
	}
}
