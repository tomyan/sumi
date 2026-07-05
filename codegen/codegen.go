package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// Generate produces Go source code for static (non-reactive) components.
// Used for .sumi files with no <script> block or no signal declarations.
// Emits func Run() and func CreateApp(w, h int) *tui.App.
func Generate(doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, packageName string) ([]byte, error) {
	annotateStyles(doc, stylesheet)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", packageName)
	hasExprs := docHasExprs(doc)
	writeImports(&buf, hasExprs, false, false)

	buf.WriteString("func Run() {\n")
	writeStaticBody(&buf, doc, stylesheet)
	buf.WriteString("}\n\n")

	buf.WriteString("func CreateApp(w, h int) *sumi.App {\n")
	writeStaticCreateAppBody(&buf, doc, stylesheet)
	buf.WriteString("}\n")

	return format.Source(buf.Bytes())
}

// usesSignals returns true if the script uses the new signal-based reactive model.
func usesSignals(sc *script.Script) bool {
	return sc != nil && (len(sc.SignalDecls) > 0 || len(sc.ComputedDecls) > 0)
}

// signalVarNames returns the set of variable names that are signals (need .Get() in templates).
func signalVarNames(sc *script.Script) map[string]bool {
	names := make(map[string]bool)
	if sc == nil {
		return names
	}
	for _, d := range sc.SignalDecls {
		names[d.Name] = true
	}
	for _, d := range sc.ComputedDecls {
		names[d.Name] = true
	}
	return names
}

// needsTimeImport returns true if any function body references the time package.
func needsTimeImport(sc *script.Script) bool {
	if sc != nil {
		for _, fd := range sc.FuncDecls {
			if strings.Contains(fd.Body, "time.") {
				return true
			}
		}
	}
	return false
}
