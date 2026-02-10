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
func ToRenderStyle(props map[string]string) render.Style {
	var s render.Style
	if v, ok := props["color"]; ok {
		s.FG = render.Color{Name: v}
	}
	if v, ok := props["border-color"]; ok {
		s.FG = render.Color{Name: v}
	}
	if v, ok := props["background"]; ok {
		s.BG = render.Color{Name: v}
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
