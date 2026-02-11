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
	scrollBoxes := findAllScrollableBoxes(doc, stylesheet)
	title := findTitleElement(doc)
	writeComponentInits(buf, instances)
	writeStateOrDirtyOnly(buf, sc)
	writeScrollStateDecls(buf, scrollBoxes)
	writeFuncClosures(buf, sc)
	writeRenderClosure(buf, doc, stylesheet, instances, scrollBoxes, title)
	writeSuppressUnusedFuncs(buf, doc, sc)
	writeTerminalSetup(buf, title, len(scrollBoxes) > 0)
	writeEventLoop(buf, doc, sc, instances, scrollBoxes)
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

// writeTreeAndSync writes the build-once layout tree and sync function at function scope.
// Expression text nodes are extracted as named variables; sync patches their Content.
func writeTreeAndSync(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, instances []componentInstance) *extractionCtx {
	ext := newExtractionCtx("")

	// Write tree to temp buffer (discovers extractions)
	var treeBuf bytes.Buffer
	writeLayoutTree(&treeBuf, doc, stylesheet, false, instances, ext)

	// Emit extracted node declarations before tree
	buf.Write(ext.declBuf.Bytes())

	// Emit tree at function scope
	buf.Write(treeBuf.Bytes())

	// Emit sync function
	writeSyncFunc(buf, ext)

	return ext
}

// writeSyncFunc writes the sync closure that patches expression nodes
// and rebuilds dynamic children.
func writeSyncFunc(buf *bytes.Buffer, ext *extractionCtx) {
	buf.WriteString("\tsync := func() {\n")
	for _, n := range ext.nodes {
		fmt.Fprintf(buf, "\t\t%s.Content = %s\n", n.varName, n.syncExpr)
	}
	buf.Write(ext.syncBuf.Bytes())
	buf.WriteString("\t}\n\n")
}

// writeRenderClosure writes the doRender closure with surgical rendering.
// The tree is built once at function scope; doRender calls sync() then re-layouts.
func writeRenderClosure(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, instances []componentInstance, scrollBoxes []scrollableBox, title *template.TitleElement) {
	writeTreeAndSync(buf, doc, stylesheet, instances)
	buf.WriteString("\tvar prevTree *layout.Box\n")
	buf.WriteString("\tvar prevW, prevH int\n")
	buf.WriteString("\tdoRender := func() {\n")
	buf.WriteString("\t\tsync()\n")
	buf.WriteString("\t\ttermW, termH := term.GetSize(int(os.Stdin.Fd()))\n")
	buf.WriteString("\t\ttree := layout.Layout(root, termW, termH)\n")
	writeScrollTreeWiring(buf, scrollBoxes)
	buf.WriteString("\t\tif prevTree == nil || termW != prevW || termH != prevH || layout.HasScrollChanged(prevTree, tree) || layout.HasOverlappingElements(tree) || layout.HasOverlappingElements(prevTree) {\n")
	buf.WriteString("\t\t\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\t\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\t\t\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\tchanges := layout.DiffTrees(prevTree, tree)\n")
	buf.WriteString("\t\t\tlayout.ApplyChanges(os.Stdout, changes)\n")
	buf.WriteString("\t\t}\n")
	writeTitleSet(buf, title)
	buf.WriteString("\t\tprevTree = tree\n")
	buf.WriteString("\t\tprevW = termW\n")
	buf.WriteString("\t\tprevH = termH\n")
	buf.WriteString("\t\tdirty = false\n")
	buf.WriteString("\t}\n\n")
}

