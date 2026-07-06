package layout

import "testing"

// B4 follow-up: text-align shifts whole IFC lines within the container.

func TestIFCTextAlignCenter(t *testing.T) {
	// Given: "hi" centred in 11 cells.
	p := &Input{Kind: KindBox, Display: "block", TextAlign: "center", Children: []*Input{
		{Kind: KindText, Content: "hi"},
	}}

	// When
	box := Layout(p, 11, 24)

	// Then
	text := box.Children[0]
	if text.X != 4 {
		t.Errorf("text X = %d, want 4 ((11-2)/2)", text.X)
	}
}

func TestIFCTextAlignRightMixedRuns(t *testing.T) {
	// Given: "a bb" right-aligned in 10 — both runs shift together.
	p := &Input{Kind: KindBox, Display: "block", TextAlign: "right", Children: []*Input{
		{Kind: KindText, Content: "a "},
		{Kind: KindText, Tag: "strong", Content: "bb"},
	}}

	// When
	box := Layout(p, 10, 24)

	// Then: line "a bb" (4 wide) starts at 6.
	if got := box.Children[0].X; got != 6 {
		t.Errorf("first run X = %d, want 6", got)
	}
	if got := box.Children[1].X; got != 8 {
		t.Errorf("strong X = %d, want 8", got)
	}
}

func TestIFCTextAlignCentersEachWrappedLine(t *testing.T) {
	// Given: "aaaa bb" centred at width 6 → lines "aaaa" (x1), "bb" (x2).
	p := &Input{Kind: KindBox, Display: "block", TextAlign: "center", Children: []*Input{
		{Kind: KindText, Content: "aaaa bb"},
	}}

	// When
	box := Layout(p, 6, 24)

	// Then: fragments are box-relative; the box spans both lines.
	text := box.Children[0]
	if len(text.Fragments) != 2 {
		t.Fatalf("fragments = %+v, want 2", text.Fragments)
	}
	first := text.X + text.Fragments[0].X
	second := text.X + text.Fragments[1].X
	if first != 1 || second != 2 {
		t.Errorf("line starts = %d, %d, want 1, 2", first, second)
	}
}

func TestIFCTextAlignLeftIsUnshifted(t *testing.T) {
	// Given
	p := &Input{Kind: KindBox, Display: "block", TextAlign: "left", Children: []*Input{
		{Kind: KindText, Content: "hi"},
	}}

	// When
	box := Layout(p, 11, 24)

	// Then
	if got := box.Children[0].X; got != 0 {
		t.Errorf("text X = %d, want 0", got)
	}
}
