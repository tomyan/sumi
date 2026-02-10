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
func writeReactiveBody(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, instances []componentInstance) {
	writeComponentInits(buf, instances)
	writeStateOrDirtyOnly(buf, sc)
	writeFuncClosures(buf, sc)
	writeRenderClosure(buf, doc, stylesheet, instances)
	writeSuppressUnusedFuncs(buf, doc, sc)
	writeTerminalSetup(buf)
	writeEventLoop(buf, doc, sc, instances)
}

// writeStateOrDirtyOnly writes state decls and env decls if present, then the dirty flag.
func writeStateOrDirtyOnly(buf *bytes.Buffer, sc *script.Script) {
	if sc != nil && len(sc.StateDecls) > 0 {
		writeStateDecls(buf, sc.StateDecls)
	}
	if sc != nil && len(sc.EnvDecls) > 0 {
		writeEnvDecls(buf, sc.EnvDecls)
	}
	buf.WriteString("\tdirty := true\n")
}

// writeFuncClosures writes function closure declarations if present.
func writeFuncClosures(buf *bytes.Buffer, sc *script.Script) {
	if sc == nil || len(sc.FuncDecls) == 0 {
		buf.WriteString("\n")
		return
	}
	for _, funcDecl := range sc.FuncDecls {
		fmt.Fprintf(buf, "\t%s := func() {\n", funcDecl.Name)
		writeReactiveFuncBody(buf, funcDecl)
		buf.WriteString("\t}\n")
	}
	buf.WriteString("\n")
}

// writeSuppressUnusedFuncs writes _ = funcName for functions not referenced in onkey handlers.
func writeSuppressUnusedFuncs(buf *bytes.Buffer, doc *template.Document, sc *script.Script) {
	if sc == nil {
		return
	}
	for _, funcDecl := range sc.FuncDecls {
		if !docHasOnkey(doc, funcDecl.Name) {
			fmt.Fprintf(buf, "\t_ = %s\n", funcDecl.Name)
		}
	}
}

// writeStateDecls writes state variable declarations.
func writeStateDecls(buf *bytes.Buffer, stateDecls []script.StateDecl) {
	for _, stateDecl := range stateDecls {
		fmt.Fprintf(buf, "\t%s := %s\n", stateDecl.Name, stateDecl.InitExpr)
	}
	buf.WriteString("\n")
}

// writeEnvDecls writes env variable initialization from term.GetSize.
func writeEnvDecls(buf *bytes.Buffer, envDecls []script.EnvDecl) {
	wName, hName := envVarNames(envDecls)
	fmt.Fprintf(buf, "\t%s, %s := term.GetSize(int(os.Stdin.Fd()))\n", wName, hName)
}

