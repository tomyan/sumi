package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestRegionDispatchesResizeAndRendersFedCells(t *testing.T) {
	// Given — a region whose resize handler feeds a cell grid
	var sizes [][2]int
	region := &layout.Input{Kind: layout.KindBox, Tag: "region",
		FixedWidth: 6, FixedHeight: 2, CursorCol: -1, CursorRow: -1}
	region.On = map[string]func(*layout.DOMEvent){
		"resize": func(evt *layout.DOMEvent) {
			w := evt.Data["width"].(int)
			h := evt.Data["height"].(int)
			sizes = append(sizes, [2]int{w, h})
			cells := render.NewBuffer(w, h)
			cells.SetStyledCell(0, 0, '#', render.Style{FG: render.Color{Name: "green"}})
			evt.Target.Cells = cells
		},
	}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{region},
	}}

	// When
	app := tui.TestApp(comp, 20, 4)

	// Then — one resize with the content size; fed cells render
	if len(sizes) != 1 || sizes[0] != [2]int{6, 2} {
		t.Fatalf("resize sizes = %v, want one 6x2", sizes)
	}
	if got := app.TestBuffer.Cell(0, 0); got.Ch != '#' || got.Style.FG.Name != "green" {
		t.Errorf("frame cell(0,0) = %c %q, want green #", got.Ch, got.Style.FG.Name)
	}

	// When — further renders without a size change stay quiet
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then
	if len(sizes) != 1 {
		t.Errorf("resize fired %d times after no size change, want 1", len(sizes))
	}
}
