package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeInlinedComponent inlines a child component's template into the parent tree.
// Props are resolved to literal values from the parent's element attributes.
func writeInlinedComponent(buf *bytes.Buffer, inst *componentInstance, indent int, tracker *instanceTracker, ext *extractionCtx) {
	info := inst.Info
	if info.Doc == nil {
		// Fallback: no AST available, use old comp.Layout() pattern
		writeComponentRefByName(buf, indent, inst.VarName)
		return
	}

	propMap := buildPropMap(inst)

	// The child template's root is a KindBox wrapper with "column" direction.
	// Its children are the actual template nodes. We inline those children directly,
	// since the child's root wrapper would add an unnecessary nesting level.
	// If the child has exactly one box child (typical pattern), inline that box.
	children := info.Doc.Children
	if len(children) == 1 {
		writeInlinedNode(buf, children[0], info.Stylesheet, indent, propMap, tracker, ext)
	} else {
		for _, child := range children {
			writeInlinedNode(buf, child, info.Stylesheet, indent, propMap, tracker, ext)
		}
	}
}

// writeInlinedNode writes a single inlined template node with prop substitution.
func writeInlinedNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, propMap map[string]string, tracker *instanceTracker, ext *extractionCtx) {
	switch n := node.(type) {
	case *template.TextElement:
		writeInlinedTextInput(buf, n, stylesheet, indent, propMap, ext)
	case *template.BoxElement:
		writeInlinedBoxInput(buf, n, stylesheet, indent, propMap, tracker, ext)
	}
}

// writeInlinedTextInput writes a text input with prop values substituted.
func writeInlinedTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int, propMap map[string]string, ext *extractionCtx) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)

	resolvedParts := resolveTextParts(n.Parts, propMap)
	expr := contentExpr(resolvedParts)

	// If after prop substitution there are still expressions, extract
	if ext != nil && hasExprParts(resolvedParts) {
		writeExtractedTextNode(buf, &ext.declBuf, &template.TextElement{Parts: resolvedParts, Attributes: n.Attributes}, props, tabs, ext)
		return
	}

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, expr)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeInlinedBoxInput writes a box input from an inlined component template.
func writeInlinedBoxInput(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, propMap map[string]string, tracker *instanceTracker, ext *extractionCtx) {
	tabs := indentStr(indent)
	boxProps := resolveProps(stylesheet, "box", n.Attributes)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, boxProps)
	if boxProps != nil {
		writeStyleLiteral(buf, tabs, boxProps)
	}
	writeInlinedBoxChildren(buf, n.Children, stylesheet, indent, tabs, propMap, tracker, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeInlinedBoxChildren writes children of an inlined box.
func writeInlinedBoxChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, propMap map[string]string, tracker *instanceTracker, ext *extractionCtx) {
	if len(children) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
	for _, child := range children {
		writeInlinedNode(buf, child, stylesheet, indent+2, propMap, tracker, ext)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// buildPropMap creates a map from prop name to literal value for an instance.
func buildPropMap(inst *componentInstance) map[string]string {
	propMap := make(map[string]string, len(inst.Info.Props))
	for _, propName := range inst.Info.Props {
		if val, ok := inst.Attrs[propName]; ok {
			propMap[propName] = val
		}
	}
	return propMap
}

// resolveTextParts substitutes prop references in text parts with literal string values.
// An ExprPart matching a prop name becomes a StringPart with the prop's value.
func resolveTextParts(parts []template.Part, propMap map[string]string) []template.Part {
	resolved := make([]template.Part, 0, len(parts))
	for _, p := range parts {
		switch pt := p.(type) {
		case *template.ExprPart:
			if val, ok := propMap[pt.Expr]; ok {
				resolved = append(resolved, &template.StringPart{Value: val})
			} else {
				resolved = append(resolved, pt)
			}
		default:
			resolved = append(resolved, p)
		}
	}
	return mergeAdjacentStrings(resolved)
}

// mergeAdjacentStrings combines consecutive StringParts into one.
func mergeAdjacentStrings(parts []template.Part) []template.Part {
	if len(parts) <= 1 {
		return parts
	}
	var merged []template.Part
	var sb strings.Builder
	for _, p := range parts {
		if sp, ok := p.(*template.StringPart); ok {
			sb.WriteString(sp.Value)
		} else {
			if sb.Len() > 0 {
				merged = append(merged, &template.StringPart{Value: sb.String()})
				sb.Reset()
			}
			merged = append(merged, p)
		}
	}
	if sb.Len() > 0 {
		merged = append(merged, &template.StringPart{Value: sb.String()})
	}
	return merged
}

// writeComponentRefByName writes a component.Layout() reference by variable name.
func writeComponentRefByName(buf *bytes.Buffer, indent int, varName string) {
	tabs := indentStr(indent)
	fmt.Fprintf(buf, "%s%s.Layout(),\n", tabs, varName)
}
