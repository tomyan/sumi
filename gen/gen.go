// Package gen turns .sumi source text into generated Go code. It holds the
// pure parse+codegen pipeline shared by the `sumi generate` CLI and the
// browser-side codegen.wasm tool, so both stay in lockstep.
package gen

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomyan/sumi/codegen"
	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/section"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// component holds a parsed but not yet generated .sumi component.
type component struct {
	path       string
	doc        *template.Document
	script     *script.Script
	scriptSrc  string // raw script source (for go/ast parsing)
	imports    string // raw <sumi:imports> content
	stylesheet *style.Stylesheet
	name       string // e.g. "counter"
	exported   string // e.g. "Counter"
}

// Generate parses .sumi source and returns the generated Go code. The path is
// used to derive the package name (main when a main.go sits beside it) and the
// component name; it need not exist on disk beyond that lookup.
func Generate(path, src string) ([]byte, error) {
	comp, err := parse(path, src)
	if err != nil {
		return nil, err
	}
	return generate(comp)
}

// OutputPath converts "foo.sumi" to "foo_sumi.go".
func OutputPath(sumiPath string) string {
	dir := filepath.Dir(sumiPath)
	base := filepath.Base(sumiPath)
	name := strings.TrimSuffix(base, ".sumi")
	return filepath.Join(dir, name+"_sumi.go")
}

// parse splits .sumi source into its sections and parses each one.
func parse(path, src string) (*component, error) {
	sections, err := section.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	doc, err := template.Parse(sections.Template)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	sc, err := parseOptionalScript(path, sections.Script)
	if err != nil {
		return nil, err
	}
	ss, err := parseOptionalStyle(path, sections.Style)
	if err != nil {
		return nil, err
	}
	name := template.ComponentName(path)
	return &component{
		path:       path,
		doc:        doc,
		script:     sc,
		scriptSrc:  sections.Script,
		imports:    sections.Imports,
		stylesheet: ss,
		name:       name,
		exported:   template.ExportedComponentName(name),
	}, nil
}

// parseOptionalScript parses a script block if non-empty.
func parseOptionalScript(path, src string) (*script.Script, error) {
	if src == "" {
		return nil, nil
	}
	sc, err := script.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return sc, nil
}

// parseOptionalStyle parses a style block if non-empty.
func parseOptionalStyle(path, src string) (*style.Stylesheet, error) {
	if src == "" {
		return nil, nil
	}
	ss, err := style.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return ss, nil
}

// generate selects the reactive or static codegen path for a parsed component.
func generate(comp *component) ([]byte, error) {
	if comp.scriptSrc != "" && isSignalScript(comp.scriptSrc) {
		out, err := codegen.GenerateComponent(comp.doc, comp.scriptSrc, comp.stylesheet, codegen.ComponentOptions{
			PackageName:   packageName(comp.path),
			ComponentName: comp.exported,
			UserImports:   comp.imports,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", comp.path, err)
		}
		return out, nil
	}
	out, err := codegen.Generate(comp.doc, comp.script, comp.stylesheet, packageName(comp.path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", comp.path, err)
	}
	return out, nil
}

// isSignalScript reports whether the script uses signals, var props, or other
// reactive features that require the component codegen path.
func isSignalScript(src string) bool {
	return strings.Contains(src, "sumi.New") || strings.Contains(src, "sumi.From") ||
		strings.Contains(src, "sumi.Effect") || strings.Contains(src, "sumi.Env") ||
		strings.Contains(src, "signal.") ||
		strings.Contains(src, "\nvar ") || strings.HasPrefix(strings.TrimSpace(src), "var ")
}

// packageName derives the Go package name from the directory containing the
// .sumi file. If a main.go exists in the same directory, returns "main" (it's
// an executable package). Otherwise uses the directory name, falling back to
// "main" if it's not a valid Go identifier.
func packageName(sumiPath string) string {
	absPath, err := filepath.Abs(sumiPath)
	if err != nil {
		return "main"
	}
	dir := filepath.Dir(absPath)
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
		return "main"
	}
	name := filepath.Base(dir)
	if !token.IsIdentifier(name) {
		return "main"
	}
	return name
}
