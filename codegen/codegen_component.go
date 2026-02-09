package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// generateComponent produces struct-based component code for child components.
func generateComponent(doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, opts Options) ([]byte, error) {
	var buf bytes.Buffer
	name := opts.ComponentName
	fmt.Fprintf(&buf, "package %s\n\n", opts.PackageName)
	hasStyles := docHasStyles(doc, stylesheet)
	writeComponentImports(&buf, docHasExprs(doc), hasStyles)
	writeComponentStruct(&buf, name, sc)
	writeComponentConstructor(&buf, name, sc)
	writeComponentLayoutMethod(&buf, name, doc, sc, stylesheet)
	writeComponentHandleKey(&buf, name, doc)
	writeComponentDirty(&buf, name)
	writeComponentMethods(&buf, name, sc)
	return format.Source(buf.Bytes())
}

// writeComponentImports writes imports needed by component code.
func writeComponentImports(buf *bytes.Buffer, hasExprs, hasStyles bool) {
	buf.WriteString("import (\n")
	if hasExprs {
		buf.WriteString("\t\"fmt\"\n\n")
	}
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/layout\"\n")
	if hasStyles {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	}
	buf.WriteString(")\n\n")
}

// docHasStyles checks if any element would produce a render.Style literal.
func docHasStyles(doc *template.Document, stylesheet *style.Stylesheet) bool {
	if stylesheet == nil {
		return false
	}
	for _, child := range doc.Children {
		if nodeHasStyles(child, stylesheet) {
			return true
		}
	}
	return false
}

// nodeHasStyles checks if a node would produce a render.Style literal.
func nodeHasStyles(node template.Node, stylesheet *style.Stylesheet) bool {
	switch n := node.(type) {
	case *template.TextElement:
		attrs := n.Attributes
		if attrs == nil {
			attrs = map[string]string{}
		}
		return resolveProps(stylesheet, "text", attrs) != nil
	case *template.BoxElement:
		if resolveProps(stylesheet, "box", n.Attributes) != nil {
			return true
		}
		for _, child := range n.Children {
			if nodeHasStyles(child, stylesheet) {
				return true
			}
		}
	}
	return false
}

// writeComponentStruct writes the component struct type definition.
func writeComponentStruct(buf *bytes.Buffer, name string, sc *script.Script) {
	fmt.Fprintf(buf, "type %sComponent struct {\n", name)
	writeStructPropFields(buf, sc.PropDecls)
	writeStructStateFields(buf, sc.StateDecls)
	buf.WriteString("\tdirty bool\n")
	buf.WriteString("}\n\n")
}

// writeStructPropFields writes prop fields to the struct.
func writeStructPropFields(buf *bytes.Buffer, props []script.PropDecl) {
	for _, p := range props {
		fmt.Fprintf(buf, "\t%s string\n", p.Name)
	}
}

// writeStructStateFields writes state fields to the struct.
func writeStructStateFields(buf *bytes.Buffer, states []script.StateDecl) {
	for _, s := range states {
		fmt.Fprintf(buf, "\t%s %s\n", s.Name, inferType(s.InitExpr))
	}
}

// inferType infers a Go type from an init expression.
func inferType(initExpr string) string {
	if initExpr == "0" || isIntLiteral(initExpr) {
		return "int"
	}
	if initExpr == "true" || initExpr == "false" {
		return "bool"
	}
	return "string"
}

