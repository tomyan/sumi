package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

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
func writeEventLoop(buf *bytes.Buffer, doc *template.Document, sc *script.Script, instances []componentInstance, scrollBoxes []scrollableBox, inlined []inlinedStateful) {
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tselect {\n")
	buf.WriteString("\t\tcase evt, ok := <-eventCh:\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	writeEventKeyHandler(buf, doc, instances, inlined)
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
func writeEventKeyHandler(buf *bytes.Buffer, doc *template.Document, instances []componentInstance, inlined []inlinedStateful) {
	buf.WriteString("\t\t\tif evt.Kind == input.EventKey {\n")
	buf.WriteString("\t\t\t\tif evt.Rune == 'q' || evt.Rune == 3 {\n")
	buf.WriteString("\t\t\t\t\treturn\n")
	buf.WriteString("\t\t\t\t}\n")
	writeOnkeyDispatchEvent(buf, doc)
	writeChildHandleKeyEvent(buf, instances)
	writeInlinedOnkeyDispatch(buf, inlined)
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
