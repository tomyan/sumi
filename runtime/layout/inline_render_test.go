package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// B4b: fragment painting and diffing.

func TestRenderInlineFlowMixedStyles(t *testing.T) {
	// Given
	p := inlineParagraph()

	// When
	box := Layout(p, 40, 3)
	buf := render.NewBuffer(40, 3)
	RenderTree(buf, box, nil)

	// Then: text flows on one row; the strong run's cells are bold.
	row := ""
	for col := 0; col < 15; col++ {
		ch := buf.Cell(0, col).Ch
		if ch == 0 {
			ch = ' '
		}
		row += string(ch)
	}
	if row != "hello bold tail" {
		t.Errorf("row = %q, want %q", row, "hello bold tail")
	}
	if !buf.Cell(0, 6).Style.Bold {
		t.Errorf("cell (0,6) should be bold: %+v", buf.Cell(0, 6).Style)
	}
	if buf.Cell(0, 0).Style.Bold || buf.Cell(0, 11).Style.Bold {
		t.Errorf("plain runs must not be bold")
	}
}

func TestRenderInlineFlowWrapped(t *testing.T) {
	// Given
	p := inlineParagraph()

	// When: width 10 wraps after "hello bold".
	tree := &Input{Kind: KindBox, FixedWidth: 10, Children: []*Input{p}}
	lines := renderToString(tree, 10, 3)

	// Then
	if lines[0] != "hello bold" {
		t.Errorf("line 0 = %q, want %q", lines[0], "hello bold")
	}
	if lines[1] != "tail" {
		t.Errorf("line 1 = %q, want %q", lines[1], "tail")
	}
}

func TestInlineFlowFromStylesheetDisplayBlock(t *testing.T) {
	// Given: display:block arrives via the cascade, strong via UA bold.
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "p", Kind: KindBox, Children: []*Input{
			{Kind: KindText, Content: "hello "},
			{Tag: "strong", Kind: KindText, Content: "bold"},
		}},
	}}
	ss := sheet(t, `p { display: block; }`)

	// When
	ResolveStyles(tree, ss, 40, 3)
	lines := renderToString(tree, 40, 3)

	// Then: UA p{margin:1 0} puts the text on row 1.
	if lines[1] != "hello bold" {
		t.Errorf("line 1 = %q, want %q", lines[1], "hello bold")
	}
	if !tree.Children[0].Children[1].Style.Bold {
		t.Errorf("strong should be bold via UA")
	}
}

func TestDiffDetectsFragmentChange(t *testing.T) {
	// Given: same box except one fragment's text differs.
	old := &Box{Kind: KindText, Fragments: []Fragment{{X: 0, Y: 0, Text: "a"}}}
	new := &Box{Kind: KindText, Fragments: []Fragment{{X: 0, Y: 0, Text: "b"}}}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
}
