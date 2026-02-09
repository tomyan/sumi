package codegen

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// Options configures code generation.
type Options struct {
	PackageName   string
	ComponentName string // empty for root components, set for child components
}

// Generate produces Go source code from a template AST, optional script, and optional stylesheet.
// When sc is nil, generates static code (render once, wait for Enter).
// When sc has state, generates reactive code with an event loop.
// When stylesheet is non-nil, styles are resolved at codegen time and emitted as render.Style literals.
func Generate(doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, opts Options) ([]byte, error) {
	if isComponentMode(sc) {
		return generateComponent(doc, sc, stylesheet, opts)
	}
	return generateRunFunc(doc, sc, stylesheet, opts)
}

// isComponentMode returns true when the script has prop declarations.
func isComponentMode(sc *script.Script) bool {
	return sc != nil && len(sc.PropDecls) > 0
}

// generateRunFunc generates the existing func Run() code path.
func generateRunFunc(doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, opts Options) ([]byte, error) {
	reactive := sc != nil && len(sc.StateDecls) > 0
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", opts.PackageName)
	writeImports(&buf, reactive, docHasExprs(doc))
	buf.WriteString("func Run() {\n")
	if reactive {
		writeReactiveBody(&buf, doc, sc, stylesheet)
	} else {
		writeStaticBody(&buf, doc, stylesheet)
	}
	buf.WriteString("}\n\n")
	writeRenderTreeFunc(&buf)
	return format.Source(buf.Bytes())
}