// writeTerminalSetup writes raw mode, alternate screen, event channel, and resize watcher setup.
func writeTerminalSetup(buf *bytes.Buffer, title *template.TitleElement, hasScroll bool) {
	writeTitleSave(buf, title)
	buf.WriteString("\trestore, _ := input.EnableRawMode(int(os.Stdin.Fd()))\n")
	buf.WriteString("\tdefer restore()\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\tdefer render.ExitAlternateScreen(os.Stdout)\n")
	if hasScroll {
		buf.WriteString("\tfmt.Fprint(os.Stdout, input.MouseEnableSeq)\n")
		buf.WriteString("\tdefer fmt.Fprint(os.Stdout, input.MouseDisableSeq)\n")
	}
	writeTitleRestore(buf, title)
	buf.WriteString("\n")
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

// writeEventLoop writes the main select-based event loop.
func writeEventLoop(buf *bytes.Buffer, doc *template.Document, sc *script.Script, instances []componentInstance, scrollBoxes []scrollableBox) {
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tselect {\n")
	buf.WriteString("\t\tcase evt, ok := <-eventCh:\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	writeEventKeyHandler(buf, doc, instances)
	writeScrollDispatch(buf, scrollBoxes)
	writeMouseScrollDispatch(buf, scrollBoxes)
	buf.WriteString("\t\tcase <-resizeCh:\n")
	writeEnvUpdate(buf, sc)
	buf.WriteString("\t\t\tdirty = true\n")
	buf.WriteString("\t\t}\n")
	writeDirtyCheck(buf, instances)
	buf.WriteString("\t}\n")
}

// writeEventKeyHandler writes the handler for EventKey events (quit, onkey, child HandleKey).
func writeEventKeyHandler(buf *bytes.Buffer, doc *template.Document, instances []componentInstance) {
	buf.WriteString("\t\t\tif evt.Kind == input.EventKey {\n")
	buf.WriteString("\t\t\t\tif evt.Rune == 'q' || evt.Rune == 3 {\n")
	buf.WriteString("\t\t\t\t\treturn\n")
	buf.WriteString("\t\t\t\t}\n")
	writeOnkeyDispatchEvent(buf, doc)
	writeChildHandleKeyEvent(buf, instances)
	buf.WriteString("\t\t\t}\n")
}

// writeEnvUpdate writes env variable updates on resize.
func writeEnvUpdate(buf *bytes.Buffer, sc *script.Script) {
	if sc == nil || len(sc.EnvDecls) == 0 {
		return
	}
	wName, hName := envVarNames(sc.EnvDecls)
	fmt.Fprintf(buf, "\t\t\t%s, %s = term.GetSize(int(os.Stdin.Fd()))\n", wName, hName)
}

// writeOnkeyDispatchEvent writes the root onkey handler call for event-based dispatch.
func writeOnkeyDispatchEvent(buf *bytes.Buffer, doc *template.Document) {
	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc != "" {
		fmt.Fprintf(buf, "\t\t\t\t%s()\n", onkeyFunc)
	}
}

// writeChildHandleKeyEvent writes HandleKey dispatch using evt.Rune for event-based loop.
// Skips instances that are inlined (have Doc available).
func writeChildHandleKeyEvent(buf *bytes.Buffer, instances []componentInstance) {
	for _, inst := range instances {
		if inst.Info.Doc != nil {
			continue // inlined, handled differently
		}
		if inst.Info.HasState {
			fmt.Fprintf(buf, "\t\t\t\t%s.HandleKey(evt.Rune)\n", inst.VarName)
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

// buildDirtyCondition builds the dirty check expression including non-inlined children.
func buildDirtyCondition(instances []componentInstance) string {
	parts := []string{"dirty"}
	for _, inst := range instances {
		if inst.Info.Doc != nil {
			continue // inlined, no separate Dirty() needed
		}
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
	switch n := node.(type) {
	case *template.BoxElement:
		if handler, ok := n.Attributes["onkey"]; ok && handler == funcName {
			return true
		}
		for _, child := range n.Children {
			if boxHasOnkey(child, funcName) {
				return true
			}
		}
	case *template.IfNode:
		for _, child := range n.Then {
			if boxHasOnkey(child, funcName) {
				return true
			}
		}
		for _, child := range n.Else {
			if boxHasOnkey(child, funcName) {
				return true
			}
		}
	case *template.ForNode:
		for _, child := range n.Children {
			if boxHasOnkey(child, funcName) {
				return true
			}
		}
	}
	return false
}
