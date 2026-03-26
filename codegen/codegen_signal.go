package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeSignalBody generates the body of Run()/CreateApp() using the signal-based reactive model.
func writeSignalBody(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet) {
	writeSignalBodyInner(buf, doc, sc, stylesheet, false)
}

func writeSignalBodyInner(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, isCreateApp bool) {
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

	// Build layout tree using existing infrastructure with signal-aware extraction.
	ext := newExtractionCtx("")
	ext.signals = signals
	var treeBuf bytes.Buffer
	writeLayoutTree(&treeBuf, doc, stylesheet, false, nil, ext)

	// Emit extracted declarations before the tree (so node0 etc. are defined).
	buf.Write(ext.declBuf.Bytes())
	buf.Write(treeBuf.Bytes())

	// Emit effect that syncs signal values to tree nodes.
	if ext.hasSyncContent() {
		writeSignalEffect(buf, ext)
	}

	// Emit render function and app setup.
	writeSignalRender(buf, sc, isCreateApp)
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

// writeSignalEffect emits an effect that syncs signal values to tree nodes.
func writeSignalEffect(buf *bytes.Buffer, ext *extractionCtx) {
	buf.WriteString("\n\tsignal.Effect(func() {\n")
	for _, n := range ext.nodes {
		fmt.Fprintf(buf, "\t\t%s.Content = %s\n", n.varName, n.syncExpr)
	}
	buf.WriteString("\t\tif app != nil {\n")
	buf.WriteString("\t\t\tapp.Dirty = true\n")
	buf.WriteString("\t\t\tapp.Wake()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t})\n")
}

// writeSignalCreateAppBody generates the CreateApp body for signal mode.
func writeSignalCreateAppBody(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet) {
	writeSignalBodyInner(buf, doc, sc, stylesheet, true)
}

// writeSignalRender emits the render function and app creation for signal mode.
func writeSignalRender(buf *bytes.Buffer, sc *script.Script, isCreateApp bool) {
	// Check if there's a handleKey function.
	hasHandler := false
	for _, fd := range sc.FuncDecls {
		if fd.Name == "handleKey" {
			hasHandler = true
			break
		}
	}

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
`)

	if hasHandler {
		buf.WriteString("\tapp = &tui.App{\n")
		buf.WriteString("\t\tOnRender: doRender,\n")
		buf.WriteString("\t\tOnEvent: func(evt input.Event) {\n")
		buf.WriteString("\t\t\thandleKey(evt)\n")
		buf.WriteString("\t\t},\n")
		buf.WriteString("\t}\n")
	} else {
		buf.WriteString("\tapp = &tui.App{\n")
		buf.WriteString("\t\tOnRender: doRender,\n")
		buf.WriteString("\t}\n")
	}
	if isCreateApp {
		buf.WriteString("\treturn app\n")
	} else {
		buf.WriteString("\tapp.Run()\n")
	}
}
