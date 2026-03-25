package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/vt100"
)

func TestSGR24BitForeground(t *testing.T) {
	// Given — ESC[38;2;255;100;50m sets 24-bit FG color
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[38;2;255;100;50mX"))

	// Then
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Fatalf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
	fg := cell.Style.FG
	if !fg.IsRGB {
		t.Fatal("FG.IsRGB = false, want true")
	}
	if fg.R != 255 || fg.G != 100 || fg.B != 50 {
		t.Errorf("FG RGB = (%d,%d,%d), want (255,100,50)", fg.R, fg.G, fg.B)
	}
}

func TestSGR24BitBackground(t *testing.T) {
	// Given — ESC[48;2;20;22;27m sets 24-bit BG color
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[48;2;20;22;27mX"))

	// Then
	cell := screen.Cell(0, 0)
	bg := cell.Style.BG
	if !bg.IsRGB {
		t.Fatal("BG.IsRGB = false, want true")
	}
	if bg.R != 20 || bg.G != 22 || bg.B != 27 {
		t.Errorf("BG RGB = (%d,%d,%d), want (20,22,27)", bg.R, bg.G, bg.B)
	}
}

func TestSGR256ColorForeground(t *testing.T) {
	// Given — ESC[38;5;196m sets 256-color FG (196 = bright red = #ff0000)
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[38;5;196mX"))

	// Then — should store as RGB
	cell := screen.Cell(0, 0)
	fg := cell.Style.FG
	if !fg.IsRGB {
		t.Fatal("FG.IsRGB = false, want true")
	}
	if fg.R != 0xff || fg.G != 0x00 || fg.B != 0x00 {
		t.Errorf("FG RGB = (%d,%d,%d), want (255,0,0)", fg.R, fg.G, fg.B)
	}
}

func TestSGR256ColorBackground(t *testing.T) {
	// Given — ESC[48;5;21m sets 256-color BG (21 = blue = #0000ff)
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[48;5;21mX"))

	// Then
	cell := screen.Cell(0, 0)
	bg := cell.Style.BG
	if !bg.IsRGB {
		t.Fatal("BG.IsRGB = false, want true")
	}
	if bg.R != 0x00 || bg.G != 0x00 || bg.B != 0xff {
		t.Errorf("BG RGB = (%d,%d,%d), want (0,0,255)", bg.R, bg.G, bg.B)
	}
}

func TestSGRResetClearsRGB(t *testing.T) {
	// Given — set an RGB color then reset
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[38;2;255;0;0m\x1b[0mX"))

	// Then — style should be zero
	cell := screen.Cell(0, 0)
	if cell.Style.FG.IsRGB {
		t.Error("FG.IsRGB = true after reset, want false")
	}
}

func TestSGRCombinedRGBAndBold(t *testing.T) {
	// Given — ESC[1m then ESC[38;2;100;200;150m
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[1m\x1b[38;2;100;200;150mX"))

	// Then — both bold and RGB FG
	cell := screen.Cell(0, 0)
	if !cell.Style.Bold {
		t.Error("Bold = false, want true")
	}
	if !cell.Style.FG.IsRGB || cell.Style.FG.R != 100 || cell.Style.FG.G != 200 || cell.Style.FG.B != 150 {
		t.Errorf("FG = %+v, want RGB(100,200,150)", cell.Style.FG)
	}
}

func TestRGBColorRenderRoundTrip(t *testing.T) {
	// Given — a buffer with an RGB-styled cell
	buf := render.NewBuffer(5, 1)
	style := render.Style{FG: render.Color{IsRGB: true, R: 128, G: 64, B: 32}}
	buf.SetStyledCell(0, 0, 'A', style)

	// When — render to a new screen
	var out []byte
	screen := vt100.NewScreen(5, 1)
	buf.RenderToOffset(&bytesWriter{&out}, 0, 0)
	screen.Write(out)

	// Then — the cell should have the same RGB color
	cell := screen.Cell(0, 0)
	if cell.Ch != 'A' {
		t.Fatalf("Ch = %c, want 'A'", cell.Ch)
	}
	if !cell.Style.FG.IsRGB || cell.Style.FG.R != 128 || cell.Style.FG.G != 64 || cell.Style.FG.B != 32 {
		t.Errorf("FG = %+v, want RGB(128,64,32)", cell.Style.FG)
	}
}

// bytesWriter implements io.Writer appending to a byte slice.
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