// envVarNames returns the variable names for width and height from env decls.
// If a key is not declared, returns "_" for that position.
func envVarNames(envDecls []script.EnvDecl) (widthName, heightName string) {
	widthName = "_"
	heightName = "_"
	for _, e := range envDecls {
		switch e.Key {
		case "width":
			widthName = e.Name
		case "height":
			heightName = e.Name
		}
	}
	return
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
func writeRenderClosure(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, instances []componentInstance) {
	buf.WriteString("\tvar prevBuf *render.Buffer\n")
	buf.WriteString("\tdoRender := func() {\n")
	buf.WriteString("\t\ttermW, termH := term.GetSize(int(os.Stdin.Fd()))\n")
	writeLayoutTree(buf, doc, stylesheet, true, instances)
	buf.WriteString("\t\ttree := layout.Layout(root, termW, termH)\n")
	buf.WriteString("\t\tbuf := render.NewBuffer(termW, termH)\n")
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

// writeTerminalSetup writes raw mode, alternate screen, key channel, and resize watcher setup.
func writeTerminalSetup(buf *bytes.Buffer) {
	buf.WriteString("\trestore, _ := input.EnableRawMode(int(os.Stdin.Fd()))\n")
	buf.WriteString("\tdefer restore()\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\tdefer render.ExitAlternateScreen(os.Stdout)\n\n")
	buf.WriteString("\tkeyCh := make(chan rune)\n")
	buf.WriteString("\tgo func() {\n")
	buf.WriteString("\t\tfor {\n")
	buf.WriteString("\t\t\tkey, err := input.ReadKey(os.Stdin)\n")
	buf.WriteString("\t\t\tif err != nil {\n")
	buf.WriteString("\t\t\t\tclose(keyCh)\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tkeyCh <- key\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}()\n\n")
	buf.WriteString("\tresizeCh, stopResize := term.WatchResize()\n")
	buf.WriteString("\tdefer stopResize()\n\n")
	buf.WriteString("\tdoRender()\n\n")
}

// writeEventLoop writes the main select-based event loop.
func writeEventLoop(buf *bytes.Buffer, doc *template.Document, sc *script.Script, instances []componentInstance) {
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tselect {\n")
	buf.WriteString("\t\tcase key, ok := <-keyCh:\n")
	buf.WriteString("\t\t\tif !ok || key == 'q' {\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	writeOnkeyDispatchIndented(buf, doc)
	writeChildHandleKeyIndented(buf, instances)
	buf.WriteString("\t\tcase <-resizeCh:\n")
	writeEnvUpdate(buf, sc)
	buf.WriteString("\t\t\tdirty = true\n")
	buf.WriteString("\t\t}\n")
	writeDirtyCheck(buf, instances)
	buf.WriteString("\t}\n")
}

// writeEnvUpdate writes env variable updates on resize.
func writeEnvUpdate(buf *bytes.Buffer, sc *script.Script) {
	if sc == nil || len(sc.EnvDecls) == 0 {
		return
	}
	wName, hName := envVarNames(sc.EnvDecls)
	fmt.Fprintf(buf, "\t\t\t%s, %s = term.GetSize(int(os.Stdin.Fd()))\n", wName, hName)
}

// writeOnkeyDispatch writes the root onkey handler call if present.
func writeOnkeyDispatch(buf *bytes.Buffer, doc *template.Document) {
	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc != "" {
		fmt.Fprintf(buf, "\t\t%s()\n", onkeyFunc)
	}
}

// writeOnkeyDispatchIndented writes the root onkey handler call indented for select case.
func writeOnkeyDispatchIndented(buf *bytes.Buffer, doc *template.Document) {
	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc != "" {
		fmt.Fprintf(buf, "\t\t\t%s()\n", onkeyFunc)
	}
}

// writeChildHandleKey writes HandleKey dispatch to stateful child components.
func writeChildHandleKey(buf *bytes.Buffer, instances []componentInstance) {
	for _, inst := range instances {
		if inst.Info.HasState {
			fmt.Fprintf(buf, "\t\t%s.HandleKey(key)\n", inst.VarName)
		}
	}
}

// writeChildHandleKeyIndented writes HandleKey dispatch indented for select case.
func writeChildHandleKeyIndented(buf *bytes.Buffer, instances []componentInstance) {
	for _, inst := range instances {
		if inst.Info.HasState {
			fmt.Fprintf(buf, "\t\t\t%s.HandleKey(key)\n", inst.VarName)
		}
	}
}

// writeDirtyCheck writes the dirty check including child component dirty flags.
func writeDirtyCheck(buf *bytes.Buffer, instances []componentInstance) {
	condition := buildDirtyCondition(instances)
	fmt.Fprintf(buf, "\t\tif %s {\n", condition)
	buf.WriteString("\t\t\tdoRender()\n")
	buf.WriteString("\t\t}\n")
}

// buildDirtyCondition builds the dirty check expression including children.
func buildDirtyCondition(instances []componentInstance) string {
	parts := []string{"dirty"}
	for _, inst := range instances {
		parts = append(parts, fmt.Sprintf("%s.Dirty()", inst.VarName))
	}
	return strings.Join(parts, " || ")
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
