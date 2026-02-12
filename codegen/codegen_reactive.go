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
	inlined := collectInlinedStateful(instances)
	writeComponentInits(buf, instances)
	writeStateOrDirtyOnly(buf, sc)
	writeInlinedStateDecls(buf, inlined)
	writeScrollStateDecls(buf, scrollBoxes)
	writeFuncClosures(buf, sc)
	writeInlinedFuncClosures(buf, inlined)
	if len(inlined) > 0 {
		buf.WriteString("\n")
	}
	writeRenderClosure(buf, doc, stylesheet, instances, scrollBoxes, title)
	writeSuppressUnusedFuncs(buf, doc, sc)
	writeSuppressInlinedFuncs(buf, inlined)
	writeTerminalSetup(buf, title, len(scrollBoxes) > 0)
	writeEventLoop(buf, doc, sc, instances, scrollBoxes, inlined)
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
