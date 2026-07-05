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
	// Given: p is display:block via the UA default (B4c-3), strong via
	// UA bold — no author CSS at all.
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "p", Kind: KindBox, Children: []*Input{
			{Kind: KindText, Content: "hello "},
			{Tag: "strong", Kind: KindText, Content: "bold"},
		}},
	}}

	// When
	ResolveStyles(tree, nil, 40, 3)
	lines := renderToString(tree, 40, 3)

	// Then: UA p{margin:1 0} puts the text on row 1.
	if lines[1] != "hello bold" {
		t.Errorf("line 1 = %q, want %q", lines[1], "hello bold")
	}
	if !tree.Children[0].Children[1].Style.Bold {
		t.Errorf("strong should be bold via UA")
	}
}

// B4g / C2 fidelity: the full text-level vocabulary flows and wraps
// through a UA-styled paragraph.
func TestMixedVocabularyWrapsWithStyles(t *testing.T) {
	// Given: <p>see <em>the</em> <mark>marked</mark> word</p> at width 12.
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "p", Kind: KindBox, Children: []*Input{
			{Kind: KindText, Content: "see "},
			{Tag: "em", Kind: KindText, Content: "the"},
			{Kind: KindText, Content: " "},
			{Tag: "mark", Kind: KindText, Content: "marked"},
			{Kind: KindText, Content: " word"},
		}},
	}}

	// When
	ResolveStyles(tree, nil, 12, 6)
	box := Layout(tree, 12, 6)
	buf := render.NewBuffer(12, 6)
	RenderTree(buf, box, nil)

	// Then: "see the" / "marked word" (rows 1-2 after p's UA margin).
	rows := []string{"", ""}
	for r := 0; r < 2; r++ {
		for c := 0; c < 12; c++ {
			ch := buf.Cell(1+r, c).Ch
			if ch == 0 {
				ch = ' '
			}
			rows[r] += string(ch)
		}
	}
	if got := rows[0][:7]; got != "see the" {
		t.Errorf("row 1 = %q, want %q", got, "see the")
	}
	if got := rows[1][:11]; got != "marked word" {
		t.Errorf("row 2 = %q, want %q", got, "marked word")
	}
	if !buf.Cell(1, 4).Style.Italic {
		t.Errorf("em should render italic: %+v", buf.Cell(1, 4).Style)
	}
	if buf.Cell(2, 0).Style.BG.Name != "yellow" {
		t.Errorf("mark should be on yellow: %+v", buf.Cell(2, 0).Style)
	}
	if buf.Cell(2, 7).Style.BG.Name == "yellow" {
		t.Errorf("plain word must not carry mark background")
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
