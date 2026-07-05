package css

import "testing"

// A1: standard CSS property names replace the legacy boolean shorthands.

func TestFontWeightBoldSetsBold(t *testing.T) {
	// Given / When
	s := ToRenderStyle(map[string]string{"font-weight": "bold"})

	// Then
	if !s.Bold {
		t.Error("font-weight: bold should set Bold")
	}
}

func TestFontWeightNumericThresholdSetsBold(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"700", true},
		{"800", true},
		{"bolder", true},
		{"600", false},
		{"400", false},
		{"normal", false},
	}
	for _, c := range cases {
		s := ToRenderStyle(map[string]string{"font-weight": c.value})
		if s.Bold != c.want {
			t.Errorf("font-weight: %s → Bold = %v, want %v", c.value, s.Bold, c.want)
		}
	}
}

func TestFontStyleItalicSetsItalic(t *testing.T) {
	for _, v := range []string{"italic", "oblique"} {
		s := ToRenderStyle(map[string]string{"font-style": v})
		if !s.Italic {
			t.Errorf("font-style: %s should set Italic", v)
		}
	}
	if ToRenderStyle(map[string]string{"font-style": "normal"}).Italic {
		t.Error("font-style: normal should not set Italic")
	}
}

func TestTextDecorationUnderline(t *testing.T) {
	// Given / When
	s := ToRenderStyle(map[string]string{"text-decoration": "underline"})

	// Then
	if !s.Underline || s.Strikethrough {
		t.Errorf("text-decoration: underline → Underline=%v Strikethrough=%v", s.Underline, s.Strikethrough)
	}
}

func TestTextDecorationLineThrough(t *testing.T) {
	s := ToRenderStyle(map[string]string{"text-decoration": "line-through"})
	if s.Underline || !s.Strikethrough {
		t.Errorf("text-decoration: line-through → Underline=%v Strikethrough=%v", s.Underline, s.Strikethrough)
	}
}

func TestTextDecorationCombined(t *testing.T) {
	s := ToRenderStyle(map[string]string{"text-decoration": "underline line-through"})
	if !s.Underline || !s.Strikethrough {
		t.Errorf("combined text-decoration → Underline=%v Strikethrough=%v", s.Underline, s.Strikethrough)
	}
}

func TestTextDecorationNone(t *testing.T) {
	s := ToRenderStyle(map[string]string{"text-decoration": "none"})
	if s.Underline || s.Strikethrough {
		t.Error("text-decoration: none should set nothing")
	}
}

func TestOpacityDimKeyword(t *testing.T) {
	s := ToRenderStyle(map[string]string{"opacity": "dim"})
	if !s.Dim {
		t.Error("opacity: dim should set Dim")
	}
}

func TestOpacityBelowOneSetsDim(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"0.5", true},
		{"0", true},
		{"1", false},
		{"1.0", false},
	}
	for _, c := range cases {
		s := ToRenderStyle(map[string]string{"opacity": c.value})
		if s.Dim != c.want {
			t.Errorf("opacity: %s → Dim = %v, want %v", c.value, s.Dim, c.want)
		}
	}
}

func TestBackgroundColorAlias(t *testing.T) {
	s := ToRenderStyle(map[string]string{"background-color": "cyan"})
	if s.BG.Name != "cyan" {
		t.Errorf("background-color should set BG, got %+v", s.BG)
	}
}

func TestInverseTerminalExtensionRetained(t *testing.T) {
	s := ToRenderStyle(map[string]string{"inverse": "true"})
	if !s.Inverse {
		t.Error("inverse: true (terminal extension) should set Inverse")
	}
}

// A1 clean break: the legacy boolean names must no longer be honoured.
func TestLegacyBooleanNamesDropped(t *testing.T) {
	s := ToRenderStyle(map[string]string{
		"bold":          "true",
		"dim":           "true",
		"italic":        "true",
		"underline":     "true",
		"strikethrough": "true",
	})
	if !s.IsZero() {
		t.Errorf("legacy boolean properties must be ignored, got %+v", s)
	}
}

func TestToRenderStyleColorAndBackground(t *testing.T) {
	s := ToRenderStyle(map[string]string{"color": "red", "background": "blue"})
	if s.FG.Name != "red" || s.BG.Name != "blue" {
		t.Errorf("FG=%+v BG=%+v", s.FG, s.BG)
	}
}

func TestToRenderStyleBorderColorOverridesColor(t *testing.T) {
	s := ToRenderStyle(map[string]string{"color": "red", "border-color": "cyan"})
	if s.FG.Name != "cyan" {
		t.Errorf("FG.Name = %q, want cyan (border-color overrides color)", s.FG.Name)
	}
}

func TestToRenderStyleCombinedProperties(t *testing.T) {
	s := ToRenderStyle(map[string]string{
		"color":           "cyan",
		"background":      "black",
		"font-weight":     "bold",
		"text-decoration": "underline",
	})
	if s.FG.Name != "cyan" || s.BG.Name != "black" || !s.Bold || !s.Underline {
		t.Errorf("style = %+v", s)
	}
}

func TestToRenderStyleEmptyPropertiesIsZero(t *testing.T) {
	if s := ToRenderStyle(map[string]string{}); !s.IsZero() {
		t.Errorf("empty props should give zero style, got %+v", s)
	}
}

// A2: unknown and pixel-derived properties drop silently.
func TestUnknownPropertiesDropSilently(t *testing.T) {
	s := ToRenderStyle(map[string]string{
		"font-size":     "14px",
		"font-family":   "monospace",
		"box-shadow":    "0 2px 8px #0004",
		"border-radius": "4px",
		"transform":     "scale(1.1)",
	})
	if !s.IsZero() {
		t.Errorf("unknown/pixel properties must produce zero style, got %+v", s)
	}
}

// F2b: numeric opacity becomes a blend factor on the element's colours;
// the dim keyword and non-RGB fallbacks keep the terminal dim attribute.
func TestOpacityNumericBecomesAlpha(t *testing.T) {
	// Given / When — RGB colours pick up alpha
	s := ToRenderStyle(map[string]string{
		"color": "#ff0000", "background": "#0000ff", "opacity": "0.5",
	})

	// Then
	if s.FG.A != 128 || s.BG.A != 128 {
		t.Errorf("alpha = fg %d bg %d, want 128/128", s.FG.A, s.BG.A)
	}
	if s.Dim {
		t.Error("numeric opacity on RGB colours should not dim")
	}

	// When — named colour can't blend: falls back to dim
	s = ToRenderStyle(map[string]string{"color": "red", "opacity": "0.5"})

	// Then
	if !s.Dim {
		t.Error("non-RGB colour with numeric opacity should fall back to dim")
	}

	// When — the dim keyword stays dim
	s = ToRenderStyle(map[string]string{"opacity": "dim"})
	if !s.Dim {
		t.Error("opacity: dim should set Dim")
	}
}
