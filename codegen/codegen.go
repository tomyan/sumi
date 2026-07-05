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

	out, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("formatting generated code: %w\n%s", err, numberedLines(buf.Bytes()))
	}
	return out, nil
}

// numberedLines renders source with 1-based line numbers for error reports.
func numberedLines(src []byte) string {
	var b strings.Builder
	for i, line := range strings.Split(string(src), "\n") {
		fmt.Fprintf(&b, "%4d\t%s\n", i+1, line)
	}
	return b.String()
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
