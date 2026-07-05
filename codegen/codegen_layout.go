package codegen

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeLayoutTree writes the layout.Input tree construction code.
// When inClosure is true, adds an extra tab of indentation.
// When ext is non-nil, expression text nodes are extracted as named variables.
func writeLayoutTree(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, inClosure bool, ext *extractionCtx) {
	baseIndent := 1
	if inClosure {
		baseIndent = 2
	}
	tabs := indentStr(baseIndent)
	rootProps := resolveRootProps(stylesheet)

	fmt.Fprintf(buf, "%sroot := &sumi.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: sumi.KindBox,\n", tabs)
	rootAttrs := map[string]string{"flex-direction": "column"}
	writeBoxAttributes(buf, tabs, rootAttrs, rootProps)
	if rootProps != nil {
		writeStyleLiteral(buf, tabs, rootProps)
	}
	if hasDynamicChildren(doc.Children) {
		if ext != nil {
			// Build-once: root.Children rebuilt in sync
			fmt.Fprintf(buf, "%s}\n", tabs)
			writeDynamicChildrenSync(&ext.syncBuf, "root", doc.Children, stylesheet, ext.signals)
			return
		}
		writeDynamicChildren(buf, doc.Children, stylesheet, baseIndent, tabs, nil)
	} else if hasSlotChild(doc.Children) {
		// Slot children require dynamic construction (can't spread in a literal).
		writeSlotChildren(buf, doc.Children, stylesheet, baseIndent, tabs, ext)
	} else {
		fmt.Fprintf(buf, "%s\tChildren: []*sumi.Input{\n", tabs)
		for _, child := range doc.Children {
			writeInputNode(buf, child, stylesheet, baseIndent+2, ext)
		}
		fmt.Fprintf(buf, "%s\t},\n", tabs)
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeInputNode writes a layout.Input literal for a template AST node.
func writeInputNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	switch n := node.(type) {
	case *template.TextElement:
		writeTextInput(buf, n, stylesheet, indent, ext)
	case *template.BoxElement:
		writeBoxInput(buf, n, stylesheet, indent, ext)
	case *template.ComponentElement:
		writeSignalComponentRef(buf, n, indent, ext)
	case *template.SlotElement:
		writeSlotReference(buf, n, indent)
	}
}

// writeSignalComponentRef writes a child component's .Tree as a layout tree entry.
func writeSignalComponentRef(buf *bytes.Buffer, comp *template.ComponentElement, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	_, name := splitComponentName(comp.Name)
	idx := 0
	if ext != nil {
		idx = ext.componentIdx
		ext.componentIdx++
	}
	varName := fmt.Sprintf("%s%d", strings.ToLower(name[:1])+name[1:], idx)
	fmt.Fprintf(buf, "%s%s.Tree,\n", tabs, varName)
}

// writeSlotReference emits slot content from props as children.
// When inside a slice literal, this won't work — the parent must detect slots
// and use a dynamic children pattern instead.
func writeSlotReference(buf *bytes.Buffer, slot *template.SlotElement, indent int) {
	// This is handled by the parent box via hasSlotChild detection.
	// Slots inside a dynamic IIFE are emitted as: cs = append(cs, props.Name...)
	tabs := indentStr(indent)
	fieldName := strings.ToUpper(slot.Name[:1]) + slot.Name[1:]
	fmt.Fprintf(buf, "%scs = append(cs, props.%s...)\n", tabs, fieldName)
}

// writeSlotChildren emits Children referencing slot props.
// For simple cases (only slots), emits direct props references.
func writeSlotChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, baseIndent int, tabs string, ext *extractionCtx) {
	// Check if only slot children (no mixed content).
	allSlots := true
	for _, c := range children {
		if _, ok := c.(*template.SlotElement); !ok {
			allSlots = false
			break
		}
	}
	if allSlots && len(children) == 1 {
		// Single slot — direct reference.
		slot := children[0].(*template.SlotElement)
		fieldName := strings.ToUpper(slot.Name[:1]) + slot.Name[1:]
		fmt.Fprintf(buf, "%s\tChildren: props.%s,\n", tabs, fieldName)
		return
	}
	// Multiple slots or mixed — use IIFE.
	fmt.Fprintf(buf, "%s\tChildren: func() []*sumi.Input {\n", tabs)
	fmt.Fprintf(buf, "%s\t\tvar cs []*sumi.Input\n", tabs)
	for _, child := range children {
		if slot, ok := child.(*template.SlotElement); ok {
			fieldName := strings.ToUpper(slot.Name[:1]) + slot.Name[1:]
			fmt.Fprintf(buf, "%s\t\tcs = append(cs, props.%s...)\n", tabs, fieldName)
		}
		// TODO: handle mixed slot + non-slot children
	}
	fmt.Fprintf(buf, "%s\t\treturn cs\n", tabs)
	fmt.Fprintf(buf, "%s\t}(),\n", tabs)
}

