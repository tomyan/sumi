package render

import (
	"strings"
	"testing"
)

// A8: colour degradation at SGR emission time.

func withDepth(t *testing.T, d ColorDepth, fn func()) {
	t.Helper()
	prev := GetColorDepth()
	SetColorDepth(d)
	defer SetColorDepth(prev)
	fn()
}

func sgrFor(s Style) string {
	return string(appendSGR(nil, s))
}

func TestTrueColorEmitsRGB(t *testing.T) {
	withDepth(t, DepthTrueColor, func() {
		got := sgrFor(Style{FG: Color{IsRGB: true, R: 255, G: 85, B: 0}})
		if !strings.Contains(got, "38;2;255;85;0") {
			t.Errorf("sgr = %q, want 38;2;255;85;0", got)
		}
	})
}

func TestDepth256QuantizesRGB(t *testing.T) {
	withDepth(t, Depth256, func() {
		got := sgrFor(Style{FG: Color{IsRGB: true, R: 255, G: 0, B: 0}})
		if !strings.Contains(got, "38;5;196") {
			t.Errorf("sgr = %q, want 38;5;196 (pure red cube entry)", got)
		}
		gray := sgrFor(Style{BG: Color{IsRGB: true, R: 128, G: 128, B: 128}})
		if !strings.Contains(gray, "48;5;2") {
			t.Errorf("sgr = %q, want a grayscale-ramp background", gray)
		}
	})
}

func TestDepth16MapsToNearestANSI(t *testing.T) {
	withDepth(t, Depth16, func() {
		got := sgrFor(Style{FG: Color{IsRGB: true, R: 250, G: 30, B: 20}})
		if !strings.Contains(got, ";31") {
			t.Errorf("sgr = %q, want ;31 (red)", got)
		}
	})
}

func TestDepthMonoDropsColorsKeepsAttrs(t *testing.T) {
	withDepth(t, DepthMono, func() {
		got := sgrFor(Style{Bold: true, FG: Color{IsRGB: true, R: 255, G: 0, B: 0}})
		if strings.Contains(got, "38;") || strings.Contains(got, ";31") {
			t.Errorf("sgr = %q, colours must be dropped in mono", got)
		}
		if !strings.Contains(got, ";1") {
			t.Errorf("sgr = %q, bold must survive mono", got)
		}
	})
}

func TestNamedColorsPassThroughAt16(t *testing.T) {
	withDepth(t, Depth16, func() {
		got := sgrFor(Style{FG: Color{Name: "cyan"}})
		if !strings.Contains(got, ";36") {
			t.Errorf("sgr = %q, want ;36", got)
		}
	})
}

func TestSetColorDepthIgnoresAuto(t *testing.T) {
	withDepth(t, Depth256, func() {
		SetColorDepth(DepthAuto)
		if GetColorDepth() != Depth256 {
			t.Error("DepthAuto must not change the active depth")
		}
	})
}

// A9: light-dark pairs resolve against the active scheme at emission time.
func TestLightDarkPairResolvesPerScheme(t *testing.T) {
	pair := Color{Pair: &ColorPair{
		Light: Color{IsRGB: true, R: 255, G: 255, B: 255},
		Dark:  Color{IsRGB: true, R: 0, G: 0, B: 0},
	}}
	withDepth(t, DepthTrueColor, func() {
		prev := GetColorScheme()
		defer SetColorScheme(prev)

		SetColorScheme(SchemeDark)
		if got := sgrFor(Style{FG: pair}); !strings.Contains(got, "38;2;0;0;0") {
			t.Errorf("dark scheme sgr = %q, want black", got)
		}
		SetColorScheme(SchemeLight)
		if got := sgrFor(Style{FG: pair}); !strings.Contains(got, "38;2;255;255;255") {
			t.Errorf("light scheme sgr = %q, want white", got)
		}
	})
}
