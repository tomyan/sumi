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
	writeAppOnEvent(buf, doc, sc, instances, scrollBoxes, inlined)
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
func writeAppOnEvent(buf *bytes.Buffer, doc *template.Document, sc *script.Script,
	instances []componentInstance, scrollBoxes []scrollableBox, inlined []inlinedStateful) {

	hasKeys := hasKeyHandlers(doc, instances, inlined)
	hasScroll := len(scrollBoxes) > 0

	if !hasKeys && !hasScroll {
		return
	}

	eventAware := buildEventAwareSet(sc, inlined)

	buf.WriteString("\t\tOnEvent: func(evt input.Event) {\n")

	if hasKeys {
		anyEventAware := len(eventAware) > 0
		if anyEventAware {
			writeEventAwareAutoQuit(buf, doc, sc, inlined, eventAware)
			writeEventAwareDispatch(buf, doc, sc, inlined, eventAware)
			writeZeroArgDispatch(buf, doc, sc, instances, inlined, eventAware)
		} else {
			writeAutoQuit(buf)
			buf.WriteString("\t\t\tif evt.Kind == input.EventKey {\n")
			writeOnkeyDispatchEvent(buf, doc, sc, eventAware)
			writeChildHandleKeyEvent(buf, instances)
			writeInlinedOnkeyDispatch(buf, inlined, eventAware)
			buf.WriteString("\t\t\t}\n")
		}
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

// buildEventAwareSet returns a set of handler names that accept event parameters.
func buildEventAwareSet(sc *script.Script, inlined []inlinedStateful) map[string]bool {
	set := make(map[string]bool)
	if sc != nil {
		for _, fd := range sc.FuncDecls {
			if fd.Params != "" {
				set[fd.Name] = true
			}
		}
	}
	for _, is := range inlined {
		if is.Instance.Info.Script == nil {
			continue
		}
		for _, fd := range is.Instance.Info.Script.FuncDecls {
			if fd.Params != "" {
				set[is.Prefix+fd.Name] = true
			}
		}
	}
	return set
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

// writeAutoQuit writes the auto-quit logic for zero-arg handler mode.
func writeAutoQuit(buf *bytes.Buffer) {
	buf.WriteString("\t\t\tif evt.Kind == input.EventKey && evt.Rune == 3 {\n")
	buf.WriteString("\t\t\t\tapp.Quit()\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif evt.Kind == input.EventSignal {\n")
	buf.WriteString("\t\t\t\tapp.Quit()\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
}

// writeEventAwareAutoQuit writes auto-quit only for zero-arg handlers when mixed with event-aware.
// If ALL handlers are event-aware, no auto-quit is emitted.
func writeEventAwareAutoQuit(buf *bytes.Buffer, doc *template.Document, sc *script.Script,
	inlined []inlinedStateful, eventAware map[string]bool) {

	// Check if there are any zero-arg handlers
	hasZeroArg := false
	if onkey := findRootOnkey(doc); onkey != "" && !eventAware[onkey] {
		hasZeroArg = true
	}
	for _, is := range inlined {
		if onkey := findChildOnkeyHandler(is.Instance.Info.Doc); onkey != "" {
			if !eventAware[is.Prefix+onkey] {
				hasZeroArg = true
			}
		}
	}
	if hasZeroArg {
		writeAutoQuit(buf)
	}
}

// writeEventAwareDispatch writes dispatch calls for event-aware handlers (no EventKey guard).
func writeEventAwareDispatch(buf *bytes.Buffer, doc *template.Document, sc *script.Script,
	inlined []inlinedStateful, eventAware map[string]bool) {

	if onkey := findRootOnkey(doc); onkey != "" && eventAware[onkey] {
		fmt.Fprintf(buf, "\t\t\t%s(evt)\n", onkey)
	}
	for _, is := range inlined {
		if onkey := findChildOnkeyHandler(is.Instance.Info.Doc); onkey != "" {
			prefixed := is.Prefix + onkey
			if eventAware[prefixed] {
				fmt.Fprintf(buf, "\t\t\t%s(evt)\n", prefixed)
			}
		}
	}
}

// writeZeroArgDispatch writes dispatch for zero-arg handlers inside EventKey guard.
func writeZeroArgDispatch(buf *bytes.Buffer, doc *template.Document, sc *script.Script,
	instances []componentInstance, inlined []inlinedStateful, eventAware map[string]bool) {

	hasZeroArg := false
	if onkey := findRootOnkey(doc); onkey != "" && !eventAware[onkey] {
		hasZeroArg = true
	}
	for _, inst := range instances {
		if inst.Info.Doc == nil && inst.Info.HasState {
			hasZeroArg = true
		}
	}
	for _, is := range inlined {
		if onkey := findChildOnkeyHandler(is.Instance.Info.Doc); onkey != "" {
			if !eventAware[is.Prefix+onkey] {
				hasZeroArg = true
			}
		}
	}

	if !hasZeroArg {
		return
	}

	buf.WriteString("\t\t\tif evt.Kind == input.EventKey {\n")
	writeOnkeyDispatchEvent(buf, doc, sc, eventAware)
	writeChildHandleKeyEvent(buf, instances)
	writeInlinedOnkeyDispatch(buf, inlined, eventAware)
	buf.WriteString("\t\t\t}\n")
}

// writeOnkeyDispatchEvent writes the root onkey handler call for event-based dispatch.
func writeOnkeyDispatchEvent(buf *bytes.Buffer, doc *template.Document, sc *script.Script, eventAware map[string]bool) {
	onkeyFunc := findRootOnkey(doc)
	if onkeyFunc != "" && !eventAware[onkeyFunc] {
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

// writeInlinedOnkeyDispatch writes inlined onkey handler calls in the event loop.
func writeInlinedOnkeyDispatch(buf *bytes.Buffer, inlined []inlinedStateful, eventAware map[string]bool) {
	for _, is := range inlined {
		onkey := findChildOnkeyHandler(is.Instance.Info.Doc)
		if onkey != "" {
			prefixed := is.Prefix + onkey
			if !eventAware[prefixed] {
				fmt.Fprintf(buf, "\t\t\t\t%s()\n", prefixed)
			}
		}
	}
}
