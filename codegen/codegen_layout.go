package codegen

import (
	"bytes"
	"fmt"
	"sort"
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

	fmt.Fprintf(buf, "%sroot := &sumi.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: sumi.KindBox,\n", tabs)
	writeIdentityFields(buf, tabs, "root", nil)
	rootAttrs := map[string]string{"flex-direction": "column"}
	writeBoxAttributes(buf, tabs, rootAttrs, nil, ext)
	if hasDynamicChildren(doc.Children) {
		if ext != nil {
			// Build-once: root.Children rebuilt in sync
			fmt.Fprintf(buf, "%s}\n", tabs)
			writeDynamicChildrenSync(&ext.syncBuf, "root", doc.Children, stylesheet, ext.signals)
			return
		}
		writeDynamicChildren(buf, doc.Children, stylesheet, baseIndent, tabs, nil)
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

// writeTextInput writes a layout.Input literal for a text element.
// When ext is non-nil and the text has expressions, the node is extracted
// as a named variable for sync patching.
func writeTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}

	if ext != nil && !ext.inDynamic && (hasExprParts(n.Parts) || hasDynamicSyncAttrs(attrs)) {
		writeExtractedTextNode(buf, &ext.declBuf, n, tabs, ext)
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
	writeIdentityFields(buf, tabs, textTagOf(n), n.Attributes)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, content)
	if ce, ok := attrs["contenteditable"]; ok && ce == "true" {
		fmt.Fprintf(buf, "%s\tContentEditable: true,\n", tabs)
		writeCursorAttr(buf, tabs, attrs, nil)
	}
	writeOnHandlers(buf, tabs, attrs, nil, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeExtractedTextNode extracts an expression text node as a named variable.
// The declaration goes to declBuf; a variable reference goes to the tree buf.
func writeExtractedTextNode(treeBuf, declBuf *bytes.Buffer, n *template.TextElement, tabs string, ext *extractionCtx) {
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
	writeIdentityFields(declBuf, "\t", textTagOf(n), n.Attributes)
	fmt.Fprintf(declBuf, "\t\tContent: %s,\n", expr)
	writeOnHandlers(declBuf, "\t", n.Attributes, nil, ext)
	fmt.Fprintf(declBuf, "\t}\n")

	// Record sync entries: content (when dynamic) and state attributes.
	if hasExprParts(n.Parts) {
		ext.nodes = append(ext.nodes, extractedNode{
			varName:  name,
			syncExpr: expr,
		})
	}
	writeAttrSync(&ext.syncBuf, name, n.Attributes)

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
	if ext != nil && !ext.inDynamic && hasDynamicChildren(n.Children) {
		writeExtractedDynamicBox(buf, n, stylesheet, indent, ext)
		return
	}
	if ext != nil && !ext.inDynamic && (hasDynamicCursor(n.Attributes) || hasDynamicSyncAttrs(n.Attributes)) {
		writeExtractedCursorBox(buf, n, stylesheet, indent, ext)
		return
	}
	tabs := indentStr(indent)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: sumi.KindBox,\n", tabs)
	writeIdentityFields(buf, tabs, boxTagOf(n), n.Attributes)
	writeBoxAttributes(buf, tabs, n.Attributes, nil, ext)
	writeBoxChildren(buf, n.Children, stylesheet, indent, tabs, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// dynamicSyncAttrs are state-carrying attributes whose {expr} values are
// patched in the sync effect. on* handlers, cursor and layout attributes
// have dedicated emitters and are not listed here.
var dynamicSyncAttrs = map[string]bool{
	"class": true, "disabled": true, "checked": true, "open": true,
	"selected": true, "value": true, "href": true, "maxlength": true,
	"readonly": true,
}

// hasDynamicSyncAttrs reports whether any state attribute is expression-valued.
func hasDynamicSyncAttrs(attrs map[string]string) bool {
	for k, v := range attrs {
		if dynamicSyncAttrs[k] && isExprValue(v) {
			return true
		}
	}
	return false
}

// writeAttrSync emits sync patches for expression-valued state attributes.
// class also refreshes the Classes slice used by selector matching.
func writeAttrSync(buf *bytes.Buffer, name string, attrs map[string]string) {
	keys := make([]string, 0, len(attrs))
	for k, v := range attrs {
		if dynamicSyncAttrs[k] && isExprValue(v) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		expr := extractExprValue(attrs[k])
		if k == "class" {
			fmt.Fprintf(buf, "\t\t%s.Classes = sumi.SplitClasses(%s)\n", name, expr)
			fmt.Fprintf(buf, "\t\t%s.Attrs[\"class\"] = %s\n", name, expr)
			continue
		}
		fmt.Fprintf(buf, "\t\t%s.Attrs[%q] = sumi.AttrString(%s)\n", name, k, expr)
	}
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

// writeExtractedCursorBox extracts a box with dynamic cursor as a named variable.
// Cursor fields are patched in sync.
func writeExtractedCursorBox(treeBuf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	name := ext.nextBoxName()

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
	writeIdentityFields(&ext.declBuf, "\t", boxTagOf(n), n.Attributes)
	writeBoxAttributes(&ext.declBuf, "\t", n.Attributes, nil, ext)
	ext.declBuf.Write(childBuf.Bytes())
	fmt.Fprintf(&ext.declBuf, "\t}\n")

	// Write sync entries: cursor and state-attribute patching.
	if hasDynamicCursor(n.Attributes) {
		writeCursorSync(&ext.syncBuf, name, n.Attributes)
	}
	writeAttrSync(&ext.syncBuf, name, n.Attributes)

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// writeCursorSync writes CursorCol/CursorRow assignments to the sync buffer.
// Cursor visibility while unfocused is the component's concern (e.g. textedit
// returns -1 from its cursor expression when its focused signal is false).
func writeCursorSync(buf *bytes.Buffer, name string, attrs map[string]string) {
	if v, ok := attrs["cursor-x"]; ok && isExprValue(v) {
		fmt.Fprintf(buf, "\t\t%s.CursorCol = %s\n", name, extractExprValue(v))
	}
	if v, ok := attrs["cursor-y"]; ok && isExprValue(v) {
		fmt.Fprintf(buf, "\t\t%s.CursorRow = %s\n", name, extractExprValue(v))
	}
}

// writeExtractedDynamicBox extracts a box with dynamic children as a named variable.
// The box declaration (without Children) goes to declBuf; Children are rebuilt in syncBuf.
func writeExtractedDynamicBox(treeBuf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	name := ext.nextBoxName()

	// Write declaration to declBuf (at function scope, no Children)
	fmt.Fprintf(&ext.declBuf, "\t%s := &sumi.Input{\n", name)
	fmt.Fprintf(&ext.declBuf, "\t\tKind: sumi.KindBox,\n")
	writeIdentityFields(&ext.declBuf, "\t", boxTagOf(n), n.Attributes)
	writeBoxAttributes(&ext.declBuf, "\t", n.Attributes, nil, ext)
	fmt.Fprintf(&ext.declBuf, "\t}\n")
	writeAttrSync(&ext.syncBuf, name, n.Attributes)

	// Write dynamic children rebuild to syncBuf
	writeDynamicChildrenSync(&ext.syncBuf, name, n.Children, stylesheet, ext.signals)

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// writeBoxAttributes writes direction, width, height, padding, and border fields.
func writeBoxAttributes(buf *bytes.Buffer, tabs string, attrs map[string]string, props map[string]string, ext *extractionCtx) {
	if dir, ok := mergedAttr(attrs, props, "flex-direction"); ok {
		fmt.Fprintf(buf, "%s\tDirection: %q,\n", tabs, dir)
	}
	writeSizeAttr(buf, tabs, attrs, props, "width", "FixedWidth", "WidthPct")
	writeSizeAttr(buf, tabs, attrs, props, "height", "FixedHeight", "HeightPct")
	writeIntAttr(buf, tabs, attrs, props, "gap", "Gap")
	writeIntAttr(buf, tabs, attrs, props, "colspan", "ColSpan")
	writeIntAttr(buf, tabs, attrs, props, "rowspan", "RowSpan")
	writeIntAttr(buf, tabs, attrs, props, "min-height", "MinHeight")
	writeIntAttr(buf, tabs, attrs, props, "max-width", "MaxWidth")
	writeIntAttr(buf, tabs, attrs, props, "max-height", "MaxHeight")
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
	if m, ok := mergedAttr(attrs, props, "margin"); ok {
		fmt.Fprintf(buf, "%s\tMargin: sumi.ParseMargin(%q),\n", tabs, m)
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
	if d, ok := mergedAttr(attrs, props, "display"); ok {
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
	writeOnHandlers(buf, tabs, attrs, props, ext)
	writeCursorAttr(buf, tabs, attrs, props)
}

// writeOnHandlers emits the On event-handler map from on<type>={expr}
// attributes ("onclick" → "click"). Handlers declared with parameters are
// referenced directly (they take *sumi.DOMEvent); anything else is treated
// as a zero-arg func and wrapped with a nil check. The string-valued onkey
// attribute is legacy and handled separately via OnKey.
func writeOnHandlers(buf *bytes.Buffer, tabs string, attrs, props map[string]string, ext *extractionCtx) {
	var types []string
	for key, val := range attrs {
		if strings.HasPrefix(key, "on") && len(key) > 2 && isExprValue(val) {
			types = append(types, key[2:])
		}
	}
	if len(types) == 0 {
		return
	}
	sort.Strings(types)
	fmt.Fprintf(buf, "%s\tOn: map[string]func(*sumi.DOMEvent){\n", tabs)
	for _, typ := range types {
		val, _ := mergedAttr(attrs, props, "on"+typ)
		expr := extractExprValue(val)
		if ext != nil && ext.eventFuncs[expr] {
			fmt.Fprintf(buf, "%s\t\t%q: %s,\n", tabs, typ, expr)
			continue
		}
		fmt.Fprintf(buf, "%s\t\t%q: func(evt *sumi.DOMEvent) {\n", tabs, typ)
		fmt.Fprintf(buf, "%s\t\t\tif h := (%s); h != nil {\n", tabs, expr)
		fmt.Fprintf(buf, "%s\t\t\t\th()\n", tabs)
		fmt.Fprintf(buf, "%s\t\t\t}\n", tabs)
		fmt.Fprintf(buf, "%s\t\t},\n", tabs)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
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
