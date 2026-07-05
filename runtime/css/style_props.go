package css

import (
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// ToRenderStyle converts resolved CSS properties to a render.Style.
// Standard CSS property names are honoured; unknown properties drop silently.
// Terminal extension: `inverse: true` (no standard CSS equivalent on a cell grid).
func ToRenderStyle(props map[string]string) render.Style {
	var s render.Style
	applyColorProps(&s, props)
	applyTextProps(&s, props)
	applyOpacity(&s, props["opacity"])
	if props["inverse"] == "true" {
		s.Inverse = true
	}
	return s
}

func applyColorProps(s *render.Style, props map[string]string) {
	setColor(&s.FG, props["color"])
	// border-color: currentColor keeps the color value already in FG.
	if v := props["border-color"]; !strings.EqualFold(v, "currentColor") {
		setColor(&s.FG, v)
	}
	setColor(&s.BG, props["background"])
	setColor(&s.BG, props["background-color"])
}

// setColor parses a colour value into dst; unset or invalid values leave
// dst untouched (graceful drop).
func setColor(dst *render.Color, v string) {
	if v == "" {
		return
	}
	if c, ok := ParseColorValue(v); ok {
		*dst = c
	}
}

func applyTextProps(s *render.Style, props map[string]string) {
	if v, ok := props["font-weight"]; ok && fontWeightIsBold(v) {
		s.Bold = true
	}
	if v, ok := props["font-style"]; ok && (v == "italic" || v == "oblique") {
		s.Italic = true
	}
	if v, ok := props["text-decoration"]; ok {
		applyTextDecoration(s, v)
	}
}

func applyTextDecoration(s *render.Style, value string) {
	for _, part := range strings.Fields(value) {
		switch part {
		case "underline":
			s.Underline = true
		case "line-through":
			s.Strikethrough = true
		}
	}
}

// fontWeightIsBold reports whether a font-weight value means bold on a
// terminal: the keywords bold/bolder, or a numeric weight of 700 or more.
func fontWeightIsBold(v string) bool {
	if v == "bold" || v == "bolder" {
		return true
	}
	n, err := strconv.Atoi(v)
	return err == nil && n >= 700
}

// applyOpacity maps opacity onto the style. Numeric values below 1
// become a blend factor (alpha) on the element's RGB colours, composited
// at paint time; the dim keyword — and numeric opacity when the colours
// are not RGB and so cannot blend — uses the terminal dim attribute.
func applyOpacity(s *render.Style, v string) {
	if v == "" {
		return
	}
	if v == "dim" {
		s.Dim = true
		return
	}
	alpha, err := strconv.ParseFloat(v, 64)
	if err != nil || alpha >= 1 {
		return
	}
	blended := false
	if s.FG.IsRGB {
		s.FG = withAlpha(s.FG, alpha)
		blended = true
	}
	if s.BG.IsRGB {
		s.BG = withAlpha(s.BG, alpha)
		blended = true
	}
	if !blended {
		s.Dim = true
	}
}