// hasSlotChild checks if any immediate child is a SlotElement.
func hasSlotChild(children []template.Node) bool {
	for _, c := range children {
		if _, ok := c.(*template.SlotElement); ok {
			return true
		}
	}
	return false
}

// writeTextInput writes a layout.Input literal for a text element.
// When ext is non-nil and the text has expressions, the node is extracted
// as a named variable for sync patching.
func writeTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := n.ResolvedStyles

	if ext != nil && hasExprParts(n.Parts) && !ext.inDynamic {
		writeExtractedTextNode(buf, &ext.declBuf, n, props, tabs, ext)
		return
	}

	// Choose content expression: with signal unwrapping if available.
	var content string
	if ext != nil && len(ext.signals) > 0 {
		content = contentExprSignals(n.Parts, ext.signals)
	} else {
		content = contentExpr(n.Parts)
	}

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    sumi.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, content)
	if ce, ok := attrs["contenteditable"]; ok && ce == "true" {
		fmt.Fprintf(buf, "%s\tContentEditable: true,\n", tabs)
		writeCursorAttr(buf, tabs, attrs, props)
	}
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeExtractedTextNode extracts an expression text node as a named variable.
// The declaration goes to declBuf; a variable reference goes to the tree buf.
func writeExtractedTextNode(treeBuf, declBuf *bytes.Buffer, n *template.TextElement, props map[string]string, tabs string, ext *extractionCtx) {
	name := ext.nextNodeName()
	var expr string
	if len(ext.signals) > 0 {
		expr = contentExprSignals(n.Parts, ext.signals)
	} else {
		expr = contentExpr(n.Parts)
	}

	// Write declaration to declBuf (at function scope indent)
	fmt.Fprintf(declBuf, "\t%s := &sumi.Input{\n", name)
	fmt.Fprintf(declBuf, "\t\tKind:    sumi.KindText,\n")
	fmt.Fprintf(declBuf, "\t\tContent: %s,\n", expr)
	if props != nil {
		writeStyleLiteral(declBuf, "\t", props)
	}
	fmt.Fprintf(declBuf, "\t}\n")

	// Record sync entry
	ext.nodes = append(ext.nodes, extractedNode{
		varName:  name,
		syncExpr: expr,
	})

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// hasExprParts returns true if any part is an ExprPart.
func hasExprParts(parts []template.Part) bool {
	for _, p := range parts {
		if _, ok := p.(*template.ExprPart); ok {
			return true
		}
	}
	return false
}

// writeBoxInput writes a layout.Input literal for a box element.
// When ext is non-nil and the box has dynamic children or cursor, the box is extracted
// as a named variable for sync patching.
func writeBoxInput(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	// Track focusable index for cursor-focus correlation
	focusIdx := -1
	if ext != nil && isFocusableBox(n.Attributes) {
		focusIdx = ext.focusablesSeen
		ext.focusablesSeen++
	}

	if ext != nil && !ext.inDynamic && hasDynamicChildren(n.Children) {
		writeExtractedDynamicBox(buf, n, stylesheet, indent, ext)
		return
	}
	if ext != nil && !ext.inDynamic && (hasDynamicCursor(n.Attributes) || (focusIdx >= 0 && n.ResolvedFocus != nil)) {
		writeExtractedCursorBox(buf, n, stylesheet, indent, ext, focusIdx)
		return
	}
	tabs := indentStr(indent)
	props := n.ResolvedStyles

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: sumi.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, props)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	hoverProps := n.ResolvedHover
	if hoverProps != nil {
		writeHoverStyleLiteral(buf, tabs, hoverProps)
	}
	if n.ResolvedFocus != nil {
		writeFocusStyleLiteral(buf, tabs, n.ResolvedFocus)
	}
	if props != nil {
		writeTransitions(buf, tabs, props)
		writeAnimationSpec(buf, tabs, props)
	}
	writeBoxChildren(buf, n.Children, stylesheet, indent, tabs, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// hasDynamicCursor returns true if cursor-x or cursor-y contains an expression.
func hasDynamicCursor(attrs map[string]string) bool {
	if v, ok := attrs["cursor-x"]; ok && isExprValue(v) {
		return true
	}
	if v, ok := attrs["cursor-y"]; ok && isExprValue(v) {
		return true
	}
	return false
}

// isFocusableBox returns true if the box has both focusable="true" and an onkey handler,
// matching the criteria used by collectFocusableHandlers.
func isFocusableBox(attrs map[string]string) bool {
	return attrs["focusable"] == "true" && attrs["onkey"] != ""
}

// writeExtractedCursorBox extracts a box with dynamic cursor as a named variable.
// Cursor fields are patched in sync. focusIdx >= 0 means cursor is conditional on focus.
func writeExtractedCursorBox(treeBuf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx, focusIdx int) {
	tabs := indentStr(indent)
	name := ext.nextBoxName()
	props := n.ResolvedStyles

	// Render children first into a temp buffer: nested expression nodes
	// extract their own declarations into declBuf, which must precede this
	// box's declaration.
	var childBuf bytes.Buffer
	if len(n.Children) > 0 {
		fmt.Fprintf(&childBuf, "\t\tChildren: []*sumi.Input{\n")
		for _, child := range n.Children {
			writeInputNode(&childBuf, child, stylesheet, 3, ext)
		}
		fmt.Fprintf(&childBuf, "\t\t},\n")
	}

	// Write declaration to declBuf (at function scope)
	fmt.Fprintf(&ext.declBuf, "\t%s := &sumi.Input{\n", name)
	fmt.Fprintf(&ext.declBuf, "\t\tKind: sumi.KindBox,\n")
	writeBoxAttributes(&ext.declBuf, "\t", n.Attributes, props)
	if props != nil {
		writeStyleLiteral(&ext.declBuf, "\t", props)
	}
	if n.ResolvedHover != nil {
		writeHoverStyleLiteral(&ext.declBuf, "\t", n.ResolvedHover)
	}
	if n.ResolvedFocus != nil {
		writeFocusStyleLiteral(&ext.declBuf, "\t", n.ResolvedFocus)
	}
	ext.declBuf.Write(childBuf.Bytes())
	fmt.Fprintf(&ext.declBuf, "\t}\n")

	// Write sync entries: cursor patching and focus state.
	if hasDynamicCursor(n.Attributes) {
		writeCursorSync(&ext.syncBuf, name, n.Attributes, focusIdx)
	}
	if focusIdx >= 0 && n.ResolvedFocus != nil {
		fmt.Fprintf(&ext.syncBuf, "\t\t%s.Focused = focusIndex == %d\n", name, focusIdx)
	}

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// writeCursorSync writes CursorCol/CursorRow assignments to the sync buffer.
// When focusIdx >= 0, cursor is conditional on focus (only visible when focused).
func writeCursorSync(buf *bytes.Buffer, name string, attrs map[string]string, focusIdx int) {
	if focusIdx >= 0 {
		writeFocusConditionalCursor(buf, name, attrs, focusIdx)
		return
	}
	if v, ok := attrs["cursor-x"]; ok && isExprValue(v) {
		fmt.Fprintf(buf, "\t\t%s.CursorCol = %s\n", name, extractExprValue(v))
	}
	if v, ok := attrs["cursor-y"]; ok && isExprValue(v) {
		fmt.Fprintf(buf, "\t\t%s.CursorRow = %s\n", name, extractExprValue(v))
	}
}

// writeFocusedStateSync emits a focused state variable assignment if the component has one.
// focusedVar is the namespaced name of the focused state variable (e.g., "textinput0_focused").
func writeFocusedStateSync(buf *bytes.Buffer, focusedVar string, focusIdx int) {
	if focusedVar == "" || focusIdx < 0 {
		return
	}
	fmt.Fprintf(buf, "\t\t%s = focusIndex == %d\n", focusedVar, focusIdx)
}

// writeFocusConditionalCursor emits cursor assignment conditional on focusIndex.
func writeFocusConditionalCursor(buf *bytes.Buffer, name string, attrs map[string]string, focusIdx int) {
	fmt.Fprintf(buf, "\t\tif focusIndex == %d {\n", focusIdx)
	if v, ok := attrs["cursor-x"]; ok && isExprValue(v) {
		fmt.Fprintf(buf, "\t\t\t%s.CursorCol = %s\n", name, extractExprValue(v))
	}
	if v, ok := attrs["cursor-y"]; ok && isExprValue(v) {
		fmt.Fprintf(buf, "\t\t\t%s.CursorRow = %s\n", name, extractExprValue(v))
	}
	fmt.Fprintf(buf, "\t\t} else {\n")
	fmt.Fprintf(buf, "\t\t\t%s.CursorCol = -1\n", name)
	fmt.Fprintf(buf, "\t\t}\n")
}

// writeExtractedDynamicBox extracts a box with dynamic children as a named variable.
// The box declaration (without Children) goes to declBuf; Children are rebuilt in syncBuf.
func writeExtractedDynamicBox(treeBuf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	name := ext.nextBoxName()
	props := n.ResolvedStyles

	// Write declaration to declBuf (at function scope, no Children)
	fmt.Fprintf(&ext.declBuf, "\t%s := &sumi.Input{\n", name)
	fmt.Fprintf(&ext.declBuf, "\t\tKind: sumi.KindBox,\n")
	writeBoxAttributes(&ext.declBuf, "\t", n.Attributes, props)
	if props != nil {
		writeStyleLiteral(&ext.declBuf, "\t", props)
	}
	fmt.Fprintf(&ext.declBuf, "\t}\n")

	// Write dynamic children rebuild to syncBuf
	writeDynamicChildrenSync(&ext.syncBuf, name, n.Children, stylesheet, ext.signals)

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// writeBoxAttributes writes direction, width, height, padding, and border fields.
func writeBoxAttributes(buf *bytes.Buffer, tabs string, attrs map[string]string, props map[string]string) {
	if dir, ok := mergedAttr(attrs, props, "flex-direction"); ok {
		fmt.Fprintf(buf, "%s\tDirection: %q,\n", tabs, dir)
	}
	writeSizeAttr(buf, tabs, attrs, props, "width", "FixedWidth", "WidthPct")
	writeSizeAttr(buf, tabs, attrs, props, "height", "FixedHeight", "HeightPct")
	writeIntAttr(buf, tabs, attrs, props, "gap", "Gap")
	writeIntAttr(buf, tabs, attrs, props, "flex-grow", "FlexGrow")
	if j, ok := mergedAttr(attrs, props, "justify-content"); ok {
		fmt.Fprintf(buf, "%s\tJustify: %q,\n", tabs, normalizeAlignValue(j))
	}
	if a, ok := mergedAttr(attrs, props, "align-items"); ok {
		fmt.Fprintf(buf, "%s\tAlign: %q,\n", tabs, normalizeAlignValue(a))
	}
	if p, ok := mergedAttr(attrs, props, "padding"); ok {
		fmt.Fprintf(buf, "%s\tPadding: sumi.ParsePadding(%q),\n", tabs, p)
	}
	if b, ok := mergedAttr(attrs, props, "border"); ok {
		fmt.Fprintf(buf, "%s\tBorder: %q,\n", tabs, b)
	}
	if bt, ok := mergedAttr(attrs, props, "border-top"); ok {
		fmt.Fprintf(buf, "%s\tBorderTop: %q,\n", tabs, bt)
	}
	if bb, ok := mergedAttr(attrs, props, "border-bottom"); ok {
		fmt.Fprintf(buf, "%s\tBorderBottom: %q,\n", tabs, bb)
	}
	if bt, ok := mergedAttr(attrs, props, "border-title"); ok {
		if isExprValue(bt) {
			fmt.Fprintf(buf, "%s\tBorderTitle: %s,\n", tabs, extractExprValue(bt))
		} else {
			fmt.Fprintf(buf, "%s\tBorderTitle: %q,\n", tabs, bt)
		}
	}
	if bc, ok := mergedAttr(attrs, props, "border-collapse"); ok && bc == "collapse" {
		fmt.Fprintf(buf, "%s\tBorderCollapse: true,\n", tabs)
	}
	if o, ok := mergedAttr(attrs, props, "overflow"); ok {
		fmt.Fprintf(buf, "%s\tOverflow: %q,\n", tabs, o)
	}
	if sc, ok := mergedAttr(attrs, props, "scroll"); ok && isExprValue(sc) {
		fmt.Fprintf(buf, "%s\tScroll: %s,\n", tabs, extractExprValue(sc))
	}
	writeIntAttr(buf, tabs, attrs, props, "min-width", "MinWidth")
	if d, ok := mergedAttr(attrs, props, "display"); ok && d == "none" {
		fmt.Fprintf(buf, "%s\tDisplay: %q,\n", tabs, d)
	}
	if p, ok := mergedAttr(attrs, props, "position"); ok {
		fmt.Fprintf(buf, "%s\tPosition: %q,\n", tabs, p)
	}
	writeIntAttr(buf, tabs, attrs, props, "top", "Top")
	writeIntAttr(buf, tabs, attrs, props, "left", "Left")
	writeIntAttr(buf, tabs, attrs, props, "right", "Right")
	writeIntAttr(buf, tabs, attrs, props, "bottom", "Bottom")
	writeIntAttr(buf, tabs, attrs, props, "z-index", "ZIndex")
	if f, ok := mergedAttr(attrs, props, "focusable"); ok && f == "true" {
		fmt.Fprintf(buf, "%s\tFocusable: true,\n", tabs)
	}
	if ce, ok := mergedAttr(attrs, props, "contenteditable"); ok && ce == "true" {
		fmt.Fprintf(buf, "%s\tContentEditable: true,\n", tabs)
	}
	if oc, ok := mergedAttr(attrs, props, "onclick"); ok && isExprValue(oc) {
		fmt.Fprintf(buf, "%s\tOnClick: %s,\n", tabs, extractExprValue(oc))
	}
	writeCursorAttr(buf, tabs, attrs, props)
}

// writeCursorAttr writes CursorCol and CursorRow fields.
// Defaults to -1 (no cursor). Static integers are emitted directly.
// Dynamic expressions like "{cursor}" emit the bare expression.
func writeCursorAttr(buf *bytes.Buffer, tabs string, attrs, props map[string]string) {
	cx := resolveCursorValue(attrs, props, "cursor-x", "CursorCol")
	cy := resolveCursorValue(attrs, props, "cursor-y", "CursorRow")
	fmt.Fprintf(buf, "%s\tCursorCol: %s,\n", tabs, cx)
	fmt.Fprintf(buf, "%s\tCursorRow: %s,\n", tabs, cy)
}

// resolveCursorValue returns the Go expression for a cursor attribute.
// Returns "-1" if not set. Static integers are returned as-is.
// Expression values like "{expr}" return just "expr".
func resolveCursorValue(attrs, props map[string]string, attrKey, fieldName string) string {
	val, ok := mergedAttr(attrs, props, attrKey)
	if !ok {
		return "-1"
	}
	if isExprValue(val) {
		return extractExprValue(val)
	}
	if _, err := strconv.Atoi(val); err == nil {
		return val
	}
	return "-1"
}

// isExprValue returns true if the value is a dynamic expression like "{cursor}".
func isExprValue(val string) bool {
	return len(val) > 2 && val[0] == '{' && val[len(val)-1] == '}'
}

// extractExprValue extracts the expression from "{expr}" → "expr".
func extractExprValue(val string) string {
	return val[1 : len(val)-1]
}

// writeIntAttr writes an integer attribute field if present.
// Handles both static integers ("10") and expression values ("{expr}").
// normalizeAlignValue maps the CSS flex-* alignment keywords onto the layout
// engine's start/end vocabulary (both spellings are valid CSS Box Alignment).
func normalizeAlignValue(v string) string {
	switch v {
	case "flex-start":
		return "start"
	case "flex-end":
		return "end"
	}
	return v
}

func writeIntAttr(buf *bytes.Buffer, tabs string, attrs, props map[string]string, attrKey, fieldName string) {
	val, ok := mergedAttr(attrs, props, attrKey)
	if !ok {
		return
	}
	if isExprValue(val) {
		fmt.Fprintf(buf, "%s\t%s: %s,\n", tabs, fieldName, extractExprValue(val))
		return
	}
	v, ok := parseCellLength(val)
	if !ok {
		return
	}
	fmt.Fprintf(buf, "%s\t%s: %d,\n", tabs, fieldName, v)
}

// writeSizeAttr handles width/height, which additionally accept percentages:
// a trailing % emits the Pct field instead of the fixed one.
func writeSizeAttr(buf *bytes.Buffer, tabs string, attrs, props map[string]string, attrKey, fixedField, pctField string) {
	val, ok := mergedAttr(attrs, props, attrKey)
	if !ok {
		return
	}
	if pct, isPct := strings.CutSuffix(val, "%"); isPct {
		v, err := strconv.Atoi(pct)
		if err != nil {
			return
		}
		fmt.Fprintf(buf, "%s\t%s: %d,\n", tabs, pctField, v)
		return
	}
	writeIntAttr(buf, tabs, attrs, props, attrKey, fixedField)
}

// parseCellLength parses a cell-count length: a bare integer, or one with the
// `cell` unit or its alias `ch`. Reports false for anything else so
// pixel-derived units drop silently.
func parseCellLength(s string) (int, bool) {
	s = strings.TrimSuffix(strings.TrimSuffix(s, "cell"), "ch")
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

// writeBoxChildren writes the Children field of a box input if there are children.
func writeBoxChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, ext *extractionCtx) {
	if len(children) == 0 {
		return
	}
	if hasSlotChild(children) {
		writeSlotChildren(buf, children, stylesheet, indent, tabs, ext)
		return
	}
	if hasDynamicChildren(children) {
		var signals map[string]bool
		if ext != nil {
			signals = ext.signals
		}
		writeDynamicChildren(buf, children, stylesheet, indent, tabs, signals)
		return
	}
	fmt.Fprintf(buf, "%s\tChildren: []*sumi.Input{\n", tabs)
	for _, child := range children {
		writeInputNode(buf, child, stylesheet, indent+2, ext)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// indentStr returns n tab characters.
func indentStr(n int) string {
	return strings.Repeat("\t", n)
}
