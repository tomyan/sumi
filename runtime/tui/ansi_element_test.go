package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestAnsiElementParsesSGRIntoCells(t *testing.T) {
	// Given — an ansi element whose body carries raw SGR sequences
	source := &layout.Input{Kind: layout.KindText, Tag: "text",
		Content: "\x1b[31mred\x1b[0m ok\nline2"}
	ansi := &layout.Input{Kind: layout.KindBox, Tag: "ansi",
		CursorCol: -1, CursorRow: -1, Children: []*layout.Input{source}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{ansi},
	}}

	// When
	app := tui.TestApp(comp, 20, 4)

	// Then — sized to the visible content, source hidden, cells styled
	if ansi.FixedWidth != 6 || ansi.FixedHeight != 2 {
		t.Errorf("size = %dx%d, want 6x2", ansi.FixedWidth, ansi.FixedHeight)
	}
	if source.Display != "none" {
		t.Error("raw source child should be hidden")
	}
	if ansi.Cells == nil {
		t.Fatal("ansi element has no cells")
	}
	if c := ansi.Cells.Cell(0, 0); c.Ch != 'r' || c.Style.FG.Name != "red" {
		t.Errorf("cell(0,0) = %c fg %q, want red r", c.Ch, c.Style.FG.Name)
	}
	if c := ansi.Cells.Cell(0, 4); c.Ch != 'o' || c.Style.FG.Name == "red" {
		t.Errorf("cell(0,4) = %c fg %q, want unstyled o", c.Ch, c.Style.FG.Name)
	}

	// And the frame shows the text at the element's position
	if got := app.TestBuffer.Cell(0, 0).Ch; got != 'r' {
		t.Errorf("frame cell(0,0) = %c, want r", got)
	}
	if got := app.TestBuffer.Cell(1, 0).Ch; got != 'l' {
		t.Errorf("frame cell(1,0) = %c, want l", got)
	}
}
