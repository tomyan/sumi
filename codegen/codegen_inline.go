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
// For stateful components, state variable references are namespaced.
func writeInlinedComponent(buf *bytes.Buffer, inst *componentInstance, indent int, tracker *instanceTracker, ext *extractionCtx) {
	info := inst.Info
	if info.Doc == nil {
		writeComponentRefByName(buf, indent, inst.VarName)
		return
	}

	propMap := buildPropMap(inst)
	stateMap := buildStateNameMap(inst)

	// Use a sub-extraction context with the instance's namespace prefix
	// so extracted nodes get names like "counter0_node0"
	var subExt *extractionCtx
	if ext != nil && info.HasState {
		subExt = newExtractionCtx(inst.VarName + "_")
	} else {
		subExt = ext
	}

	children := info.Doc.Children
	if len(children) == 1 {
		writeInlinedNode(buf, children[0], info.Stylesheet, indent, propMap, stateMap, tracker, subExt)
	} else {
		for _, child := range children {
			writeInlinedNode(buf, child, info.Stylesheet, indent, propMap, stateMap, tracker, subExt)
		}
	}

	// Merge sub-extraction context back into parent
	if subExt != nil && subExt != ext && ext != nil {
		mergeExtractionCtx(ext, subExt)
	}
}

// writeInlinedNode writes a single inlined template node with prop substitution.
func writeInlinedNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, propMap, stateMap map[string]string, tracker *instanceTracker, ext *extractionCtx) {
	switch n := node.(type) {
	case *template.TextElement:
		writeInlinedTextInput(buf, n, stylesheet, indent, propMap, stateMap, ext)
	case *template.BoxElement:
		writeInlinedBoxInput(buf, n, stylesheet, indent, propMap, stateMap, tracker, ext)
	}
}

// writeInlinedTextInput writes a text input with prop values substituted and state vars namespaced.
func writeInlinedTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int, propMap, stateMap map[string]string, ext *extractionCtx) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)

	resolvedParts := resolveTextParts(n.Parts, propMap)
	resolvedParts = namespaceExprParts(resolvedParts, stateMap)
	expr := contentExpr(resolvedParts)

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
func writeInlinedBoxInput(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, propMap, stateMap map[string]string, tracker *instanceTracker, ext *extractionCtx) {
	tabs := indentStr(indent)
	boxProps := resolveProps(stylesheet, "box", n.Attributes)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, boxProps)
	if boxProps != nil {
		writeStyleLiteral(buf, tabs, boxProps)
	}
	writeInlinedBoxChildren(buf, n.Children, stylesheet, indent, tabs, propMap, stateMap, tracker, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeInlinedBoxChildren writes children of an inlined box.
func writeInlinedBoxChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, propMap, stateMap map[string]string, tracker *instanceTracker, ext *extractionCtx) {
	if len(children) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
	for _, child := range children {
		writeInlinedNode(buf, child, stylesheet, indent+2, propMap, stateMap, tracker, ext)
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

// buildStateNameMap creates a map from state variable name to its namespaced version.
// e.g., "count" → "counter0_count" when the instance VarName is "counter0".
func buildStateNameMap(inst *componentInstance) map[string]string {
	if inst.Info.Script == nil || !inst.Info.HasState {
		return nil
	}
	m := make(map[string]string)
	prefix := inst.VarName + "_"
	for _, sd := range inst.Info.Script.StateDecls {
		m[sd.Name] = prefix + sd.Name
	}
	return m
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

// namespaceExprParts renames ExprPart expressions that match state variable names.
// e.g., ExprPart{Expr: "count"} with stateMap["count"] = "counter0_count" → ExprPart{Expr: "counter0_count"}
func namespaceExprParts(parts []template.Part, stateMap map[string]string) []template.Part {
	if len(stateMap) == 0 {
		return parts
	}
	result := make([]template.Part, len(parts))
	for i, p := range parts {
		if ep, ok := p.(*template.ExprPart); ok {
			if newName, found := stateMap[ep.Expr]; found {
				result[i] = &template.ExprPart{Expr: newName}
				continue
			}
		}
		result[i] = p
	}
	return result
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

// mergeExtractionCtx merges a sub-context's extractions back into the parent.
func mergeExtractionCtx(parent, sub *extractionCtx) {
	parent.declBuf.Write(sub.declBuf.Bytes())
	parent.syncBuf.Write(sub.syncBuf.Bytes())
	parent.nodes = append(parent.nodes, sub.nodes...)
}

// writeComponentRefByName writes a component.Layout() reference by variable name.
func writeComponentRefByName(buf *bytes.Buffer, indent int, varName string) {
	tabs := indentStr(indent)
	fmt.Fprintf(buf, "%s%s.Layout(),\n", tabs, varName)
}
