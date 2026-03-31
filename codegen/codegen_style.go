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

// resolveHoverProps resolves :hover CSS properties for a node.
func resolveHoverProps(stylesheet *style.Stylesheet, tag string, attrs map[string]string) map[string]string {
	if stylesheet == nil {
		return nil
	}
	classes := parseClasses(attrs)
	return css.ResolveHover(stylesheet, tag, classes)
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

// writeHoverStyleLiteral writes a HoverStyle: sumi.Style{...} literal.
func writeHoverStyleLiteral(buf *bytes.Buffer, tabs string, props map[string]string) {
	s := css.ToRenderStyle(props)
	if s.IsZero() {
		return
	}
	fmt.Fprintf(buf, "%s\tHoverStyle: sumi.Style{\n", tabs)
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
	writeColorField(buf, tabs, "FG", s.FG)
	writeColorField(buf, tabs, "BG", s.BG)
}

func writeColorField(buf *bytes.Buffer, tabs, field string, c render.Color) {
	if c.IsRGB {
		fmt.Fprintf(buf, "%s\t\t%s: sumi.Color{IsRGB: true, R: %d, G: %d, B: %d},\n", tabs, field, c.R, c.G, c.B)
	} else if c.Name != "" {
		fmt.Fprintf(buf, "%s\t\t%s: sumi.Color{Name: %q},\n", tabs, field, c.Name)
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

// writeTransitions emits Transitions field from CSS transition properties.
func writeTransitions(buf *bytes.Buffer, tabs string, props map[string]string) {
	specs := css.ParseTransitions(props)
	if len(specs) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s\tTransitions: []sumi.TransitionSpec{\n", tabs)
	for _, s := range specs {
		fmt.Fprintf(buf, "%s\t\t{Property: %q, DurationMs: %d, DelayMs: %d, TimingFunction: sumi.TimingFunction{Name: %q, X1: %v, Y1: %v, X2: %v, Y2: %v}},\n",
			tabs, s.Property, s.DurationMs, s.DelayMs,
			s.TimingFunction.Name, s.TimingFunction.X1, s.TimingFunction.Y1, s.TimingFunction.X2, s.TimingFunction.Y2)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// writeAnimationSpec emits AnimationSpec field from CSS animation properties.
func writeAnimationSpec(buf *bytes.Buffer, tabs string, props map[string]string) {
	spec := css.ParseAnimation(props)
	if spec == nil {
		return
	}
	fmt.Fprintf(buf, "%s\tAnimationSpec: &sumi.AnimationSpec{\n", tabs)
	fmt.Fprintf(buf, "%s\t\tName: %q,\n", tabs, spec.Name)
	fmt.Fprintf(buf, "%s\t\tDurationMs: %d,\n", tabs, spec.DurationMs)
	fmt.Fprintf(buf, "%s\t\tTimingFunction: sumi.TimingFunction{Name: %q, X1: %v, Y1: %v, X2: %v, Y2: %v},\n",
		tabs, spec.TimingFunction.Name, spec.TimingFunction.X1, spec.TimingFunction.Y1, spec.TimingFunction.X2, spec.TimingFunction.Y2)
	fmt.Fprintf(buf, "%s\t\tDelayMs: %d,\n", tabs, spec.DelayMs)
	fmt.Fprintf(buf, "%s\t\tIterationCount: %d,\n", tabs, spec.IterationCount)
	fmt.Fprintf(buf, "%s\t\tDirection: %q,\n", tabs, spec.Direction)
	fmt.Fprintf(buf, "%s\t\tFillMode: %q,\n", tabs, spec.FillMode)
	fmt.Fprintf(buf, "%s\t\tPlayState: %q,\n", tabs, spec.PlayState)
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// writeInlineStyleFields writes style fields inline (for compact keyframe emission).
func writeInlineStyleFields(buf *bytes.Buffer, s render.Style) {
	if s.FG.IsRGB {
		fmt.Fprintf(buf, "FG: sumi.Color{IsRGB: true, R: %d, G: %d, B: %d}, ", s.FG.R, s.FG.G, s.FG.B)
	} else if s.FG.Name != "" {
		fmt.Fprintf(buf, "FG: sumi.Color{Name: %q}, ", s.FG.Name)
	}
	if s.BG.IsRGB {
		fmt.Fprintf(buf, "BG: sumi.Color{IsRGB: true, R: %d, G: %d, B: %d}, ", s.BG.R, s.BG.G, s.BG.B)
	} else if s.BG.Name != "" {
		fmt.Fprintf(buf, "BG: sumi.Color{Name: %q}, ", s.BG.Name)
	}
	if s.Bold { buf.WriteString("Bold: true, ") }
	if s.Dim { buf.WriteString("Dim: true, ") }
	if s.Italic { buf.WriteString("Italic: true, ") }
	if s.Underline { buf.WriteString("Underline: true, ") }
	if s.Strikethrough { buf.WriteString("Strikethrough: true, ") }
	if s.Inverse { buf.WriteString("Inverse: true, ") }
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
