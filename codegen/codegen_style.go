package codegen

import (
	"bytes"
	"fmt"
	"github.com/tomyan/sumi/parser/template"
	"sort"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// textTagOf returns the CSS tag for a text-bearing element.
func textTagOf(n *template.TextElement) string {
	if n.Tag != "" {
		return n.Tag
	}
	return "text"
}

// boxTagOf returns the CSS tag for a container element.
func boxTagOf(n *template.BoxElement) string {
	if n.Tag != "" {
		return n.Tag
	}
	return "box"
}

// writeIdentityFields emits the element identity used by runtime CSS
// resolution (Tag/ID/Classes/Attrs on layout.Input).
func writeIdentityFields(buf *bytes.Buffer, tabs, tag string, attrs map[string]string) {
	fmt.Fprintf(buf, "%s\tTag: %q,\n", tabs, tag)
	if id := attrs["id"]; id != "" {
		fmt.Fprintf(buf, "%s\tID: %q,\n", tabs, id)
	}
	if classes := parseClasses(attrs); len(classes) > 0 {
		quoted := make([]string, len(classes))
		for i, c := range classes {
			quoted[i] = fmt.Sprintf("%q", c)
		}
		fmt.Fprintf(buf, "%s\tClasses: []string{%s},\n", tabs, strings.Join(quoted, ", "))
	}
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		// bind:* carry Go expressions, not attribute strings; codegen wires
		// them as handlers + sync, so they never belong in the Attrs map.
		if strings.HasPrefix(name, "bind:") {
			continue
		}
		names = append(names, name)
	}
	if len(names) > 0 {
		sort.Strings(names)
		fmt.Fprintf(buf, "%s\tAttrs: map[string]string{", tabs)
		for i, name := range names {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "%q: %q", name, attrs[name])
		}
		buf.WriteString("},\n")
	}
}

// parseClasses extracts CSS class names from an element's attributes.
func parseClasses(attrs map[string]string) []string {
	classAttr, ok := attrs["class"]
	if !ok || classAttr == "" {
		return nil
	}
	return strings.Fields(classAttr)
}

// colorLiteral renders a render.Color as a sumi.Color literal; "" for the
// zero colour.
func colorLiteral(c render.Color) string {
	switch {
	case c.Pair != nil:
		return fmt.Sprintf("sumi.Color{Pair: &sumi.ColorPair{Light: %s, Dark: %s}}",
			colorLiteralOrZero(c.Pair.Light), colorLiteralOrZero(c.Pair.Dark))
	case c.IsRGB:
		return fmt.Sprintf("sumi.Color{IsRGB: true, R: %d, G: %d, B: %d}", c.R, c.G, c.B)
	case c.Name != "":
		return fmt.Sprintf("sumi.Color{Name: %q}", c.Name)
	}
	return ""
}

func colorLiteralOrZero(c render.Color) string {
	if lit := colorLiteral(c); lit != "" {
		return lit
	}
	return "sumi.Color{}"
}

// writeInlineStyleFields writes style fields inline (for compact keyframe emission).
func writeInlineStyleFields(buf *bytes.Buffer, s render.Style) {
	if lit := colorLiteral(s.FG); lit != "" {
		fmt.Fprintf(buf, "FG: %s, ", lit)
	}
	if lit := colorLiteral(s.BG); lit != "" {
		fmt.Fprintf(buf, "BG: %s, ", lit)
	}
	if s.Bold {
		buf.WriteString("Bold: true, ")
	}
	if s.Dim {
		buf.WriteString("Dim: true, ")
	}
	if s.Italic {
		buf.WriteString("Italic: true, ")
	}
	if s.Underline {
		buf.WriteString("Underline: true, ")
	}
	if s.Strikethrough {
		buf.WriteString("Strikethrough: true, ")
	}
	if s.Inverse {
		buf.WriteString("Inverse: true, ")
	}
}

// mergedAttr returns the value for a layout-affecting attribute.
// Inline attributes (from the element) override stylesheet properties.
func mergedAttr(attrs map[string]string, props map[string]string, key string) (string, bool) {
	if v, ok := attrs[key]; ok {
		return v, true
	}
	if props != nil {
		if v, ok := props[key]; ok {
			return v, true
		}
	}
	return "", false
}
