package vt100_test

import (
	"bytes"
	"testing"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/vt100"
)

func TestParseSGRBoldRoundTrip(t *testing.T) {
	// Given — buffer with bold text
	buf := render.NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Bold", render.Style{Bold: true})

	// When — render to ANSI and parse back
	screen := roundTrip(t, buf)

	// Then — cells have bold style
	assertFullCellsMatch(t, buf, screen)
}

func TestParseSGRColorRoundTrip(t *testing.T) {
	// Given — buffer with colored text
	buf := render.NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Red", render.Style{FG: render.Color{Name: "red"}})
	buf.WriteStyledText(0, 5, "BGBlue", render.Style{BG: render.Color{Name: "blue"}})

	// When — render to ANSI and parse back
	screen := roundTrip(t, buf)

	// Then — cells have correct colors
	assertFullCellsMatch(t, buf, screen)
}

func TestParseSGRMultipleAttrsRoundTrip(t *testing.T) {
	// Given — buffer with multiple style attributes
	buf := render.NewBuffer(15, 1)
	style := render.Style{
		Bold:   true,
		Italic: true,
		FG:     render.Color{Name: "green"},
	}
	buf.WriteStyledText(0, 0, "Fancy", style)

	// When — render to ANSI and parse back
	screen := roundTrip(t, buf)

	// Then — all style attributes preserved
	assertFullCellsMatch(t, buf, screen)
}

func TestParseSGRAllAttrsRoundTrip(t *testing.T) {
	// Given — buffer with every style attribute
	buf := render.NewBuffer(20, 1)
	style := render.Style{
		Bold:          true,
		Dim:           true,
		Italic:        true,
		Underline:     true,
		Inverse:       true,
		Strikethrough: true,
		FG:            render.Color{Name: "cyan"},
		BG:            render.Color{Name: "yellow"},
	}
	buf.WriteStyledText(0, 0, "All", style)

	// When — render to ANSI and parse back
	screen := roundTrip(t, buf)

	// Then — all attributes preserved
	assertFullCellsMatch(t, buf, screen)
}

func TestParseSGRResetBetweenStyles(t *testing.T) {
	// Given — buffer with styled text followed by plain text
	buf := render.NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Hi", render.Style{Bold: true})
	buf.WriteText(0, 2, "Lo")

	// When — render to ANSI and parse back
	screen := roundTrip(t, buf)

	// Then — 'H' and 'i' are bold, 'L' and 'o' are plain
	assertFullCellsMatch(t, buf, screen)
}

func TestParseSGRMultiParamSequence(t *testing.T) {
	// Given — a multi-param SGR sequence (not from sumi renderer, but valid ANSI)
	screen := vt100.NewScreen(5, 1)

	// When — ESC[0;1;32m sets reset+bold+green, then write text
	screen.Write([]byte("\x1b[1;1H\x1b[0;1;32mHi"))

	// Then — cells are bold + green
	cell := screen.Cell(0, 0)
	if cell.Ch != 'H' {
		t.Errorf("char: got %q, want 'H'", cell.Ch)
	}
	if !cell.Style.Bold {
		t.Error("expected bold")
	}
	if cell.Style.FG.Name != "green" {
		t.Errorf("fg: got %q, want green", cell.Style.FG.Name)
	}
}

func TestParseSGRMixedRowRoundTrip(t *testing.T) {
	// Given — a row with alternating styled and unstyled text
	buf := render.NewBuffer(20, 1)
	buf.WriteStyledText(0, 0, "RED", render.Style{FG: render.Color{Name: "red"}})
	buf.WriteText(0, 3, " ")
	buf.WriteStyledText(0, 4, "BLUE", render.Style{FG: render.Color{Name: "blue"}})

	// When
	screen := roundTrip(t, buf)

	// Then
	assertFullCellsMatch(t, buf, screen)
}

// roundTrip renders a buffer to ANSI and parses it back into a Screen.
func roundTrip(t *testing.T, buf *render.Buffer) *vt100.Screen {
	t.Helper()
	var ansi bytes.Buffer
	buf.RenderTo(&ansi)

	screen := vt100.NewScreen(buf.Width(), buf.Height())
	if _, err := screen.Write(ansi.Bytes()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return screen
}

// assertFullCellsMatch compares char and style for every cell.
func assertFullCellsMatch(t *testing.T, expected *render.Buffer, screen *vt100.Screen) {
	t.Helper()
	for row := 0; row < expected.Height(); row++ {
		for col := 0; col < expected.Width(); col++ {
			exp := expected.Cell(row, col)
			got := screen.Cell(row, col)
			if exp.Ch != got.Ch {
				t.Errorf("char mismatch at (%d,%d): got %q, want %q", row, col, got.Ch, exp.Ch)
			}
			if exp.Style != got.Style {
				t.Errorf("style mismatch at (%d,%d): got %+v, want %+v", row, col, got.Style, exp.Style)
			}
		}
	}
}
