package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/css"
	"github.com/tomyan/sumi/runtime/render"
)

// resolveProps resolves CSS properties for a node using the stylesheet.
// Returns nil if no stylesheet or no matching rules.
func resolveProps(stylesheet *style.Stylesheet, tag string, attrs map[string]string) map[string]string {
	if stylesheet == nil {
		return nil
	}
	classes := parseClasses(attrs)
	props := css.Resolve(stylesheet, tag, classes)
	if len(props) == 0 {
		return nil
	}
	return props
}

// parseClasses extracts CSS class names from an element's attributes.
func parseClasses(attrs map[string]string) []string {
	classAttr, ok := attrs["class"]
	if !ok || classAttr == "" {
		return nil
	}
	return strings.Fields(classAttr)
}

// writeStyleLiteral writes a sumi.Style{...} literal from resolved CSS properties.
func writeStyleLiteral(buf *bytes.Buffer, tabs string, props map[string]string) {
	s := css.ToRenderStyle(props)
	if s.IsZero() {
		return
	}
	fmt.Fprintf(buf, "%s\tStyle: sumi.Style{\n", tabs)
	writeStyleFields(buf, tabs, s)
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// writeStyleFields writes individual style fields (FG, BG, Bold, etc.).
func writeStyleFields(buf *bytes.Buffer, tabs string, s render.Style) {
	writeColorFields(buf, tabs, s)
	writeBoolStyleFields(buf, tabs, s)
}

// writeColorFields writes FG and BG color fields.
func writeColorFields(buf *bytes.Buffer, tabs string, s render.Style) {
	if s.FG.Name != "" {
		fmt.Fprintf(buf, "%s\t\tFG: sumi.Color{Name: %q},\n", tabs, s.FG.Name)
	}
	if s.BG.Name != "" {
		fmt.Fprintf(buf, "%s\t\tBG: sumi.Color{Name: %q},\n", tabs, s.BG.Name)
	}
}

// writeBoolStyleFields writes boolean style fields (Bold, Dim, Italic, etc.).
func writeBoolStyleFields(buf *bytes.Buffer, tabs string, s render.Style) {
	boolFields := []struct {
		set  bool
		name string
	}{
		{s.Bold, "Bold"}, {s.Dim, "Dim"}, {s.Italic, "Italic"},
		{s.Underline, "Underline"}, {s.Strikethrough, "Strikethrough"}, {s.Inverse, "Inverse"},
	}
	for _, f := range boolFields {
		if f.set {
			fmt.Fprintf(buf, "%s\t\t%s: true,\n", tabs, f.name)
		}
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
