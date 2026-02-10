package codegen

import (
	"bytes"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeImports writes the import block for the generated code.
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
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/term\"\n")
	buf.WriteString(")\n\n")
}

// writeStaticBody generates the static (non-reactive) function body.
func writeStaticBody(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	writeLayoutTree(buf, doc, stylesheet, false, nil)

	buf.WriteString("\ttermW, termH := term.GetSize(int(os.Stdin.Fd()))\n")
	buf.WriteString("\ttree := layout.Layout(root, termW, termH)\n")
	buf.WriteString("\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\tbufio.NewScanner(os.Stdin).Scan()\n")
	buf.WriteString("\trender.ExitAlternateScreen(os.Stdout)\n")
}
