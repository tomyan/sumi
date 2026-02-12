package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeTreeAndSync writes the build-once layout tree and sync function at function scope.
// Expression text nodes are extracted as named variables; sync patches their Content.
func writeTreeAndSync(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, instances []componentInstance, scrollBoxes []scrollableBox, derivedDecls []script.DerivedDecl, selfDecls []script.SelfDecl) *extractionCtx {
	ext := newExtractionCtx("")

	// Write tree to temp buffer (discovers extractions)
	var treeBuf bytes.Buffer
	writeLayoutTree(&treeBuf, doc, stylesheet, false, instances, ext)

	// Emit extracted node declarations before tree
	buf.Write(ext.declBuf.Bytes())

	// Emit tree at function scope
	buf.Write(treeBuf.Bytes())

	// Wire self-measurement pointers on root
	writeSelfWiring(buf, selfDecls, "root")

	// Emit sync function
	dynamic := isDynamic(ext, scrollBoxes)
	writeSyncFunc(buf, ext, dynamic, derivedDecls)

	return ext
}

// writeSelfPrevDecls emits tracking variables for self-measurement change detection.
func writeSelfPrevDecls(buf *bytes.Buffer, selfDecls []script.SelfDecl) {
	for _, sd := range selfDecls {
		fmt.Fprintf(buf, "\tvar prev%s int\n", capitalizeFirst(sd.Name))
	}
}

// writeSelfChangeDetection emits self-width/height change detection after layout.
func writeSelfChangeDetection(buf *bytes.Buffer, selfDecls []script.SelfDecl) {
	for _, sd := range selfDecls {
		prevName := "prev" + capitalizeFirst(sd.Name)
		fmt.Fprintf(buf, "\t\tif %s != %s {\n", sd.Name, prevName)
		fmt.Fprintf(buf, "\t\t\t%s = %s\n", prevName, sd.Name)
		fmt.Fprintf(buf, "\t\t\tapp.Dirty = true\n")
		buf.WriteString("\t\t}\n")
	}
}

// capitalizeFirst returns the string with its first letter uppercased.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// writeInlinedSelfWiring writes self pointer wiring for an inlined component's root box.
func writeInlinedSelfWiring(buf *bytes.Buffer, selfDecls []script.SelfDecl, boxName, prefix string) {
	for _, sd := range selfDecls {
		switch sd.Key {
		case "width":
			fmt.Fprintf(buf, "\t%s.SelfW = &%s%s\n", boxName, prefix, sd.Name)
		case "height":
			fmt.Fprintf(buf, "\t%s.SelfH = &%s%s\n", boxName, prefix, sd.Name)
		}
	}
}

// rootOnlySelfDecls returns the self decls from the root script only (not inlined components).
func rootOnlySelfDecls(sc *script.Script) []script.SelfDecl {
	if sc == nil {
		return nil
	}
	return sc.SelfDecls
}

// collectAllSelfDecls gathers self decls from root and inlined components, with namespacing.
func collectAllSelfDecls(sc *script.Script, inlined []inlinedStateful) []script.SelfDecl {
	var all []script.SelfDecl
	if sc != nil {
		all = append(all, sc.SelfDecls...)
	}
	for _, is := range inlined {
		childSc := is.Instance.Info.Script
		if childSc == nil {
			continue
		}
		for _, sd := range childSc.SelfDecls {
			all = append(all, script.SelfDecl{
				Name: is.Prefix + sd.Name,
				Key:  sd.Key,
			})
		}
	}
	return all
}

// writeSelfWiring writes pointer assignments for self-measurement decls on a named box.
func writeSelfWiring(buf *bytes.Buffer, selfDecls []script.SelfDecl, boxName string) {
	for _, sd := range selfDecls {
		switch sd.Key {
		case "width":
			fmt.Fprintf(buf, "\t%s.SelfW = &%s\n", boxName, sd.Name)
		case "height":
			fmt.Fprintf(buf, "\t%s.SelfH = &%s\n", boxName, sd.Name)
		}
	}
}

// isDynamic returns true when the document has control flow or scroll containers,
// requiring the full Layout+Diff path on every render.
func isDynamic(ext *extractionCtx, scrollBoxes []scrollableBox) bool {
	return ext.syncBuf.Len() > 0 || len(scrollBoxes) > 0
}

