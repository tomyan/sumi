package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeReactiveBody generates the reactive function body with event loop.
func writeReactiveBody(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet) {
	writeStateDecls(buf, sc.StateDecls)
	writeDirtyFlagAndFuncClosures(buf, sc.FuncDecls)
	writeRenderClosure(buf, doc, stylesheet)
	writeSuppressUnused(buf, doc, sc.FuncDecls)
	writeTerminalSetup(buf)
	writeEventLoop(buf, doc)
}

// writeStateDecls writes state variable declarations.
func writeStateDecls(buf *bytes.Buffer, stateDecls []script.StateDecl) {
	for _, stateDecl := range stateDecls {
		fmt.Fprintf(buf, "\t%s := %s\n", stateDecl.Name, stateDecl.InitExpr)
	}
	buf.WriteString("\n")
}

// writeDirtyFlagAndFuncClosures writes the dirty flag and function closure declarations.
func writeDirtyFlagAndFuncClosures(buf *bytes.Buffer, funcDecls []script.FuncDecl) {
	buf.WriteString("\tdirty := true\n")
	for _, funcDecl := range funcDecls {
		fmt.Fprintf(buf, "\t%s := func() {\n", funcDecl.Name)
		writeReactiveFuncBody(buf, funcDecl)
		buf.WriteString("\t}\n")
	}
	buf.WriteString("\n")
}

// writeReactiveFuncBody writes a function body, adding dirty=true after each state assignment.
func writeReactiveFuncBody(buf *bytes.Buffer, funcDecl script.FuncDecl) {
	stateLines := buildStateLinesSet(funcDecl.StateAssignments)
	for _, line := range strings.Split(funcDecl.Body, "\n") {
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

// buildStateLinesSet builds a set of lines that are state assignments.
func buildStateLinesSet(assignments []script.StateAssignment) map[string]bool {
	set := make(map[string]bool, len(assignments))
	for _, sa := range assignments {
		set[sa.Line] = true
	}
	return set
}

// writeRenderClosure writes the doRender closure.
func writeRenderClosure(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet) {
	buf.WriteString("\tvar prevBuf *render.Buffer\n")
	buf.WriteString("\tdoRender := func() {\n")
	writeLayoutTree(buf, doc, stylesheet, true)
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
}

// writeSuppressUnused writes _ = funcName for functions not referenced in onkey handlers.
func writeSuppressUnused(buf *bytes.Buffer, doc *template.Document, funcDecls []script.FuncDecl) {
	for _, funcDecl := range funcDecls {
		if !docHasOnkey(doc, funcDecl.Name) {
			fmt.Fprintf(buf, "\t_ = %s\n", funcDecl.Name)
		}
	}
}

// writeTerminalSetup writes raw mode and alternate screen setup.
func writeTerminalSetup(buf *bytes.Buffer) {
	buf.WriteString("\trestore, _ := input.EnableRawMode(int(os.Stdin.Fd()))\n")
	buf.WriteString("\tdefer restore()\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\tdefer render.ExitAlternateScreen(os.Stdout)\n\n")
	buf.WriteString("\tdoRender()\n\n")
}

// writeEventLoop writes the main event loop.
func writeEventLoop(buf *bytes.Buffer, doc *template.Document) {
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tkey, err := input.ReadKey(os.Stdin)\n")
	buf.WriteString("\t\tif err != nil || key == 'q' {\n")
	buf.WriteString("\t\t\tbreak\n")
	buf.WriteString("\t\t}\n")

	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc != "" {
		fmt.Fprintf(buf, "\t\t%s()\n", onkeyFunc)
	}

	buf.WriteString("\t\tif dirty {\n")
	buf.WriteString("\t\t\tdoRender()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
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

// boxHasOnkey recursively checks if a node has an onkey attribute matching funcName.
func boxHasOnkey(node template.Node, funcName string) bool {
	box, ok := node.(*template.BoxElement)
	if !ok {
		return false
	}
	if handler, ok := box.Attributes["onkey"]; ok && handler == funcName {
		return true
	}
	for _, child := range box.Children {
		if boxHasOnkey(child, funcName) {
			return true
		}
	}
	return false
}
