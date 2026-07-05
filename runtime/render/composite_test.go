package render

import "testing"

func rgba(r, g, b, a uint8) Color {
	return Color{IsRGB: true, R: r, G: g, B: b, A: a}
}

func TestAlphaBackgroundBlendsOverExistingCell(t *testing.T) {
	// Given — a cell painted with an opaque blue background
	buf := NewBuffer(3, 1)
	buf.SetStyledCell(0, 0, ' ', Style{BG: rgba(0, 0, 255, 0)})

	// When — a 50% red background paints over it
	buf.SetStyledCell(0, 0, 'x', Style{BG: rgba(255, 0, 0, 128)})

	// Then — the stored background is the blend, marked opaque
	got := buf.Cell(0, 0).Style.BG
	if !got.IsRGB || got.A != 0 {
		t.Fatalf("BG = %+v, want opaque RGB blend", got)
	}
	if got.R < 120 || got.R > 135 || got.B < 120 || got.B > 135 || got.G != 0 {
		t.Errorf("BG = (%d,%d,%d), want ~ (128,0,127)", got.R, got.G, got.B)
	}
}

func TestAlphaForegroundBlendsOverBackdrop(t *testing.T) {
	// Given — an opaque black background under the cell
	buf := NewBuffer(3, 1)
	buf.SetStyledCell(0, 0, ' ', Style{BG: rgba(0, 0, 0, 0)})

	// When — white text at 50% alpha, no background of its own
	buf.SetStyledCell(0, 0, 'x', Style{FG: rgba(255, 255, 255, 128)})

	// Then — grey text
	got := buf.Cell(0, 0).Style.FG
	if got.A != 0 || got.R < 120 || got.R > 135 {
		t.Errorf("FG = %+v, want ~50%% grey, opaque", got)
	}
}

func TestAlphaOverNonRGBBackdropFallsBackToOpaque(t *testing.T) {
	// Given — a named-colour backdrop the compositor can't blend with
	buf := NewBuffer(3, 1)
	buf.SetStyledCell(0, 0, ' ', Style{BG: Color{Name: "blue"}})

	// When
	buf.SetStyledCell(0, 0, 'x', Style{BG: rgba(255, 0, 0, 128)})

	// Then — the source paints opaque
	got := buf.Cell(0, 0).Style.BG
	if got.A != 0 || got.R != 255 || got.B != 0 {
		t.Errorf("BG = %+v, want opaque red fallback", got)
	}
}

func TestOpaqueWritesAreUntouched(t *testing.T) {
	// Given / When — no alpha anywhere
	buf := NewBuffer(3, 1)
	buf.SetStyledCell(0, 0, 'x', Style{FG: Color{Name: "red"}, BG: rgba(0, 255, 0, 0)})

	// Then
	c := buf.Cell(0, 0)
	if c.Style.FG.Name != "red" || c.Style.BG.G != 255 {
		t.Errorf("style = %+v, want unchanged", c.Style)
	}
}
