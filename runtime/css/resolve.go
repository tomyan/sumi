package css

import (
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/render"
)

// Resolve finds all matching rules for an element and merges their properties.
// tag is the element type ("text", "box"), classes is the list of class names.
// Rules are matched in order; later properties override earlier ones.
func Resolve(stylesheet *style.Stylesheet, tag string, classes []string) map[string]string {
	props := make(map[string]string)
	for _, rule := range stylesheet.Rules {
		if matchesSelector(rule.Selector, tag, classes) {
			for k, v := range rule.Properties {
				props[k] = v
			}
		}
	}
	return props
}

func matchesSelector(selector, tag string, classes []string) bool {
	if strings.HasPrefix(selector, ".") {
		className := selector[1:]
		for _, c := range classes {
			if c == className {
				return true
			}
		}
		return false
	}
	return selector == tag
}

// ToRenderStyle converts resolved CSS properties to a render.Style.
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

func ToRenderStyle(props map[string]string) render.Style {
	var s render.Style
	if v, ok := props["color"]; ok {
		s.FG = parseColor(v)
	}
	if v, ok := props["border-color"]; ok {
		s.FG = parseColor(v)
	}
	if v, ok := props["background"]; ok {
		s.BG = parseColor(v)
	}
	if props["bold"] == "true" {
		s.Bold = true
	}
	if props["dim"] == "true" {
		s.Dim = true
	}
	if props["italic"] == "true" {
		s.Italic = true
	}
	if props["underline"] == "true" {
		s.Underline = true
	}
	if props["strikethrough"] == "true" {
		s.Strikethrough = true
	}
	if props["inverse"] == "true" {
		s.Inverse = true
	}
	return s
}
