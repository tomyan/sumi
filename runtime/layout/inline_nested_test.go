package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// B4c-2: nested inline elements — display:inline boxes join the IFC;
// their descendants' runs flow through line breaking, and their Box
// keeps the pairwise child structure with a union bounding rect.

func nestedInlineParagraph() *Input {
	return &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "a "},
		{Kind: KindBox, Tag: "em", Display: "inline", Style: render.Style{Italic: true}, Children: []*Input{
			{Kind: KindText, Content: "b "},
			{Kind: KindText, Tag: "strong", Content: "c", Style: render.Style{Bold: true}},
		}},
		{Kind: KindText, Content: " d"},
	}}
}

func TestNestedInlineElementJoinsFlow(t *testing.T) {
	// Given / When
	box := Layout(nestedInlineParagraph(), 40, 24)

	// Then: one line "a b c d"; em is the union of its runs.
	if box.Height != 1 {
		t.Fatalf("height = %d, want 1", box.Height)
	}
	em := box.Children[1]
	if em.X != 2 || em.Width != 3 || em.Height != 1 {
		t.Errorf("em rect = (%d w%d h%d), want (2 w3 h1)", em.X, em.Width, em.Height)
	}
	if len(em.Children) != 2 {
		t.Fatalf("em children = %d, want 2 (pairwise mapping)", len(em.Children))
	}
	// Layout returns absolute positions: a=0, b=2, c=4.
	assertFragment(t, "em text", em.Children[0], 0, Fragment{X: 0, Y: 0, Text: "b "})
	if em.Children[1].X != 4 {
		t.Errorf("strong X = %d, want 4 (absolute)", em.Children[1].X)
	}
	assertFragment(t, "strong", em.Children[1], 0, Fragment{X: 0, Y: 0, Text: "c"})
	// Trailing text continues after the em.
	assertFragment(t, "tail", box.Children[2], 0, Fragment{X: 0, Y: 0, Text: " d"})
	if box.Children[2].X != 5 {
		t.Errorf("tail X = %d, want 5", box.Children[2].X)
	}
}

func TestNestedInlineRendersInheritedStyles(t *testing.T) {
	// Given / When
	box := Layout(nestedInlineParagraph(), 40, 3)
	buf := render.NewBuffer(40, 3)
	RenderTree(buf, box, nil)

	// Then: "a b c d" with em italic on "b", em+strong on "c".
	row := ""
	for col := 0; col < 7; col++ {
		ch := buf.Cell(0, col).Ch
		if ch == 0 {
			ch = ' '
		}
		row += string(ch)
	}
	if row != "a b c d" {
		t.Errorf("row = %q, want %q", row, "a b c d")
	}
	if !buf.Cell(0, 2).Style.Italic {
		t.Errorf("'b' should inherit em italic: %+v", buf.Cell(0, 2).Style)
	}
	if !buf.Cell(0, 4).Style.Italic || !buf.Cell(0, 4).Style.Bold {
		t.Errorf("'c' should be italic+bold: %+v", buf.Cell(0, 4).Style)
	}
	if buf.Cell(0, 0).Style.Italic {
		t.Errorf("'a' must not be italic")
	}
}

func TestNestedInlineWrapsAcrossLines(t *testing.T) {
	// Given
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "aa "},
		{Kind: KindBox, Display: "inline", Children: []*Input{
			{Kind: KindText, Content: "bb cc"},
		}},
	}}

	// When
	box := Layout(p, 5, 24)

	// Then: "aa bb" / "cc" — the inline box spans two lines.
	em := box.Children[1]
	if em.Height != 2 {
		t.Errorf("inline box height = %d, want 2: %+v", em.Height, em)
	}
	inner := em.Children[0]
	if len(inner.Fragments) != 2 {
		t.Fatalf("inner fragments = %+v, want 2", inner.Fragments)
	}
}

func TestInlineBoxWithBlockContentStacks(t *testing.T) {
	// Given: an inline element containing a real box is not
	// inline-eligible — it stacks as block-level.
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{
		{Kind: KindText, Content: "text"},
		{Kind: KindBox, Display: "inline", Children: []*Input{
			{Kind: KindBox, FixedWidth: 3, FixedHeight: 1},
		}},
	}}

	// When
	box := Layout(p, 40, 24)

	// Then
	if box.Children[1].Y != 1 {
		t.Errorf("inline-with-box Y = %d, want 1 (stacked)", box.Children[1].Y)
	}
}
