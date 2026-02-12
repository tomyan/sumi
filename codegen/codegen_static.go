package codegen

import (
	"bytes"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeImports writes the import block for the generated code.
// hasEvents indicates whether the generated code references input.Event types
// (reactive path with event handlers).
func writeImports(buf *bytes.Buffer, hasExprs bool, hasEvents bool, hasTime bool) {
	buf.WriteString("import (\n")
	if hasExprs {
		buf.WriteString("\t\"fmt\"\n")
	}
	buf.WriteString("\t\"os\"\n")
	if hasTime {
		buf.WriteString("\t\"time\"\n")
	}
	buf.WriteString("\n")
	if hasEvents {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/input\"\n")
	}
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/layout\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/term\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/tui\"\n")
	buf.WriteString(")\n\n")
}

// writeStaticBody generates the static (non-reactive) function body.
// Static apps have no state but still handle terminal resize and quit via tui.App.
func writeStaticBody(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	writeLayoutTree(buf, doc, stylesheet, false, nil, nil)
	writeStaticRenderFunc(buf)
	buf.WriteString("\tapp := &tui.App{\n")
	buf.WriteString("\t\tOnRender: doRender,\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tapp.Run()\n")
}

// writeStaticRenderFunc writes the doRender closure for static apps.
func writeStaticRenderFunc(buf *bytes.Buffer) {
	buf.WriteString("\tdoRender := func() {\n")
	buf.WriteString("\t\ttermW, termH := term.GetSize(int(os.Stdin.Fd()))\n")
	buf.WriteString("\t\ttree := layout.Layout(root, termW, termH)\n")
	buf.WriteString("\t\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\t\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t}\n\n")
}
