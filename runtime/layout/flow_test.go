package layout

import "testing"

// B4c-1: block flow — display:block containers stack block-level
// children, group consecutive inline-level children into IFC segments,
// fill the available width, and ignore flex attributes.

func TestBlockFlowMixedInlineAndBlockChildren(t *testing.T) {
	// Given: text, box, text — the classic anonymous-flow shape.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "above"},
		{Kind: KindBox, FixedWidth: 5, FixedHeight: 2},
		{Kind: KindText, Content: "below"},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then: first IFC line, then the box, then the second IFC.
	above, mid, below := box.Children[0], box.Children[1], box.Children[2]
	assertFragment(t, "above", above, 0, Fragment{X: 0, Y: 0, Text: "above"})
	if above.Y != 0 {
		t.Errorf("above Y = %d, want 0", above.Y)
	}
	if mid.Y != 1 {
		t.Errorf("box Y = %d, want 1", mid.Y)
	}
	if below.Y != 3 {
		t.Errorf("below Y = %d, want 3", below.Y)
	}
	assertFragment(t, "below", below, 0, Fragment{X: 0, Y: 0, Text: "below"})
}

func TestBlockFlowChildBlockFillsWidth(t *testing.T) {
	// Given: an auto-width block child.
	p := &Input{Kind: KindBox, Display: "block", FixedWidth: 20, Children: []*Input{
		{Kind: KindBox, Display: "block", Children: []*Input{
			{Kind: KindText, Content: "x"},
		}},
	}}

	// When
	box := Layout(p, 40, 24)

	// Then: the child block spans the parent's content width.
	if got := box.Children[0].Width; got != 20 {
		t.Errorf("child width = %d, want 20", got)
	}
}

func TestBlockFlowIgnoresFlexAttributes(t *testing.T) {
	// Given: direction/gap must not apply to a block container.
	p := &Input{Kind: KindBox, Display: "block", Direction: "row", Gap: 3, Children: []*Input{
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1},
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then: stacked vertically, no gap.
	if box.Children[0].Y != 0 || box.Children[1].Y != 1 {
		t.Errorf("children Y = %d, %d, want 0, 1", box.Children[0].Y, box.Children[1].Y)
	}
	if box.Children[1].X != 0 {
		t.Errorf("second child X = %d, want 0 (not a row)", box.Children[1].X)
	}
}

func TestBlockFlowExplicitFlexStillFlexes(t *testing.T) {
	// Given
	p := &Input{Kind: KindBox, Display: "flex", Direction: "row", Gap: 1, Children: []*Input{
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1},
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then: row with gap.
	if box.Children[1].X != 5 {
		t.Errorf("second child X = %d, want 5", box.Children[1].X)
	}
}

func TestBlockFlowMarginsAndAutoCentring(t *testing.T) {
	// Given: plain margins stack (no collapse until B4e); auto margins centre.
	p := &Input{Kind: KindBox, Display: "block", FixedWidth: 20, Children: []*Input{
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1, Margin: Margin{Bottom: 2}},
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1, Margin: Margin{Top: 1, AutoLeft: true, AutoRight: true}},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then
	second := box.Children[1]
	if second.Y != 4 {
		t.Errorf("second Y = %d, want 4 (1 + 2 + 1)", second.Y)
	}
	if second.X != 8 {
		t.Errorf("second X = %d, want 8 (centred in 20)", second.X)
	}
}

func TestBlockFlowFillsAvailableWidth(t *testing.T) {
	// Given: a block container with no fixed width.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "hi"},
	}}

	// When
	box := Layout(p, 30, 24)

	// Then
	if box.Width != 30 {
		t.Errorf("block width = %d, want 30 (fills available)", box.Width)
	}
}
