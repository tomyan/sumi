package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// B4b: inline formatting context — sibling text runs share line breaking;
// each run's Box carries box-relative Fragments.

func inlineParagraph() *Input {
	return &Input{
		Kind:    KindBox,
		Tag:     "p",
		Display: "block",
		Children: []*Input{
			{Kind: KindText, Content: "hello "},
			{Kind: KindText, Tag: "strong", Content: "bold", Style: render.Style{Bold: true}},
			{Kind: KindText, Content: " tail"},
		},
	}
}

func TestInlineFlowRunsShareOneLine(t *testing.T) {
	// Given
	p := inlineParagraph()

	// When
	box := Layout(p, 40, 24)

	// Then: all three runs flow on one line, positioned consecutively.
	if box.Height != 1 {
		t.Fatalf("height = %d, want 1", box.Height)
	}
	hello, strong, tail := box.Children[0], box.Children[1], box.Children[2]
	assertFragment(t, "hello", hello, 0, Fragment{X: 0, Y: 0, Text: "hello "})
	assertFragment(t, "strong", strong, 0, Fragment{X: 0, Y: 0, Text: "bold"})
	assertFragment(t, "tail", tail, 0, Fragment{X: 0, Y: 0, Text: " tail"})
	if strong.X != 6 || tail.X != 10 {
		t.Errorf("run X positions = %d, %d, want 6, 10", strong.X, tail.X)
	}
}

func TestInlineFlowWrapsAtRunBoundary(t *testing.T) {
	// Given: "hello bold" fills exactly 10 cells; the breaking space
	// before "tail" is consumed at the wrap.
	p := inlineParagraph()

	// When
	box := Layout(p, 10, 24)

	// Then
	if box.Height != 2 {
		t.Fatalf("height = %d, want 2: %+v", box.Height, box)
	}
	strong, tail := box.Children[1], box.Children[2]
	if strong.X != 6 || strong.Y != 0 {
		t.Errorf("strong at (%d,%d), want (6,0)", strong.X, strong.Y)
	}
	assertFragment(t, "tail", tail, 0, Fragment{X: 0, Y: 0, Text: "tail"})
	if tail.X != 0 || tail.Y != 1 {
		t.Errorf("tail box at (%d,%d), want (0,1)", tail.X, tail.Y)
	}
}

func TestInlineFlowWordSpansRunBoundary(t *testing.T) {
	// Given: "foo"+"bar" with no space — one unbreakable word across two
	// runs; width 4 forces a hard break inside the strong run.
	p := &Input{
		Kind:    KindBox,
		Display: "block",
		Children: []*Input{
			{Kind: KindText, Content: "foo"},
			{Kind: KindText, Tag: "strong", Content: "bar", Style: render.Style{Bold: true}},
		},
	}

	// When
	box := Layout(p, 4, 24)

	// Then: foob / ar — the strong run splits into two fragments.
	strong := box.Children[1]
	if len(strong.Fragments) != 2 {
		t.Fatalf("strong fragments = %+v, want 2", strong.Fragments)
	}
	assertFragment(t, "strong", strong, 0, Fragment{X: 3, Y: 0, Text: "b"})
	assertFragment(t, "strong", strong, 1, Fragment{X: 0, Y: 1, Text: "ar"})
	if box.Height != 2 {
		t.Errorf("height = %d, want 2", box.Height)
	}
}

func TestInlineFlowCollapsesWhitespace(t *testing.T) {
	// Given: runs of spaces/newlines/tabs collapse to single spaces.
	p := &Input{
		Kind:    KindBox,
		Display: "block",
		Children: []*Input{
			{Kind: KindText, Content: "a   b\n\tc "},
			{Kind: KindText, Tag: "em", Content: " d"},
		},
	}

	// When
	box := Layout(p, 40, 24)

	// Then: "a b c d" — inter-run whitespace collapses too.
	assertFragment(t, "text", box.Children[0], 0, Fragment{X: 0, Y: 0, Text: "a b c "})
	assertFragment(t, "em", box.Children[1], 0, Fragment{X: 0, Y: 0, Text: "d"})
	if box.Children[1].X != 6 {
		t.Errorf("em X = %d, want 6", box.Children[1].X)
	}
}

func TestInlineFlowSingleRunDegeneratesToWrap(t *testing.T) {
	// Given: one text child under display:block still lays out correctly.
	p := &Input{
		Kind:    KindBox,
		Display: "block",
		Children: []*Input{
			{Kind: KindText, Content: "hello world"},
		},
	}

	// When
	box := Layout(p, 6, 24)

	// Then
	text := box.Children[0]
	if len(text.Fragments) != 2 {
		t.Fatalf("fragments = %+v, want 2", text.Fragments)
	}
	assertFragment(t, "text", text, 0, Fragment{X: 0, Y: 0, Text: "hello"})
	assertFragment(t, "text", text, 1, Fragment{X: 0, Y: 1, Text: "world"})
}

func TestLegacyContainersNeverFragment(t *testing.T) {
	// Given: no display set — the legacy flex-column default keeps
	// per-node text layout (no IFC).
	p := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "text"},
			{Kind: KindText, Content: "more"},
		},
	}

	// When
	box := Layout(p, 40, 24)

	// Then: vertical stacking, no fragments.
	if box.Children[0].Fragments != nil {
		t.Errorf("unexpected fragments: %+v", box.Children[0].Fragments)
	}
	if box.Children[1].Y != 1 {
		t.Errorf("second text Y = %d, want 1 (stacked)", box.Children[1].Y)
	}
}

func assertFragment(t *testing.T, name string, box *Box, i int, want Fragment) {
	t.Helper()
	if len(box.Fragments) <= i {
		t.Fatalf("%s: fragments = %+v, want index %d", name, box.Fragments, i)
	}
	if box.Fragments[i] != want {
		t.Errorf("%s fragment[%d] = %+v, want %+v", name, i, box.Fragments[i], want)
	}
}
