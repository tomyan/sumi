package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeSignalBody generates the body of Run() using the signal-based reactive model.
func writeSignalBody(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet) {
	signals := signalVarNames(sc)

	// Declare signals.
	for _, d := range sc.SignalDecls {
		fmt.Fprintf(buf, "\t%s := signal.New(%s)\n", d.Name, d.InitExpr)
	}
	for _, d := range sc.ComputedDecls {
		fmt.Fprintf(buf, "\t%s := signal.From(%s)\n", d.Name, d.Expr)
	}
	buf.WriteString("\n")

	// Declare app variable.
	buf.WriteString("\tvar app *tui.App\n")

	// Emit function closures.
	for _, fd := range sc.FuncDecls {
		writeSignalFuncDecl(buf, fd)
	}

	// Build layout tree with signal-aware expressions.
	ext := newExtractionCtx("")
	buf.WriteString("\n")
	writeSignalTree(buf, doc, sc, stylesheet, signals, ext)

	// Emit effect that syncs signal values to tree nodes.
	if ext.hasSyncContent() {
		writeSignalEffect(buf, ext)
	}

	// Emit render function and app setup.
	writeSignalRender(buf, ext)
}

// writeSignalFuncDecl emits a function closure for signal-based components.
func writeSignalFuncDecl(buf *bytes.Buffer, fd script.FuncDecl) {
	if fd.Params == "" {
		fmt.Fprintf(buf, "\t%s := func() {\n", fd.Name)
	} else {
		fmt.Fprintf(buf, "\t%s := func(%s) {\n", fd.Name, fd.Params)
	}
	buf.WriteString(fd.Body)
	buf.WriteString("\n\t}\n")
}

// writeSignalTree builds the layout tree, using signal-aware content expressions.
func writeSignalTree(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, signals map[string]bool, ext *extractionCtx) {
	// Write extracted node declarations.
	var treeBuf bytes.Buffer
	writeSignalTreeChildren(&treeBuf, doc.Children, stylesheet, signals, ext, "\t\t\t")

	// Write extracted declarations first.
	buf.Write(ext.declBuf.Bytes())

	// Write root tree.
	buf.WriteString("\troot := &layout.Input{\n")
	buf.WriteString("\t\tKind: layout.KindBox,\n")
	buf.WriteString("\t\tCursorCol: -1,\n")
	buf.WriteString("\t\tCursorRow: -1,\n")
	buf.WriteString("\t\tChildren: []*layout.Input{\n")
	buf.Write(treeBuf.Bytes())
	buf.WriteString("\t\t},\n")
	buf.WriteString("\t}\n")
}

// writeSignalTreeChildren writes child nodes with signal-aware expressions.
func writeSignalTreeChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, signals map[string]bool, ext *extractionCtx, tabs string) {
	for _, child := range children {
		switch n := child.(type) {
		case *template.TextElement:
			writeSignalTextNode(buf, n, stylesheet, signals, ext, tabs)
		case *template.BoxElement:
			// TODO: full box support in signal mode
			_ = n
		}
	}
}

// writeSignalTextNode writes a text node, extracting expression nodes for signal sync.
func writeSignalTextNode(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, signals map[string]bool, ext *extractionCtx, tabs string) {
	if hasExprParts(n.Parts) {
		name := ext.nextNodeName()
		expr := contentExprSignals(n.Parts, signals)

		// Write declaration.
		fmt.Fprintf(&ext.declBuf, "\t%s := &layout.Input{\n", name)
		fmt.Fprintf(&ext.declBuf, "\t\tKind:    layout.KindText,\n")
		fmt.Fprintf(&ext.declBuf, "\t\tContent: %s,\n", expr)
		fmt.Fprintf(&ext.declBuf, "\t\tCursorCol: -1,\n")
		fmt.Fprintf(&ext.declBuf, "\t\tCursorRow: -1,\n")
		fmt.Fprintf(&ext.declBuf, "\t}\n")

		// Record sync.
		ext.nodes = append(ext.nodes, extractedNode{varName: name, syncExpr: expr})

		// Reference in tree.
		fmt.Fprintf(buf, "%s%s,\n", tabs, name)
	} else {
		fmt.Fprintf(buf, "%s{\n", tabs)
		fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
		fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, contentExpr(n.Parts))
		fmt.Fprintf(buf, "%s\tCursorCol: -1,\n", tabs)
		fmt.Fprintf(buf, "%s\tCursorRow: -1,\n", tabs)
		fmt.Fprintf(buf, "%s},\n", tabs)
	}
}

// writeSignalEffect emits an effect that syncs signal values to tree nodes.
func writeSignalEffect(buf *bytes.Buffer, ext *extractionCtx) {
	buf.WriteString("\n\tsignal.Effect(func() {\n")
	for _, n := range ext.nodes {
		fmt.Fprintf(buf, "\t\t%s.Content = %s\n", n.varName, n.syncExpr)
	}
	buf.WriteString("\t\tapp.Dirty = true\n")
	buf.WriteString("\t\tapp.Wake()\n")
	buf.WriteString("\t})\n")
}

// writeSignalRender emits the render function and app creation for signal mode.
func writeSignalRender(buf *bytes.Buffer, ext *extractionCtx) {
	buf.WriteString(`
	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		tree := layout.Layout(root, termW, termH)
		if prevTree == nil || termW != prevW || termH != prevH {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
			changes, _ := layout.DiffTrees(prevTree, tree)
			layout.ApplyChanges(os.Stdout, changes)
		}
		prevTree = tree
		prevW = termW
		prevH = termH
	}

	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			handleKey(evt)
		},
	}
	app.Run()
`)
}
