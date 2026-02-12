package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// writeAppDecl writes the forward declaration of the app variable.
// This must be emitted before any closures that reference app.Dirty.
func writeAppDecl(buf *bytes.Buffer) {
	buf.WriteString("\tvar app *tui.App\n")
}

// writeAppRun writes the tui.App construction and app.Run() call.
func writeAppRun(buf *bytes.Buffer, doc *template.Document, sc *script.Script,
	instances []componentInstance, scrollBoxes []scrollableBox, inlined []inlinedStateful, title *template.TitleElement) {

	buf.WriteString("\tapp = &tui.App{\n")

	if len(scrollBoxes) > 0 {
		buf.WriteString("\t\tHasMouse: true,\n")
	}

	writeAppTitle(buf, title)
	buf.WriteString("\t\tOnRender: doRender,\n")
	writeAppOnEvent(buf, doc, instances, scrollBoxes, inlined)
	writeAppOnResize(buf, sc)

	buf.WriteString("\t}\n")
	buf.WriteString("\tapp.Run()\n")
}

// writeAppTitle writes the Title or SaveTitle field if a title element exists.
func writeAppTitle(buf *bytes.Buffer, title *template.TitleElement) {
	if title == nil {
		return
	}
	if isStaticTitle(title) {
		content := buildStaticTitleString(title)
		fmt.Fprintf(buf, "\t\tTitle:     %q,\n", content)
	} else {
		buf.WriteString("\t\tSaveTitle: true,\n")
	}
}

// writeAppOnEvent writes the OnEvent closure for the App.
func writeAppOnEvent(buf *bytes.Buffer, doc *template.Document,
	instances []componentInstance, scrollBoxes []scrollableBox, inlined []inlinedStateful) {

	hasKeys := hasKeyHandlers(doc, instances, inlined)
	hasScroll := len(scrollBoxes) > 0

	if !hasKeys && !hasScroll {
		return
	}

	buf.WriteString("\t\tOnEvent: func(evt input.Event) {\n")

	if hasKeys {
		buf.WriteString("\t\t\tif evt.Kind == input.EventKey {\n")
		writeOnkeyDispatchEvent(buf, doc)
		writeChildHandleKeyEvent(buf, instances)
		writeInlinedOnkeyDispatch(buf, inlined)
		buf.WriteString("\t\t\t}\n")
	}

	writeScrollDispatch(buf, scrollBoxes)
	writeMouseScrollDispatch(buf, scrollBoxes)

	buf.WriteString("\t\t},\n")
}

// writeAppOnResize writes the OnResize closure if env decls exist.
func writeAppOnResize(buf *bytes.Buffer, sc *script.Script) {
	if sc == nil || len(sc.EnvDecls) == 0 {
		return
	}
	wName, hName := envVarNames(sc.EnvDecls)
	buf.WriteString("\t\tOnResize: func() {\n")
	fmt.Fprintf(buf, "\t\t\t%s, %s = term.GetSize(int(os.Stdin.Fd()))\n", wName, hName)
	buf.WriteString("\t\t},\n")
}

// hasKeyHandlers returns true if there are any key event handlers to dispatch.
func hasKeyHandlers(doc *template.Document, instances []componentInstance, inlined []inlinedStateful) bool {
	if findRootOnkey(doc) != "" {
		return true
	}
	for _, inst := range instances {
		if inst.Info.Doc != nil {
			continue
		}
		if inst.Info.HasState {
			return true
		}
	}
	for _, is := range inlined {
		if findChildOnkeyHandler(is.Instance.Info.Doc) != "" {
			return true
		}
	}
	return false
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

