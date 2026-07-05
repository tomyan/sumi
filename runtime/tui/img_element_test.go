package tui_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// writeTestPNG writes a 2x4 image: left column red over blue (repeated),
// right column green over transparent-bottom rows.
func writeTestPNG(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 4))
	for y := 0; y < 4; y++ {
		if y%2 == 0 {
			img.Set(0, y, color.RGBA{255, 0, 0, 255}) // red top halves
			img.Set(1, y, color.RGBA{0, 255, 0, 255}) // green top halves
		} else {
			img.Set(0, y, color.RGBA{0, 0, 255, 255}) // blue bottom halves
			img.Set(1, y, color.RGBA{0, 0, 0, 0})     // transparent bottoms
		}
	}
	path := filepath.Join(t.TempDir(), "test.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImgRendersHalfBlocks(t *testing.T) {
	// Given
	src := writeTestPNG(t)
	img := &layout.Input{Kind: layout.KindBox, Tag: "img",
		Attrs: map[string]string{"src": src}, CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{img},
	}}

	// When
	tui.TestApp(comp, 10, 4)

	// Then — 2x4 pixels = 2 cols × 2 rows of half-blocks
	if img.FixedWidth != 2 || img.FixedHeight != 2 {
		t.Fatalf("intrinsic size = %dx%d cells, want 2x2", img.FixedWidth, img.FixedHeight)
	}
	if img.Cells == nil {
		t.Fatal("img has no cells")
	}
	// Left column: red over blue → '▀' fg red, bg blue.
	c := img.Cells.Cell(0, 0)
	if c.Ch != '▀' || !c.Style.FG.IsRGB || c.Style.FG.R != 255 || !c.Style.BG.IsRGB || c.Style.BG.B != 255 {
		t.Errorf("cell(0,0) = %c fg %+v bg %+v, want red-over-blue ▀", c.Ch, c.Style.FG, c.Style.BG)
	}
	// Right column: green over transparent → '▀' fg green, no bg.
	c = img.Cells.Cell(0, 1)
	if c.Ch != '▀' || c.Style.FG.G != 255 || c.Style.BG.IsRGB {
		t.Errorf("cell(0,1) = %c fg %+v bg %+v, want green over terminal bg", c.Ch, c.Style.FG, c.Style.BG)
	}
}

func TestImgWidthAttrScales(t *testing.T) {
	// Given — the 2x4 image scaled up to 4 cells wide
	src := writeTestPNG(t)
	img := &layout.Input{Kind: layout.KindBox, Tag: "img",
		Attrs: map[string]string{"src": src, "width": "4", "height": "2"},
		FixedWidth: 4, FixedHeight: 2, CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{img},
	}}

	// When
	tui.TestApp(comp, 10, 4)

	// Then — cells match the attr size; left half still red-topped
	if img.Cells == nil || img.Cells.Width() != 4 || img.Cells.Height() != 2 {
		t.Fatalf("cells = %v, want 4x2 grid", img.Cells)
	}
	if c := img.Cells.Cell(0, 1); c.Style.FG.R != 255 {
		t.Errorf("scaled cell(0,1) fg = %+v, want red", c.Style.FG)
	}
}

func TestImgMissingFileRendersNothing(t *testing.T) {
	// Given
	img := &layout.Input{Kind: layout.KindBox, Tag: "img",
		Attrs: map[string]string{"src": "/nonexistent.png"}, CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{img},
	}}

	// When / Then — no panic, no cells
	tui.TestApp(comp, 10, 4)
	if img.Cells != nil {
		t.Error("missing image should render nothing")
	}
}
