package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strconv"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
	"github.com/tomyan/sumi/runtime/css"
)

// Options configures code generation.
type Options struct {
	PackageName string
}

// Generate produces Go source code from a template AST, optional script, and optional stylesheet.
// When script is nil, generates static code (render once, wait for Enter).
// When script has state, generates reactive code with an event loop.
// When stylesheet is non-nil, styles are resolved at codegen time and emitted as render.Style literals.
func Generate(doc *template.Document, sc *script.Script, ss *style.Stylesheet, opts Options) ([]byte, error) {
	reactive := sc != nil && len(sc.StateDecls) > 0
	hasExprs := docHasExprs(doc)

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "package %s\n\n", opts.PackageName)

	// Imports
	writeImports(&buf, reactive, hasExprs)

	buf.WriteString("func Run() {\n")

	if reactive {
		writeReactiveBody(&buf, doc, sc, ss)
	} else {
		writeStaticBody(&buf, doc, ss)
	}

	buf.WriteString("}\n\n")

	// Generate the renderTree helper function.
	writeRenderTreeFunc(&buf)

	return format.Source(buf.Bytes())
}

// writeImports writes the import block.
func writeImports(buf *bytes.Buffer, reactive, hasExprs bool) {
	buf.WriteString("import (\n")
	if hasExprs {
		buf.WriteString("\t\"fmt\"\n")
	}
	if !reactive {
		buf.WriteString("\t\"bufio\"\n")
	}
	buf.WriteString("\t\"os\"\n\n")
	if reactive {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/input\"\n")
	}
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/layout\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	buf.WriteString(")\n\n")
}

// writeStaticBody generates the static (non-reactive) function body.
func writeStaticBody(buf *bytes.Buffer, doc *template.Document, ss *style.Stylesheet) {
	writeLayoutTree(buf, doc, ss, false)

	buf.WriteString("\ttree := layout.Layout(root, 80, 24)\n")
	buf.WriteString("\tbuf := render.NewBuffer(80, 24)\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\trenderTree(buf, tree)\n")
	buf.WriteString("\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\tbufio.NewScanner(os.Stdin).Scan()\n")
	buf.WriteString("\trender.ExitAlternateScreen(os.Stdout)\n")
}

