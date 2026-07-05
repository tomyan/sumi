package css

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// A7: CSS colour value parsing.

func rgbNear(t *testing.T, got render.Color, r, g, b int, label string) {
	t.Helper()
	if !got.IsRGB {
		t.Fatalf("%s: not RGB: %+v", label, got)
	}
	const tol = 3
	for _, ch := range []struct {
		name string
		got  uint8
		want int
	}{{"R", got.R, r}, {"G", got.G, g}, {"B", got.B, b}} {
		diff := int(ch.got) - ch.want
		if diff < -tol || diff > tol {
			t.Errorf("%s: %s = %d, want ~%d (got %+v)", label, ch.name, ch.got, ch.want, got)
		}
	}
}

func mustColor(t *testing.T, v string) render.Color {
	t.Helper()
	c, ok := ParseColorValue(v)
	if !ok {
		t.Fatalf("ParseColorValue(%q) failed", v)
	}
	return c
}

func TestParseAnsiKeywordsStayNamed(t *testing.T) {
	c := mustColor(t, "cyan")
	if c.IsRGB || c.Name != "cyan" {
		t.Errorf("cyan = %+v, want palette name", c)
	}
}

func TestParseNamedColor(t *testing.T) {
	rgbNear(t, mustColor(t, "rebeccapurple"), 0x66, 0x33, 0x99, "rebeccapurple")
	rgbNear(t, mustColor(t, "tomato"), 0xff, 0x63, 0x47, "tomato")
}

func TestParseHexForms(t *testing.T) {
	rgbNear(t, mustColor(t, "#ff5500"), 255, 85, 0, "#rrggbb")
	rgbNear(t, mustColor(t, "#f50"), 255, 85, 0, "#rgb")
	rgbNear(t, mustColor(t, "#ff550080"), 255, 85, 0, "#rrggbbaa (alpha dropped)")
	rgbNear(t, mustColor(t, "#f508"), 255, 85, 0, "#rgba (alpha dropped)")
}

func TestParseRgbFunctions(t *testing.T) {
	rgbNear(t, mustColor(t, "rgb(255, 85, 0)"), 255, 85, 0, "legacy rgb")
	rgbNear(t, mustColor(t, "rgb(255 85 0)"), 255, 85, 0, "modern rgb")
	rgbNear(t, mustColor(t, "rgba(255, 85, 0, 0.5)"), 255, 85, 0, "rgba alpha dropped")
	rgbNear(t, mustColor(t, "rgb(100% 0% 50%)"), 255, 0, 128, "percentage rgb")
	rgbNear(t, mustColor(t, "rgb(255 85 0 / 0.5)"), 255, 85, 0, "slash alpha")
}

func TestParseHslFunctions(t *testing.T) {
	rgbNear(t, mustColor(t, "hsl(0, 100%, 50%)"), 255, 0, 0, "red hsl")
	rgbNear(t, mustColor(t, "hsl(120 100% 25%)"), 0, 128, 0, "green hsl")
	rgbNear(t, mustColor(t, "hsl(240deg 100% 50%)"), 0, 0, 255, "blue hsl deg")
	rgbNear(t, mustColor(t, "hsl(0.5turn 100% 50%)"), 0, 255, 255, "turn hue")
}

func TestParseHwb(t *testing.T) {
	rgbNear(t, mustColor(t, "hwb(0 0% 0%)"), 255, 0, 0, "pure red hwb")
	rgbNear(t, mustColor(t, "hwb(0 100% 0%)"), 255, 255, 255, "all white")
	rgbNear(t, mustColor(t, "hwb(0 0% 100%)"), 0, 0, 0, "all black")
}

func TestParseLab(t *testing.T) {
	// CSS red #ff0000 ≈ lab(54.29 80.81 69.89)
	rgbNear(t, mustColor(t, "lab(54.29 80.81 69.89)"), 255, 0, 0, "lab red")
	rgbNear(t, mustColor(t, "lch(54.29 106.84 40.86)"), 255, 0, 0, "lch red")
}

func TestParseOklab(t *testing.T) {
	// #ff0000 ≈ oklab(0.628 0.2249 0.1258) ≈ oklch(0.628 0.2577 29.23)
	rgbNear(t, mustColor(t, "oklab(0.628 0.2249 0.1258)"), 255, 0, 0, "oklab red")
	rgbNear(t, mustColor(t, "oklch(0.628 0.2577 29.23)"), 255, 0, 0, "oklch red")
	rgbNear(t, mustColor(t, "oklch(70% 0.1 200)"), 64, 177, 183, "oklch teal-ish")
}

func TestParseTransparent(t *testing.T) {
	c, ok := ParseColorValue("transparent")
	if !ok || c.IsRGB || c.Name != "" {
		t.Errorf("transparent = %+v ok=%v, want zero colour", c, ok)
	}
}

func TestParseInvalidColorsDrop(t *testing.T) {
	for _, v := range []string{"", "notacolor", "#12", "#12345", "rgb(a b c)", "blah(1 2 3)"} {
		if _, ok := ParseColorValue(v); ok {
			t.Errorf("ParseColorValue(%q) should fail", v)
		}
	}
}

// Integration with ToRenderStyle.

func TestToRenderStyleModernColorValues(t *testing.T) {
	s := ToRenderStyle(map[string]string{
		"color":      "rebeccapurple",
		"background": "hsl(120 100% 25%)",
	})
	rgbNear(t, s.FG, 0x66, 0x33, 0x99, "FG named")
	rgbNear(t, s.BG, 0, 128, 0, "BG hsl")
}

func TestToRenderStyleShortHex(t *testing.T) {
	s := ToRenderStyle(map[string]string{"color": "#f50"})
	rgbNear(t, s.FG, 255, 85, 0, "short hex FG")
}

func TestToRenderStyleCurrentColorBorder(t *testing.T) {
	// border-color: currentColor leaves FG as the color value.
	s := ToRenderStyle(map[string]string{"color": "red", "border-color": "currentColor"})
	if s.FG.Name != "red" {
		t.Errorf("FG = %+v, want red (currentColor keeps color)", s.FG)
	}
}

func TestToRenderStyleInvalidColorDropped(t *testing.T) {
	s := ToRenderStyle(map[string]string{"color": "bogus(1)"})
	if !s.IsZero() {
		t.Errorf("invalid colour must drop, got %+v", s)
	}
}

func TestParseLightDark(t *testing.T) {
	c := mustColor(t, "light-dark(#fff, rgb(0 0 0))")
	if c.Pair == nil {
		t.Fatalf("light-dark should produce a Pair, got %+v", c)
	}
	rgbNear(t, c.Pair.Light, 255, 255, 255, "light arm")
	rgbNear(t, c.Pair.Dark, 0, 0, 0, "dark arm")
}

func TestParseLightDarkInvalidArm(t *testing.T) {
	if _, ok := ParseColorValue("light-dark(bogus, #000)"); ok {
		t.Error("invalid arm should fail the whole value")
	}
}