// writeSyncFunc writes the sync closure. Static documents get a returning sync
// that compares before assigning and returns changed nodes. Dynamic documents
// get a void sync that always patches. Derived values are recalculated first.
func writeSyncFunc(buf *bytes.Buffer, ext *extractionCtx, dynamic bool, derivedDecls []script.DerivedDecl) {
	if dynamic {
		writeVoidSync(buf, ext, derivedDecls)
	} else {
		writeReturningSync(buf, ext, derivedDecls)
	}
}

// writeVoidSync writes a sync closure that unconditionally patches all nodes.
func writeVoidSync(buf *bytes.Buffer, ext *extractionCtx, derivedDecls []script.DerivedDecl) {
	buf.WriteString("\tsync := func() {\n")
	writeDerivedRecalc(buf, derivedDecls)
	for _, n := range ext.nodes {
		fmt.Fprintf(buf, "\t\t%s.Content = %s\n", n.varName, n.syncExpr)
	}
	buf.Write(ext.syncBuf.Bytes())
	buf.WriteString("\t}\n\n")
}

// writeReturningSync writes a sync closure that compares before assigning
// and returns a slice of changed Input nodes (nil when nothing changed).
func writeReturningSync(buf *bytes.Buffer, ext *extractionCtx, derivedDecls []script.DerivedDecl) {
	buf.WriteString("\tsync := func() []*layout.Input {\n")
	writeDerivedRecalc(buf, derivedDecls)
	buf.WriteString("\t\tvar changed []*layout.Input\n")
	for _, n := range ext.nodes {
		fmt.Fprintf(buf, "\t\tif v := %s; v != %s.Content {\n", n.syncExpr, n.varName)
		fmt.Fprintf(buf, "\t\t\t%s.Content = v\n", n.varName)
		fmt.Fprintf(buf, "\t\t\tchanged = append(changed, %s)\n", n.varName)
		buf.WriteString("\t\t}\n")
	}
	buf.WriteString("\t\treturn changed\n")
	buf.WriteString("\t}\n\n")
}

// collectAllDerivedDecls gathers derived decls from the root script and all inlined components,
// with inlined derived decls namespaced (e.g., counter0_doubled = counter0_count * 2).
func collectAllDerivedDecls(sc *script.Script, inlined []inlinedStateful) []script.DerivedDecl {
	var all []script.DerivedDecl
	if sc != nil {
		all = append(all, sc.DerivedDecls...)
	}
	for _, is := range inlined {
		childSc := is.Instance.Info.Script
		if childSc == nil {
			continue
		}
		for _, dd := range childSc.DerivedDecls {
			all = append(all, script.DerivedDecl{
				Name: is.Prefix + dd.Name,
				Expr: namespaceDerivedExpr(dd.Expr, childSc, is.Prefix),
			})
		}
	}
	return all
}

// writeDerivedRecalc writes derived value reassignments at the top of a sync closure.
func writeDerivedRecalc(buf *bytes.Buffer, derivedDecls []script.DerivedDecl) {
	for _, dd := range derivedDecls {
		fmt.Fprintf(buf, "\t\t%s = %s\n", dd.Name, dd.Expr)
	}
}

// writeRenderClosure writes the doRender closure with surgical rendering.
// The tree is built once at function scope; doRender calls sync() then re-layouts.
// Static documents get a no-op skip fast path; dynamic documents always run full Layout+Diff.
func writeRenderClosure(buf *bytes.Buffer, doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, instances []componentInstance, scrollBoxes []scrollableBox, title *template.TitleElement, inlined []inlinedStateful) {
	derivedDecls := collectAllDerivedDecls(sc, inlined)
	allSelfDecls := collectAllSelfDecls(sc, inlined)
	rootSelfDecls := rootOnlySelfDecls(sc)
	ext := writeTreeAndSync(buf, doc, stylesheet, instances, scrollBoxes, derivedDecls, rootSelfDecls)
	dynamic := isDynamic(ext, scrollBoxes)

	buf.WriteString("\tvar prevTree *layout.Box\n")
	buf.WriteString("\tvar prevW, prevH int\n")
	writeSelfPrevDecls(buf, allSelfDecls)
	if !dynamic {
		buf.WriteString("\tvar nodeBoxMap map[*layout.Input]*layout.Box\n")
	}
	buf.WriteString("\tdoRender := func() {\n")

	if dynamic {
		writeDynamicDoRender(buf, scrollBoxes, title)
	} else {
		writeStaticDoRender(buf, title)
	}
	writeSelfChangeDetection(buf, allSelfDecls)
	if ext.hasCursor {
		writeCursorPositioning(buf)
	}

	buf.WriteString("\t}\n\n")
}

