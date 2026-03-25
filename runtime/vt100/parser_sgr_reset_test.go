package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestSGRBoldOff(t *testing.T) {
	// Given — bold then bold off
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[1mA\x1b[22mB"))

	// Then — A bold, B not bold
	if !screen.Cell(0, 0).Style.Bold {
		t.Error("A should be bold")
	}
	if screen.Cell(0, 1).Style.Bold {
		t.Error("B should not be bold")
	}
	if screen.Cell(0, 1).Style.Dim {
		t.Error("B should not be dim")
	}
}

func TestSGRItalicOff(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[3mA\x1b[23mB"))

	if !screen.Cell(0, 0).Style.Italic {
		t.Error("A should be italic")
	}
	if screen.Cell(0, 1).Style.Italic {
		t.Error("B should not be italic")
	}
}

func TestSGRUnderlineOff(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[4mA\x1b[24mB"))

	if !screen.Cell(0, 0).Style.Underline {
		t.Error("A should be underlined")
	}
	if screen.Cell(0, 1).Style.Underline {
		t.Error("B should not be underlined")
	}
}

func TestSGRInverseOff(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[7mA\x1b[27mB"))

	if !screen.Cell(0, 0).Style.Inverse {
		t.Error("A should be inverse")
	}
	if screen.Cell(0, 1).Style.Inverse {
		t.Error("B should not be inverse")
	}
}

func TestSGRStrikethroughOff(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[9mA\x1b[29mB"))

	if !screen.Cell(0, 0).Style.Strikethrough {
		t.Error("A should be strikethrough")
	}
	if screen.Cell(0, 1).Style.Strikethrough {
		t.Error("B should not be strikethrough")
	}
}

func TestSGRDefaultForeground(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[31mA\x1b[39mB"))

	// Then — A has red FG, B has default FG
	if screen.Cell(0, 0).Style.FG.Name != "red" {
		t.Errorf("A FG = %q, want red", screen.Cell(0, 0).Style.FG.Name)
	}
	b := screen.Cell(0, 1).Style.FG
	if b.Name != "" || b.IsRGB {
		t.Errorf("B FG = %+v, want default", b)
	}
}

func TestSGRDefaultBackground(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[41mA\x1b[49mB"))

	if screen.Cell(0, 0).Style.BG.Name != "red" {
		t.Errorf("A BG = %q, want red", screen.Cell(0, 0).Style.BG.Name)
	}
	b := screen.Cell(0, 1).Style.BG
	if b.Name != "" || b.IsRGB {
		t.Errorf("B BG = %+v, want default", b)
	}
}

func TestSGRDefaultFGClearsRGB(t *testing.T) {
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[38;2;255;0;0mA\x1b[39mB"))

	if !screen.Cell(0, 0).Style.FG.IsRGB {
		t.Error("A FG should be RGB")
	}
	b := screen.Cell(0, 1).Style.FG
	if b.IsRGB || b.Name != "" {
		t.Errorf("B FG = %+v, want default", b)
	}
}
