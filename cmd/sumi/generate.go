package main

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/tomyan/sumi/codegen"
	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/section"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// parsedComponent holds a parsed but not yet generated .sumi component.
type parsedComponent struct {
	path       string
	doc        *template.Document
	script     *script.Script
	scriptSrc  string // raw script source (for go/ast parsing)
	stylesheet *style.Stylesheet
	name       string // e.g. "counter"
	exported   string // e.g. "Counter"
}

// generateFile compiles a single .sumi file to a _sumi.go file in the same directory.
func generateFile(path string) error {
	comp, err := parseSumiFile(path)
	if err != nil {
		return err
	}
	return generateComponent(comp)
}

// generateDir parses and generates all .sumi files in a directory.
func generateDir(dir string) error {
	components, err := parseAllSumiFiles(dir)
	if err != nil {
		return err
	}
	for _, comp := range components {
		if err := generateComponent(comp); err != nil {
			return err
		}
	}
	return nil
}

// parseSumiFile reads and parses a single .sumi file into a parsedComponent.
func parseSumiFile(path string) (*parsedComponent, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return parseSumiSource(path, string(src))
}

// parseSumiSource parses .sumi source text into a parsedComponent.
func parseSumiSource(path, src string) (*parsedComponent, error) {
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
	name := componentName(path)
	return &parsedComponent{
		path:       path,
		doc:        doc,
		script:     sc,
		scriptSrc:  sections.Script,
		stylesheet: ss,
		name:       name,
		exported:   exportedName(name),
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

// parseAllSumiFiles reads all .sumi files in a directory and parses them.
func parseAllSumiFiles(dir string) ([]*parsedComponent, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}
	var components []*parsedComponent
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sumi") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		comp, err := parseSumiFile(path)
		if err != nil {
			return nil, err
		}
		components = append(components, comp)
	}
	return components, nil
}

// generateComponent generates Go code for a single parsed component.
func generateComponent(comp *parsedComponent) error {
	// Use new signal-based codegen if script uses signal.New/signal.From/tui.Env.
	if comp.scriptSrc != "" && isSignalScript(comp.scriptSrc) {
		return generateSignalComponent(comp)
	}
	// Static path (no reactive state).
	opts := codegen.Options{PackageName: packageName(comp.path)}
	out, err := codegen.Generate(comp.doc, comp.script, comp.stylesheet, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", comp.path, err)
	}
	return writeOutput(comp.path, out)
}

// isSignalScript returns true if the script source uses signal.New or signal.From.
func isSignalScript(src string) bool {
	return strings.Contains(src, "signal.New") || strings.Contains(src, "signal.From") ||
		strings.Contains(src, "sumi.New") || strings.Contains(src, "sumi.From") ||
		strings.Contains(src, "tui.Env")
}

// generateSignalComponent generates code using the new component codegen.
func generateSignalComponent(comp *parsedComponent) error {
	out, err := codegen.GenerateComponent(comp.doc, comp.scriptSrc, comp.stylesheet, codegen.ComponentOptions{
		PackageName:   packageName(comp.path),
		ComponentName: comp.exported,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", comp.path, err)
	}
	return writeOutput(comp.path, out)
}

// buildCodegenOptions builds codegen.Options for a component.
// writeOutput writes generated Go source to the output file.
func writeOutput(path string, out []byte) error {
	outPath := outputPath(path)
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// componentName derives a component name from a file path.
// Strips the extension and removes hyphens: "my-widget.sumi" -> "mywidget".
func componentName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".sumi")
	return strings.ReplaceAll(name, "-", "")
}

// exportedName capitalizes the first letter of a component name.
func exportedName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// outputPath converts "foo.sumi" to "foo_sumi.go".
func outputPath(sumiPath string) string {
	dir := filepath.Dir(sumiPath)
	base := filepath.Base(sumiPath)
	name := strings.TrimSuffix(base, ".sumi")
	return filepath.Join(dir, name+"_sumi.go")
}

// packageName derives the Go package name from the directory containing the .sumi file.
// If a main.go exists in the same directory, returns "main" (it's an executable package).
// Otherwise uses the directory name, falling back to "main" if it's not a valid Go identifier.
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
