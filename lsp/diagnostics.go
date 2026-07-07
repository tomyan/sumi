package lsp

import (
	"errors"
	"fmt"

	"github.com/tomyan/sumi/codegen"
	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/section"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// Diagnostics compiles a .sumi document and returns the problems found. It
// never panics: any panic in the compile pipeline is recovered and reported
// as a whole-first-line diagnostic.
func Diagnostics(text string) (diags []Diagnostic) {
	defer func() {
		if r := recover(); r != nil {
			diags = []Diagnostic{firstLineDiagnostic(text, fmt.Sprintf("internal error: %v", r))}
		}
	}()
	return diagnose(text)
}

// diagnose runs the same pipeline as cmd/sumi/generate.go, stopping at the
// first stage that reports an error.
func diagnose(text string) []Diagnostic {
	sections, err := section.Parse(text)
	if err != nil {
		return []Diagnostic{firstLineDiagnostic(text, err.Error())}
	}
	doc, err := template.Parse(sections.Template)
	if err != nil {
		return []Diagnostic{templateDiagnostic(text, sections, err)}
	}
	sc, d, ok := parseScript(text, sections)
	if !ok {
		return []Diagnostic{d}
	}
	ss, d, ok := parseStyle(text, sections)
	if !ok {
		return []Diagnostic{d}
	}
	if _, err := codegen.Generate(doc, sc, ss, "app"); err != nil {
		return []Diagnostic{firstLineDiagnostic(text, err.Error())}
	}
	return []Diagnostic{}
}

// parseScript parses the script section if present. The bool is false when a
// diagnostic should be reported instead.
func parseScript(text string, s section.Sections) (*script.Script, Diagnostic, bool) {
	if s.Script == "" {
		return nil, Diagnostic{}, true
	}
	sc, err := script.Parse(s.Script)
	if err != nil {
		return nil, firstLineDiagnostic(text, err.Error()), false
	}
	return sc, Diagnostic{}, true
}

// parseStyle parses the style section if present. The bool is false when a
// diagnostic should be reported instead.
func parseStyle(text string, s section.Sections) (*style.Stylesheet, Diagnostic, bool) {
	if s.Style == "" {
		return nil, Diagnostic{}, true
	}
	ss, err := style.Parse(s.Style)
	if err != nil {
		return nil, firstLineDiagnostic(text, err.Error()), false
	}
	return ss, Diagnostic{}, true
}

// templateDiagnostic maps a template parse error to a precise range when the
// error carries an offset, falling back to a whole-first-line diagnostic.
func templateDiagnostic(text string, s section.Sections, err error) Diagnostic {
	var perr *template.Error
	if errors.As(err, &perr) && s.TemplateStart >= 0 {
		pos := offsetToPosition(text, s.TemplateStart+perr.Offset)
		return Diagnostic{
			Range:    Range{Start: pos, End: pos},
			Severity: SeverityError,
			Source:   "sumi",
			Message:  perr.Msg,
		}
	}
	return firstLineDiagnostic(text, err.Error())
}

// firstLineDiagnostic reports a problem spanning the document's first line.
func firstLineDiagnostic(text, msg string) Diagnostic {
	end := utf16Len(firstLine(text))
	return Diagnostic{
		Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: end}},
		Severity: SeverityError,
		Source:   "sumi",
		Message:  msg,
	}
}
