package codegen

import "bytes"

// writeTermSizeWithTestMode emits code that gets terminal dimensions,
// using test fields when available, falling back to term.GetSize.
func writeTermSizeWithTestMode(buf *bytes.Buffer) {
	buf.WriteString("\t\tvar termW, termH int\n")
	buf.WriteString("\t\tif app.TestWidth > 0 {\n")
	buf.WriteString("\t\t\ttermW, termH = app.TestWidth, app.TestHeight\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\ttermW, termH = term.GetSize(int(os.Stdin.Fd()))\n")
	buf.WriteString("\t\t}\n")
}

// writeBufferOutputWithTestMode emits code that either populates app.TestBuffer
// or writes to stdout, depending on test mode.
func writeBufferOutputWithTestMode(buf *bytes.Buffer) {
	buf.WriteString("\t\tif app.TestBuffer != nil {\n")
	buf.WriteString("\t\t\tapp.TestBuffer = buf\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t}\n")
}

// writeFullRedrawWithTestMode emits the full redraw path with test-mode branch.
// The surgical apply path is only used in non-test mode.
func writeFullRedrawWithTestMode(buf *bytes.Buffer) {
	buf.WriteString("\t\tif prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {\n")
	buf.WriteString("\t\t\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\t\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\t\t\tif app.TestBuffer != nil {\n")
	buf.WriteString("\t\t\t\tapp.TestBuffer = buf\n")
	buf.WriteString("\t\t\t} else {\n")
	buf.WriteString("\t\t\t\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\t\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t} else if app.TestBuffer != nil {\n")
	buf.WriteString("\t\t\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\t\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\t\t\tapp.TestBuffer = buf\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\tlayout.ApplyChanges(os.Stdout, changes)\n")
	buf.WriteString("\t\t}\n")
}

// writeCursorPositioningWithTestMode emits cursor show/hide with test-mode guard.
func writeCursorPositioningWithTestMode(buf *bytes.Buffer) {
	buf.WriteString("\t\tif app.TestBuffer == nil {\n")
	buf.WriteString("\t\t\tif cursorBox := layout.FindCursor(tree); cursorBox != nil {\n")
	buf.WriteString("\t\t\t\trender.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)\n")
	buf.WriteString("\t\t\t} else {\n")
	buf.WriteString("\t\t\t\trender.HideCursor(os.Stdout)\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t}\n")
}

// writeDirectWriteWithTestMode emits the direct-write fast path with test-mode guard.
func writeDirectWriteWithTestMode(buf *bytes.Buffer) {
	buf.WriteString("\t\tif app.TestBuffer == nil && prevTree != nil && len(changed) > 0 && termW == prevW && termH == prevH && !prevTree.HasOverlap && nodeBoxMap != nil {\n")
	buf.WriteString("\t\t\tallDirect := true\n")
	buf.WriteString("\t\t\tfor _, inp := range changed {\n")
	buf.WriteString("\t\t\t\tbox := nodeBoxMap[inp]\n")
	buf.WriteString("\t\t\t\tif !layout.DirectWriteText(os.Stdout, box, inp.Content, box.Content) {\n")
	buf.WriteString("\t\t\t\t\tallDirect = false\n")
	buf.WriteString("\t\t\t\t\tbreak\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tbox.Content = inp.Content\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif allDirect {\n")
	buf.WriteString("\t\t\t\treturn\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t}\n")
}
