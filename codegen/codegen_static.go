package codegen

import (
	"bytes"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeImports writes the import block for the generated code.
func writeImports(buf *bytes.Buffer, hasExprs bool, hasEvents bool, hasTime bool) {
	buf.WriteString("import (\n")
	buf.WriteString("\t\"os\"\n")
	if hasTime {
		buf.WriteString("\t\"time\"\n")
	}
	buf.WriteString("\n")
	buf.WriteString("\tsumi \"github.com/tomyan/sumi/runtime/prelude\"\n")
	buf.WriteString(")\n\n")
}

// writeStaticBody generates the static (non-reactive) function body.
// Static apps have no state but still handle terminal resize and quit via sumi.App.
func writeStaticBody(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	writeStaticSharedSetup(buf, doc, stylesheet)
	buf.WriteString("\tapp.Run()\n")
}

// writeStaticCreateAppBody generates the static CreateApp body.
func writeStaticCreateAppBody(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	writeStaticSharedSetup(buf, doc, stylesheet)
	writeCreateAppReturn(buf)
}

// writeStaticSharedSetup writes the shared setup for static Run and CreateApp.
func writeStaticSharedSetup(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	buf.WriteString("\tvar app *sumi.App\n")
	writeLayoutTree(buf, doc, stylesheet, false, nil)
	writeStaticRenderFunc(buf)
	buf.WriteString("\tapp = &sumi.App{\n")
	buf.WriteString("\t\tOnRender: doRender,\n")
	buf.WriteString("\t}\n")
}

// writeStaticRenderFunc writes the doRender closure for static apps.
func writeStaticRenderFunc(buf *bytes.Buffer) {
	buf.WriteString("\tdoRender := func() {\n")
	writeTermSizeWithTestMode(buf)
	buf.WriteString("\t\ttree := sumi.Layout(root, termW, termH)\n")
	buf.WriteString("\t\tbuf := sumi.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\tsumi.RenderTree(buf, tree, nil)\n")
	writeBufferOutputWithTestMode(buf)
	buf.WriteString("\t}\n\n")
}