// writeCursorPositioning emits cursor show/hide logic at the end of doRender.
func writeCursorPositioning(buf *bytes.Buffer) {
	buf.WriteString("\t\tif cursorBox := layout.FindCursor(tree); cursorBox != nil {\n")
	buf.WriteString("\t\t\trender.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\trender.HideCursor(os.Stdout)\n")
	buf.WriteString("\t\t}\n")
}

// writeStaticDoRender emits the doRender body for static documents (Pattern A).
// Includes three fast paths: no-op skip, direct-write, and full layout fallback.
func writeStaticDoRender(buf *bytes.Buffer, title *template.TitleElement) {
	buf.WriteString("\t\tchanged := sync()\n")
	buf.WriteString("\t\ttermW, termH := term.GetSize(int(os.Stdin.Fd()))\n")
	// No-op skip: nothing changed and no resize
	buf.WriteString("\t\tif prevTree != nil && len(changed) == 0 && termW == prevW && termH == prevH {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	// Direct-write fast path: same-length text changes without relayout
	writeDirectWriteFastPath(buf)
	writeFullLayoutAndDiff(buf, title)
}

// writeDirectWriteFastPath emits the direct-write block that skips Layout+Diff
// when all changed nodes have same-length content and no overlap is present.
func writeDirectWriteFastPath(buf *bytes.Buffer) {
	buf.WriteString("\t\tif prevTree != nil && len(changed) > 0 && termW == prevW && termH == prevH && !prevTree.HasOverlap && nodeBoxMap != nil {\n")
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

// writeDynamicDoRender emits the doRender body for dynamic documents (Pattern B).
// Sync is void; always runs full Layout+Diff.
func writeDynamicDoRender(buf *bytes.Buffer, scrollBoxes []scrollableBox, title *template.TitleElement) {
	buf.WriteString("\t\tsync()\n")
	buf.WriteString("\t\ttermW, termH := term.GetSize(int(os.Stdin.Fd()))\n")
	writeFullLayoutBody(buf, scrollBoxes, title)
}

// writeFullLayoutBody emits Layout, scroll wiring, DiffTrees, and redraw/apply logic.
func writeFullLayoutBody(buf *bytes.Buffer, scrollBoxes []scrollableBox, title *template.TitleElement) {
	buf.WriteString("\t\ttree := layout.Layout(root, termW, termH)\n")
	writeScrollTreeWiring(buf, scrollBoxes)
	buf.WriteString("\t\tchanges, scrollChanged := layout.DiffTrees(prevTree, tree)\n")
	buf.WriteString("\t\tif prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {\n")
	buf.WriteString("\t\t\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\t\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\t\t\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\tlayout.ApplyChanges(os.Stdout, changes)\n")
	buf.WriteString("\t\t}\n")
	writeTitleSet(buf, title)
	buf.WriteString("\t\tprevTree = tree\n")
	buf.WriteString("\t\tprevW = termW\n")
	buf.WriteString("\t\tprevH = termH\n")
}

// writeFullLayoutAndDiff emits the full layout path for static documents
// (no scroll wiring needed). Also builds the nodeBoxMap for direct-write.
func writeFullLayoutAndDiff(buf *bytes.Buffer, title *template.TitleElement) {
	buf.WriteString("\t\ttree := layout.Layout(root, termW, termH)\n")
	buf.WriteString("\t\tnodeBoxMap = layout.MapInputToBox(root, tree)\n")
	buf.WriteString("\t\tchanges, scrollChanged := layout.DiffTrees(prevTree, tree)\n")
	buf.WriteString("\t\tif prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {\n")
	buf.WriteString("\t\t\tbuf := render.NewBuffer(termW, termH)\n")
	buf.WriteString("\t\t\tlayout.RenderTree(buf, tree, nil)\n")
	buf.WriteString("\t\t\trender.ClearScreen(os.Stdout)\n")
	buf.WriteString("\t\t\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\tlayout.ApplyChanges(os.Stdout, changes)\n")
	buf.WriteString("\t\t}\n")
	writeTitleSet(buf, title)
	buf.WriteString("\t\tprevTree = tree\n")
	buf.WriteString("\t\tprevW = termW\n")
	buf.WriteString("\t\tprevH = termH\n")
}