// writeReactiveBody generates the reactive function body with event loop.
func writeReactiveBody(buf *bytes.Buffer, doc *template.Document, sc *script.Script, ss *style.Stylesheet) {
	// State declarations
	for _, sd := range sc.StateDecls {
		fmt.Fprintf(buf, "\t%s := %s\n", sd.Name, sd.InitExpr)
	}
	buf.WriteString("\n")

	// Dirty flag and function declarations
	buf.WriteString("\tdirty := true\n")

	for _, fd := range sc.FuncDecls {
		fmt.Fprintf(buf, "\t%s := func() {\n", fd.Name)
		// Write the function body, adding dirty=true after state assignments
		writeReactiveFuncBody(buf, fd)
		buf.WriteString("\t}\n")
	}
	buf.WriteString("\n")

	// doRender closure
	buf.WriteString("\tvar prevBuf *render.Buffer\n")
	buf.WriteString("\tdoRender := func() {\n")

	writeLayoutTree(buf, doc, ss, true)

	buf.WriteString("\t\ttree := layout.Layout(root, 80, 24)\n")
	buf.WriteString("\t\tbuf := render.NewBuffer(80, 24)\n")
	buf.WriteString("\t\trenderTree(buf, tree)\n")
	buf.WriteString("\t\tif prevBuf != nil {\n")
	buf.WriteString("\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tprevBuf = buf\n")
	buf.WriteString("\t\tdirty = false\n")
	buf.WriteString("\t}\n\n")

	// Suppress unused variable warnings for functions without onkey
	for _, fd := range sc.FuncDecls {
		if !docHasOnkey(doc, fd.Name) {
			fmt.Fprintf(buf, "\t_ = %s\n", fd.Name)
		}
	}

	// Setup
	buf.WriteString("\trestore, _ := input.EnableRawMode(int(os.Stdin.Fd()))\n")
	buf.WriteString("\tdefer restore()\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\tdefer render.ExitAlternateScreen(os.Stdout)\n\n")

	// Initial render
	buf.WriteString("\tdoRender()\n\n")

	// Event loop
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tkey, err := input.ReadKey(os.Stdin)\n")
	buf.WriteString("\t\tif err != nil || key == 'q' {\n")
	buf.WriteString("\t\t\tbreak\n")
	buf.WriteString("\t\t}\n")

	// Find onkey handler on root element
	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc != "" {
		fmt.Fprintf(buf, "\t\t%s()\n", onkeyFunc)
	}

	buf.WriteString("\t\tif dirty {\n")
	buf.WriteString("\t\t\tdoRender()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
}

// writeReactiveFuncBody writes a function body, adding dirty=true after each state assignment line.
func writeReactiveFuncBody(buf *bytes.Buffer, fd script.FuncDecl) {
	lines := strings.Split(fd.Body, "\n")
	stateLines := make(map[string]bool)
	for _, sa := range fd.StateAssignments {
		stateLines[sa.Line] = true
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fmt.Fprintf(buf, "\t\t%s\n", trimmed)
		if stateLines[trimmed] {
			buf.WriteString("\t\tdirty = true\n")
		}
	}
}

// writeLayoutTree writes the layout.Input tree construction code.
// When inClosure is true, adds an extra tab of indentation.
func writeLayoutTree(buf *bytes.Buffer, doc *template.Document, ss *style.Stylesheet, inClosure bool) {
	baseIndent := 1
	if inClosure {
		baseIndent = 2
	}
	tabs := indentStr(baseIndent)

	fmt.Fprintf(buf, "%sroot := &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:      layout.KindBox,\n", tabs)
	fmt.Fprintf(buf, "%s\tDirection: \"column\",\n", tabs)
	fmt.Fprintf(buf, "%s\tChildren:  []*layout.Input{\n", tabs)
	for _, child := range doc.Children {
		writeInputNode(buf, child, ss, baseIndent+2)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// resolveProps resolves CSS properties for a node using the stylesheet.
// Returns nil if no stylesheet or no matching rules.
func resolveProps(ss *style.Stylesheet, tag string, attrs map[string]string) map[string]string {
	if ss == nil {
		return nil
	}
	var classes []string
	if classAttr, ok := attrs["class"]; ok && classAttr != "" {
		classes = strings.Fields(classAttr)
	}
	props := css.Resolve(ss, tag, classes)
	if len(props) == 0 {
		return nil
	}
	return props
}

// writeStyleLiteral writes a render.Style{...} literal from resolved CSS properties.
func writeStyleLiteral(buf *bytes.Buffer, tabs string, props map[string]string) {
	s := css.ToRenderStyle(props)
	if s.IsZero() {
		return
	}
	fmt.Fprintf(buf, "%s\tStyle: render.Style{\n", tabs)
	if s.FG.Name != "" {
		fmt.Fprintf(buf, "%s\t\tFG: render.Color{Name: %q},\n", tabs, s.FG.Name)
	}
	if s.BG.Name != "" {
		fmt.Fprintf(buf, "%s\t\tBG: render.Color{Name: %q},\n", tabs, s.BG.Name)
	}
	if s.Bold {
		fmt.Fprintf(buf, "%s\t\tBold: true,\n", tabs)
	}
	if s.Dim {
		fmt.Fprintf(buf, "%s\t\tDim: true,\n", tabs)
	}
	if s.Italic {
		fmt.Fprintf(buf, "%s\t\tItalic: true,\n", tabs)
	}
	if s.Underline {
		fmt.Fprintf(buf, "%s\t\tUnderline: true,\n", tabs)
	}
	if s.Strikethrough {
		fmt.Fprintf(buf, "%s\t\tStrikethrough: true,\n", tabs)
	}
	if s.Inverse {
		fmt.Fprintf(buf, "%s\t\tInverse: true,\n", tabs)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
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

// writeInputNode writes a layout.Input literal for a template AST node.
func writeInputNode(buf *bytes.Buffer, node template.Node, ss *style.Stylesheet, indent int) {
	tabs := indentStr(indent)

	switch n := node.(type) {
	case *template.TextElement:
		attrs := n.Attributes
		if attrs == nil {
			attrs = map[string]string{}
		}
		props := resolveProps(ss, "text", attrs)

		fmt.Fprintf(buf, "%s{\n", tabs)
		fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
		fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, contentExpr(n.Parts))
		if props != nil {
			writeStyleLiteral(buf, tabs, props)
		}
		fmt.Fprintf(buf, "%s},\n", tabs)

	case *template.BoxElement:
		props := resolveProps(ss, "box", n.Attributes)

		fmt.Fprintf(buf, "%s{\n", tabs)
		fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)

		if dir, ok := mergedAttr(n.Attributes, props, "direction"); ok {
			fmt.Fprintf(buf, "%s\tDirection: %q,\n", tabs, dir)
		}
		if w, ok := mergedAttr(n.Attributes, props, "width"); ok {
			if v, err := strconv.Atoi(w); err == nil {
				fmt.Fprintf(buf, "%s\tFixedWidth:  %d,\n", tabs, v)
			}
		}
		if h, ok := mergedAttr(n.Attributes, props, "height"); ok {
			if v, err := strconv.Atoi(h); err == nil {
				fmt.Fprintf(buf, "%s\tFixedHeight: %d,\n", tabs, v)
			}
		}
		if p, ok := mergedAttr(n.Attributes, props, "padding"); ok {
			fmt.Fprintf(buf, "%s\tPadding: layout.ParsePadding(%q),\n", tabs, p)
		}
		if b, ok := mergedAttr(n.Attributes, props, "border"); ok {
			fmt.Fprintf(buf, "%s\tBorder: %q,\n", tabs, b)
		}
		if props != nil {
			writeStyleLiteral(buf, tabs, props)
		}

		if len(n.Children) > 0 {
			fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
			for _, child := range n.Children {
				writeInputNode(buf, child, ss, indent+2)
			}
			fmt.Fprintf(buf, "%s\t},\n", tabs)
		}

		fmt.Fprintf(buf, "%s},\n", tabs)
	}
}

// contentExpr generates the Go expression for a TextElement's content.
// Pure string parts → quoted string literal.
// Mixed parts with expressions → fmt.Sprintf call.
func contentExpr(parts []template.Part) string {
	if len(parts) == 0 {
		return `""`
	}

	// Check if all parts are string literals.
	allStrings := true
	for _, p := range parts {
		if _, ok := p.(*template.ExprPart); ok {
			allStrings = false
			break
		}
	}

	if allStrings {
		// Concatenate all string parts into a single quoted string.
		var sb strings.Builder
		for _, p := range parts {
			sb.WriteString(p.(*template.StringPart).Value)
		}
		return fmt.Sprintf("%q", sb.String())
	}

	// Build fmt.Sprintf format string and args.
	var fmtStr strings.Builder
	var args []string
	for _, p := range parts {
		switch pt := p.(type) {
		case *template.StringPart:
			// Escape % in the format string
			fmtStr.WriteString(strings.ReplaceAll(pt.Value, "%", "%%"))
		case *template.ExprPart:
			fmtStr.WriteString("%v")
			args = append(args, pt.Expr)
		}
	}

	return fmt.Sprintf("fmt.Sprintf(%q, %s)", fmtStr.String(), strings.Join(args, ", "))
}

// writeRenderTreeFunc generates the renderTree helper function.
func writeRenderTreeFunc(buf *bytes.Buffer) {
	buf.WriteString("func renderTree(buf *render.Buffer, box *layout.Box) {\n")
	buf.WriteString("\tif box.Border != \"\" && box.Border != \"none\" {\n")
	buf.WriteString("\t\tbuf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif box.Content != \"\" {\n")
	buf.WriteString("\t\tbuf.WriteStyledText(box.Y, box.X, box.Content, box.Style)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tfor _, child := range box.Children {\n")
	buf.WriteString("\t\trenderTree(buf, child)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

// docHasExprs checks if any TextElement in the document has ExprParts.
func docHasExprs(doc *template.Document) bool {
	for _, child := range doc.Children {
		if nodeHasExprs(child) {
			return true
		}
	}
	return false
}

func nodeHasExprs(node template.Node) bool {
	switch n := node.(type) {
	case *template.TextElement:
		for _, p := range n.Parts {
			if _, ok := p.(*template.ExprPart); ok {
				return true
			}
		}
	case *template.BoxElement:
		for _, child := range n.Children {
			if nodeHasExprs(child) {
				return true
			}
		}
	}
	return false
}

// findRootOnkey finds an onkey attribute on the root-level box element.
func findRootOnkey(doc *template.Document) string {
	for _, child := range doc.Children {
		if box, ok := child.(*template.BoxElement); ok {
			if handler, ok := box.Attributes["onkey"]; ok {
				return handler
			}
		}
	}
	return ""
}

// docHasOnkey checks if any box element in the document has onkey referencing the given function name.
func docHasOnkey(doc *template.Document, funcName string) bool {
	for _, child := range doc.Children {
		if boxHasOnkey(child, funcName) {
			return true
		}
	}
	return false
}

func boxHasOnkey(node template.Node, funcName string) bool {
	if box, ok := node.(*template.BoxElement); ok {
		if handler, ok := box.Attributes["onkey"]; ok && handler == funcName {
			return true
		}
		for _, child := range box.Children {
			if boxHasOnkey(child, funcName) {
				return true
			}
		}
	}
	return false
}

func indentStr(n int) string {
	s := ""
	for range n {
		s += "\t"
	}
	return s
}