// isIntLiteral returns true if the expression is a simple integer literal.
func isIntLiteral(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// writeComponentConstructor writes the NewXxxComponent constructor function.
func writeComponentConstructor(buf *bytes.Buffer, name string, sc *script.Script) {
	params := buildConstructorParams(sc.PropDecls)
	fmt.Fprintf(buf, "func New%sComponent(%s) *%sComponent {\n", name, params, name)
	fmt.Fprintf(buf, "\treturn &%sComponent{\n", name)
	writeConstructorPropAssignments(buf, sc.PropDecls)
	writeConstructorStateAssignments(buf, sc.StateDecls)
	buf.WriteString("\t\tdirty: true,\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")
}

// buildConstructorParams builds the parameter list for the constructor.
func buildConstructorParams(props []script.PropDecl) string {
	params := make([]string, len(props))
	for i, p := range props {
		params[i] = p.Name + " string"
	}
	return strings.Join(params, ", ")
}

// writeConstructorPropAssignments writes prop field assignments in the constructor.
func writeConstructorPropAssignments(buf *bytes.Buffer, props []script.PropDecl) {
	for _, p := range props {
		fmt.Fprintf(buf, "\t\t%s: %s,\n", p.Name, p.Name)
	}
}

// writeConstructorStateAssignments writes state field assignments in the constructor.
func writeConstructorStateAssignments(buf *bytes.Buffer, states []script.StateDecl) {
	for _, s := range states {
		fmt.Fprintf(buf, "\t\t%s: %s,\n", s.Name, s.InitExpr)
	}
}

// writeComponentLayoutMethod writes the Layout() method that returns *layout.Input.
func writeComponentLayoutMethod(buf *bytes.Buffer, name string, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet) {
	fmt.Fprintf(buf, "func (c *%sComponent) Layout() *layout.Input {\n", name)
	varNames := collectVarNames(sc)
	writeComponentLayoutTree(buf, doc, stylesheet, varNames)
	buf.WriteString("\treturn root\n")
	buf.WriteString("}\n\n")
}

// collectVarNames gathers all prop and state variable names for receiver prefixing.
func collectVarNames(sc *script.Script) map[string]bool {
	names := make(map[string]bool)
	for _, p := range sc.PropDecls {
		names[p.Name] = true
	}
	for _, s := range sc.StateDecls {
		names[s.Name] = true
	}
	return names
}

// writeComponentLayoutTree writes layout tree construction with receiver-prefixed expressions.
func writeComponentLayoutTree(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, varNames map[string]bool) {
	tabs := "\t"
	fmt.Fprintf(buf, "%sroot := &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:      layout.KindBox,\n", tabs)
	fmt.Fprintf(buf, "%s\tDirection: \"column\",\n", tabs)
	fmt.Fprintf(buf, "%s\tChildren:  []*layout.Input{\n", tabs)
	for _, child := range doc.Children {
		writeComponentInputNode(buf, child, stylesheet, 3, varNames)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeComponentInputNode writes a layout.Input literal with receiver-prefixed expressions.
func writeComponentInputNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, varNames map[string]bool) {
	switch n := node.(type) {
	case *template.TextElement:
		writeComponentTextInput(buf, n, stylesheet, indent, varNames)
	case *template.BoxElement:
		writeComponentBoxInput(buf, n, stylesheet, indent, varNames)
	}
}

// writeComponentTextInput writes a text input node with receiver-prefixed expressions.
func writeComponentTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int, varNames map[string]bool) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)
	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, componentContentExpr(n.Parts, varNames))
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeComponentBoxInput writes a box input node with receiver-prefixed expressions.
func writeComponentBoxInput(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, varNames map[string]bool) {
	tabs := indentStr(indent)
	props := resolveProps(stylesheet, "box", n.Attributes)
	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, props)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	writeComponentBoxChildren(buf, n.Children, stylesheet, indent, tabs, varNames)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeComponentBoxChildren writes children of a box with receiver-prefixed expressions.
func writeComponentBoxChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	if len(children) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
	for _, child := range children {
		writeComponentInputNode(buf, child, stylesheet, indent+2, varNames)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// componentContentExpr generates a content expression with receiver-prefixed variables.
func componentContentExpr(parts []template.Part, varNames map[string]bool) string {
	if len(parts) == 0 {
		return `""`
	}
	if allStringParts(parts) {
		return fmt.Sprintf("%q", concatStringParts(parts))
	}
	return buildReceiverSprintfExpr(parts, varNames)
}

// buildReceiverSprintfExpr builds fmt.Sprintf with c. prefix on known variables.
func buildReceiverSprintfExpr(parts []template.Part, varNames map[string]bool) string {
	var fmtStr strings.Builder
	var args []string
	for _, p := range parts {
		switch pt := p.(type) {
		case *template.StringPart:
			fmtStr.WriteString(strings.ReplaceAll(pt.Value, "%", "%%"))
		case *template.ExprPart:
			fmtStr.WriteString("%v")
			args = append(args, prefixExpr(pt.Expr, varNames))
		}
	}
	return fmt.Sprintf("fmt.Sprintf(%q, %s)", fmtStr.String(), strings.Join(args, ", "))
}

// prefixExpr adds "c." prefix to an expression if it matches a known variable name.
func prefixExpr(expr string, varNames map[string]bool) string {
	if varNames[expr] {
		return "c." + expr
	}
	return expr
}

// writeComponentHandleKey writes the HandleKey method if the document has an onkey handler.
func writeComponentHandleKey(buf *bytes.Buffer, name string, doc *template.Document) {
	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc == "" {
		return
	}
	fmt.Fprintf(buf, "func (c *%sComponent) HandleKey(key rune) {\n", name)
	fmt.Fprintf(buf, "\tc.%s()\n", onkeyFunc)
	buf.WriteString("}\n\n")
}

// writeComponentDirty writes the Dirty() method that returns and resets the dirty flag.
func writeComponentDirty(buf *bytes.Buffer, name string) {
	fmt.Fprintf(buf, "func (c *%sComponent) Dirty() bool {\n", name)
	buf.WriteString("\td := c.dirty\n")
	buf.WriteString("\tc.dirty = false\n")
	buf.WriteString("\treturn d\n")
	buf.WriteString("}\n\n")
}

// writeComponentMethods writes user-defined functions as methods on the component.
func writeComponentMethods(buf *bytes.Buffer, name string, sc *script.Script) {
	for _, funcDecl := range sc.FuncDecls {
		writeComponentMethod(buf, name, funcDecl)
	}
}

// writeComponentMethod writes a single user function as a method on the component.
func writeComponentMethod(buf *bytes.Buffer, name string, funcDecl script.FuncDecl) {
	fmt.Fprintf(buf, "func (c *%sComponent) %s() {\n", name, funcDecl.Name)
	writeComponentMethodBody(buf, funcDecl)
	buf.WriteString("}\n\n")
}

// writeComponentMethodBody writes the body of a component method with c. prefix and dirty flags.
func writeComponentMethodBody(buf *bytes.Buffer, funcDecl script.FuncDecl) {
	stateLines := buildStateLinesSet(funcDecl.StateAssignments)
	for _, line := range strings.Split(funcDecl.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		prefixed := prefixStateAssignment(trimmed, funcDecl.StateAssignments)
		fmt.Fprintf(buf, "\t%s\n", prefixed)
		if stateLines[trimmed] {
			buf.WriteString("\tc.dirty = true\n")
		}
	}
}

// prefixStateAssignment rewrites a line to use c. prefix for state variable assignments.
func prefixStateAssignment(line string, assignments []script.StateAssignment) string {
	for _, sa := range assignments {
		if line == sa.Line {
			return prefixAssignmentLine(line, sa.VarName)
		}
	}
	return line
}

// prefixAssignmentLine adds c. prefix to both sides of an assignment for the given variable.
func prefixAssignmentLine(line, varName string) string {
	result := strings.Replace(line, varName+" =", "c."+varName+" =", 1)
	return replaceRightSideVarRefs(result, varName)
}

// replaceRightSideVarRefs replaces variable references on the right side of an assignment with c. prefix.
func replaceRightSideVarRefs(line, varName string) string {
	eqIdx := strings.Index(line, "=")
	if eqIdx < 0 {
		return line
	}
	left := line[:eqIdx+1]
	right := line[eqIdx+1:]
	right = replaceVarRef(right, varName)
	return left + right
}

// replaceVarRef replaces standalone variable references with c. prefix.
func replaceVarRef(s, varName string) string {
	result := strings.ReplaceAll(s, " "+varName+" ", " c."+varName+" ")
	result = strings.ReplaceAll(result, " "+varName+"\n", " c."+varName+"\n")
	// Handle at end of string
	if strings.HasSuffix(result, " "+varName) {
		result = result[:len(result)-len(varName)] + "c." + varName
	}
	return result
}
