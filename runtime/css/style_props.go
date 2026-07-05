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
	if v, ok := props["opacity"]; ok && opacityIsDim(v) {
		s.Dim = true
	}
	if props["inverse"] == "true" {
		s.Inverse = true
	}
	return s
}

func applyColorProps(s *render.Style, props map[string]string) {
	if v, ok := props["color"]; ok {
		s.FG = parseColor(v)
	}
	if v, ok := props["border-color"]; ok {
		s.FG = parseColor(v)
	}
	if v, ok := props["background"]; ok {
		s.BG = parseColor(v)
	}
	if v, ok := props["background-color"]; ok {
		s.BG = parseColor(v)
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

// opacityIsDim reports whether an opacity value maps to the terminal dim
// attribute: the non-standard keyword `dim`, or any numeric value below 1.
func opacityIsDim(v string) bool {
	if v == "dim" {
		return true
	}
	f, err := strconv.ParseFloat(v, 64)
	return err == nil && f < 1
}

// parseColor parses a CSS colour value — either a named colour ("red") or hex ("#ff5555").
func parseColor(v string) render.Color {
	if len(v) == 7 && v[0] == '#' {
		r := hexByte(v[1], v[2])
		g := hexByte(v[3], v[4])
		b := hexByte(v[5], v[6])
		return render.Color{IsRGB: true, R: r, G: g, B: b}
	}
	return render.Color{Name: v}
}

func hexByte(hi, lo byte) uint8 {
	return hexNibble(hi)<<4 | hexNibble(lo)
}

func hexNibble(b byte) uint8 {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	default:
		return 0
	}
}
