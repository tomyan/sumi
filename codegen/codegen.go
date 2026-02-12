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

// ComponentInfo describes a child component available in templates.
type ComponentInfo struct {
	Name         string   // component name as used in template, e.g. "counter"
	ExportedName string   // Go exported name, e.g. "Counter"
	Props        []string // prop names in order, e.g. ["label"]
	HasState     bool     // whether component has state (affects if HandleKey exists)
	Doc          *template.Document  // full parsed template (for inlining)
	Script       *script.Script      // full parsed script (for inlining)
	Stylesheet   *style.Stylesheet   // full parsed stylesheet (for inlining)
}

// Options configures code generation.
type Options struct {
	PackageName string
	Components  map[string]*ComponentInfo // child components available in templates
}

// Generate produces Go source code from a template AST, optional script, and optional stylesheet.
// When sc is nil, generates static code (render once, wait for Enter).
// When sc has state, generates reactive code with an event loop.
// When stylesheet is non-nil, styles are resolved at codegen time and emitted as render.Style literals.
func Generate(doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, opts Options) ([]byte, error) {
	return generateRunFunc(doc, sc, stylesheet, opts)
}

// generateRunFunc generates the existing func Run() code path.
func generateRunFunc(doc *template.Document, sc *script.Script, stylesheet *style.Stylesheet, opts Options) ([]byte, error) {
	instances := collectComponentInstances(doc, opts.Components)
	reactive := hasReactiveContent(sc, instances)
	inlined := collectInlinedStateful(instances)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", opts.PackageName)
	hasTitle := findTitleElement(doc) != nil
	hasScroll := len(findAllScrollableBoxes(doc, stylesheet)) > 0
	hasTime := needsTimeImport(sc, inlined)
	writeImports(&buf, docHasExprs(doc) || hasTitle || hasScroll || docHasForKey(doc) || instancesHaveExprs(instances), reactive, hasTime)
	buf.WriteString("func Run() {\n")
	if reactive {
		writeReactiveBody(&buf, doc, sc, stylesheet, instances)
	} else {
		writeStaticBody(&buf, doc, stylesheet)
	}
	buf.WriteString("}\n")
	return format.Source(buf.Bytes())
}

// hasReactiveContent returns true when the document needs the reactive code path.
func hasReactiveContent(sc *script.Script, instances []componentInstance) bool {
	if sc != nil && (len(sc.StateDecls) > 0 || len(sc.EnvDecls) > 0 || len(sc.DerivedDecls) > 0 || len(sc.SelfDecls) > 0) {
		return true
	}
	return len(instances) > 0
}

// needsTimeImport returns true if any function body references the time package.
func needsTimeImport(sc *script.Script, inlined []inlinedStateful) bool {
	if sc != nil {
		for _, fd := range sc.FuncDecls {
			if strings.Contains(fd.Body, "time.") {
				return true
			}
		}
	}
	for _, is := range inlined {
		if is.Instance.Info.Script == nil {
			continue
		}
		for _, fd := range is.Instance.Info.Script.FuncDecls {
			if strings.Contains(fd.Body, "time.") {
				return true
			}
		}
	}
	return false
}

// instancesHaveExprs returns true if any inlined component has expression parts.
func instancesHaveExprs(instances []componentInstance) bool {
	for _, inst := range instances {
		if inst.Info.Doc != nil && docHasExprs(inst.Info.Doc) {
			return true
		}
	}
	return false
}
