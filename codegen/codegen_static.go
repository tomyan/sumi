package codegen

import (
	"bytes"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeImports writes the import block for the generated code.
func writeImports(buf *bytes.Buffer, hasExprs bool) {
	buf.WriteString("import (\n")
	if hasExprs {
		buf.WriteString("\t\"fmt\"\n")
	}
	buf.WriteString("\t\"os\"\n\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/input\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/layout\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/term\"\n")
	buf.WriteString(")\n\n")
}

// writeStaticBody generates the static (non-reactive) function body.
// Static apps have no state but still handle terminal resize and quit keys.
func writeStaticBody(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	writeLayoutTree(buf, doc, stylesheet, false, nil)
	writeStaticRenderFunc(buf)
	writeStaticTerminalSetup(buf)
	writeStaticEventLoop(buf)
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

// writeStaticTerminalSetup writes raw mode, alternate screen, event channel, and resize watcher.
func writeStaticTerminalSetup(buf *bytes.Buffer) {
	buf.WriteString("\trestore, _ := input.EnableRawMode(int(os.Stdin.Fd()))\n")
	buf.WriteString("\tdefer restore()\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\tdefer render.ExitAlternateScreen(os.Stdout)\n\n")
	buf.WriteString("\teventCh := make(chan input.Event)\n")
	buf.WriteString("\tgo func() {\n")
	buf.WriteString("\t\tfor {\n")
	buf.WriteString("\t\t\tevt, err := input.ReadEvent(os.Stdin)\n")
	buf.WriteString("\t\t\tif err != nil {\n")
	buf.WriteString("\t\t\t\tclose(eventCh)\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\teventCh <- evt\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}()\n\n")
	buf.WriteString("\tresizeCh, stopResize := term.WatchResize()\n")
	buf.WriteString("\tdefer stopResize()\n\n")
	buf.WriteString("\tdoRender()\n\n")
}

// writeStaticEventLoop writes the select-based event loop for static apps.
func writeStaticEventLoop(buf *bytes.Buffer) {
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tselect {\n")
	buf.WriteString("\t\tcase evt, ok := <-eventCh:\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif evt.Kind == input.EventKey {\n")
	buf.WriteString("\t\t\t\tif evt.Rune == 'q' || evt.Rune == 3 {\n")
	buf.WriteString("\t\t\t\t\treturn\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\tcase <-resizeCh:\n")
	buf.WriteString("\t\t\tdoRender()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
}
