package layout

import "testing"

// B4f: display:contents — the element generates no box of its own; its
// children participate directly in the parent's flow. The Input↔Box
// pairwise mapping keeps a union-rect placeholder.

func TestContentsChildrenJoinParentIFC(t *testing.T) {
	// Given: text inside contents merges with sibling text runs.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "a "},
		{Kind: KindBox, Display: "contents", Children: []*Input{
			{Kind: KindText, Content: "b"},
		}},
		{Kind: KindText, Content: " c"},
	}}

	// When
	box := Layout(p, 40, 24)

	// Then: one line "a b c".
	if box.Height != 1 {
		t.Fatalf("height = %d, want 1", box.Height)
	}
	contents := box.Children[1]
	if len(contents.Children) != 1 {
		t.Fatalf("contents children = %d, want 1", len(contents.Children))
	}
	assertFragment(t, "inner", contents.Children[0], 0, Fragment{X: 0, Y: 0, Text: "b"})
	if contents.X != 2 || contents.Width != 1 {
		t.Errorf("contents rect = (%d w%d), want (2 w1)", contents.X, contents.Width)
	}
	assertFragment(t, "tail", box.Children[2], 0, Fragment{X: 0, Y: 0, Text: " c"})
}

func TestContentsBlockChildrenFlattenIntoFlow(t *testing.T) {
	// Given: contents wrapping two block children between text rows.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "top"},
		{Kind: KindBox, Display: "contents", Children: []*Input{
			{Kind: KindBox, Display: "block", FixedHeight: 1, Children: []*Input{{Kind: KindText, Content: "x"}}},
			{Kind: KindBox, Display: "block", FixedHeight: 1, Children: []*Input{{Kind: KindText, Content: "y"}}},
		}},
		{Kind: KindText, Content: "bottom"},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then: four rows — top, x, y, bottom.
	contents := box.Children[1]
	if contents.Children[0].Y != 1 || contents.Children[1].Y != 2 {
		t.Errorf("blocks at Y %d,%d, want 1,2",
			contents.Children[0].Y, contents.Children[1].Y)
	}
	if got := box.Children[2].Y; got != 3 {
		t.Errorf("bottom Y = %d, want 3", got)
	}
	if contents.Y != 1 || contents.Height != 2 {
		t.Errorf("contents rect Y=%d H=%d, want Y=1 H=2", contents.Y, contents.Height)
	}
}

func TestNestedContentsFlatten(t *testing.T) {
	// Given
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindBox, Display: "contents", Children: []*Input{
			{Kind: KindBox, Display: "contents", Children: []*Input{
				{Kind: KindText, Content: "deep"},
			}},
		}},
	}}

	// When
	box := Layout(p, 20, 24)

	// Then
	inner := box.Children[0].Children[0].Children[0]
	assertFragment(t, "deep", inner, 0, Fragment{X: 0, Y: 0, Text: "deep"})
}
