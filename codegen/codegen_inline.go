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

	// Wire self-measurement pointers on the inlined component's root box
	if subExt != nil && subExt != ext && info.Script != nil && len(info.Script.SelfDecls) > 0 {
		rootBoxName := subExt.firstBoxName()
		if rootBoxName != "" {
			prefix := inst.VarName + "_"
			writeInlinedSelfWiring(&subExt.declBuf, info.Script.SelfDecls, rootBoxName, prefix)
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
	attrs := namespaceExprAttrs(n.Attributes, stateMap)

	// Track focusable index for cursor-focus correlation
	focusIdx := -1
	if ext != nil && isFocusableBox(attrs) {
		focusIdx = ext.focusablesSeen
		ext.focusablesSeen++
	}

	if ext != nil && hasDynamicCursor(attrs) {
		writeExtractedInlinedCursorBox(buf, n, stylesheet, indent, attrs, propMap, stateMap, tracker, ext, focusIdx)
		return
	}

	tabs := indentStr(indent)
	boxProps := resolveProps(stylesheet, "box", attrs)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, attrs, boxProps)
	if boxProps != nil {
		writeStyleLiteral(buf, tabs, boxProps)
	}
	writeInlinedBoxChildren(buf, n.Children, stylesheet, indent, tabs, propMap, stateMap, tracker, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeExtractedInlinedCursorBox extracts an inlined box with dynamic cursor as a named variable.
// Children are written to a temp buffer so text extractions go to declBuf first,
// then the cursor box declaration follows with references to extracted text nodes.
// focusIdx >= 0 means cursor is conditional on focus.
func writeExtractedInlinedCursorBox(treeBuf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, attrs map[string]string, propMap, stateMap map[string]string, tracker *instanceTracker, ext *extractionCtx, focusIdx int) {
	tabs := indentStr(indent)
	name := ext.nextBoxName()
	props := resolveProps(stylesheet, "box", attrs)

	// Process children first — text extractions go to ext.declBuf
	var childBuf bytes.Buffer
	writeInlinedBoxChildren(&childBuf, n.Children, stylesheet, 1, "\t", propMap, stateMap, tracker, ext)

	// Now write cursor box declaration after child extractions
	fmt.Fprintf(&ext.declBuf, "\t%s := &layout.Input{\n", name)
	fmt.Fprintf(&ext.declBuf, "\t\tKind: layout.KindBox,\n")
	writeBoxAttributes(&ext.declBuf, "\t", attrs, props)
	if props != nil {
		writeStyleLiteral(&ext.declBuf, "\t", props)
	}
	ext.declBuf.Write(childBuf.Bytes())
	fmt.Fprintf(&ext.declBuf, "\t}\n")

	ext.hasCursor = true
	writeCursorSync(&ext.syncBuf, name, attrs, focusIdx)

	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// namespaceExprAttrs returns a copy of attrs with expression values namespaced.
// e.g., cursor-x={cursor} with stateMap["cursor"]="textinput0_cursor" → cursor-x={textinput0_cursor}
func namespaceExprAttrs(attrs map[string]string, stateMap map[string]string) map[string]string {
	if len(stateMap) == 0 || len(attrs) == 0 {
		return attrs
	}
	result := make(map[string]string, len(attrs))
	for k, v := range attrs {
		if isExprValue(v) {
			expr := extractExprValue(v)
			expr = namespaceLineVars(expr, stateMap)
			result[k] = "{" + expr + "}"
		} else {
			result[k] = v
		}
	}
	return result
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

// buildStateNameMap creates a map from state/derived/bound-prop variable name to its namespaced version.
// e.g., "count" → "counter0_count" when the instance VarName is "counter0".
// For bind: attributes, the child's variable maps to the parent's variable instead.
func buildStateNameMap(inst *componentInstance) map[string]string {
	if inst.Info.Script == nil || !inst.Info.HasState {
		return nil
	}
	m := make(map[string]string)
	prefix := inst.VarName + "_"
	bindings := extractBindings(inst.Attrs)
	for _, sd := range inst.Info.Script.StateDecls {
		if parentVar, ok := bindings[sd.Name]; ok {
			m[sd.Name] = parentVar
		} else {
			m[sd.Name] = prefix + sd.Name
		}
	}
	for _, dd := range inst.Info.Script.DerivedDecls {
		m[dd.Name] = prefix + dd.Name
	}
	for _, sd := range inst.Info.Script.SelfDecls {
		m[sd.Name] = prefix + sd.Name
	}
	for _, fd := range inst.Info.Script.FuncDecls {
		m[fd.Name] = prefix + fd.Name
	}
	// Map props: bound → parent variable, expression → parent expression, literal → namespaced variable
	for _, pd := range inst.Info.Script.PropDecls {
		if _, already := m[pd.Name]; already {
			continue // already mapped via state decl
		}
		if parentVar, ok := bindings[pd.Name]; ok {
			m[pd.Name] = parentVar
		} else if val, ok := inst.Attrs[pd.Name]; ok && isExprValue(val) {
			m[pd.Name] = extractExprValue(val)
		} else {
			m[pd.Name] = prefix + pd.Name
		}
	}
	return m
}

// extractBindings returns a map from child variable name to parent variable name
// for bind: prefixed attributes. e.g., bind:value="name" → {"value": "name"}
func extractBindings(attrs map[string]string) map[string]string {
	bindings := make(map[string]string)
	for k, v := range attrs {
		if strings.HasPrefix(k, "bind:") {
			bindings[strings.TrimPrefix(k, "bind:")] = v
		}
	}
	return bindings
}

// resolveTextParts substitutes prop references in text parts.
// Literal props (no curlies) become StringParts. Expression props ({expr}) stay as ExprParts
// with the expression substituted.
func resolveTextParts(parts []template.Part, propMap map[string]string) []template.Part {
	resolved := make([]template.Part, 0, len(parts))
	for _, p := range parts {
		switch pt := p.(type) {
		case *template.ExprPart:
			if val, ok := propMap[pt.Expr]; ok {
				if isExprValue(val) {
					resolved = append(resolved, &template.ExprPart{Expr: extractExprValue(val)})
				} else {
					resolved = append(resolved, &template.StringPart{Value: val})
				}
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
			namespaced := namespaceLineVars(ep.Expr, stateMap)
			if namespaced != ep.Expr {
				result[i] = &template.ExprPart{Expr: namespaced}
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
	if sub.hasCursor {
		parent.hasCursor = true
	}
}

// writeComponentRefByName writes a component.Layout() reference by variable name.
func writeComponentRefByName(buf *bytes.Buffer, indent int, varName string) {
	tabs := indentStr(indent)
	fmt.Fprintf(buf, "%s%s.Layout(),\n", tabs, varName)
}
