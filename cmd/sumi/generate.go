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
	stylesheet *style.Stylesheet
	name       string // e.g. "counter"
	exported   string // e.g. "Counter"
}

// generateFile compiles a single .sumi file to a _sumi.go file in the same directory.
// Works standalone without a component registry (backward compat).
func generateFile(path string) error {
	comp, err := parseSumiFile(path)
	if err != nil {
		return err
	}
	return generateComponent(comp, nil)
}

// generateDir uses two-pass compilation: parse all, build registry, generate all.
// Embedded built-in components are automatically merged into the registry.
func generateDir(dir string) error {
	components, err := parseAllSumiFiles(dir)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return nil
	}
	registry := buildComponentRegistry(components)
	if err := mergeEmbeddedComponents(registry); err != nil {
		return err
	}
	if err := validateAllComponentRefs(components, registry); err != nil {
		return err
	}
	return generateAllComponents(components, registry)
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

// buildComponentRegistry creates a registry of child components (those with props).
// Registry keys are normalized via TagRegistryKey.
func buildComponentRegistry(components []*parsedComponent) map[string]*codegen.ComponentInfo {
	registry := make(map[string]*codegen.ComponentInfo)
	for _, comp := range components {
		if !isChildComponent(comp) {
			continue
		}
		key := template.TagRegistryKey(comp.name)
		registry[key] = buildComponentInfo(comp)
	}
	return registry
}

// mergeEmbeddedComponents parses all embedded .sumi files and adds them to the registry.
// User-defined components with the same key take precedence (already in the registry).
func mergeEmbeddedComponents(registry map[string]*codegen.ComponentInfo) error {
	for _, tagName := range listEmbeddedComponents() {
		key := template.TagRegistryKey(tagName)
		if _, exists := registry[key]; exists {
			continue // user-defined component takes precedence
		}
		src, err := readEmbeddedComponent(tagName)
		if err != nil {
			return err
		}
		comp, err := parseSumiSource("embedded:"+tagName, src)
		if err != nil {
			return err
		}
		if !isChildComponent(comp) {
			continue
		}
		registry[key] = buildComponentInfo(comp)
	}
	return nil
}

// isChildComponent returns true if the component has prop declarations.
func isChildComponent(comp *parsedComponent) bool {
	return comp.script != nil && len(comp.script.PropDecls) > 0
}

// buildComponentInfo creates a ComponentInfo from a parsed component.
// Includes the full parsed AST for template inlining.
func buildComponentInfo(comp *parsedComponent) *codegen.ComponentInfo {
	props := make([]string, len(comp.script.PropDecls))
	for i, p := range comp.script.PropDecls {
		props[i] = p.Name
	}
	hasState := comp.script != nil && (len(comp.script.StateDecls) > 0 || len(comp.script.SelfDecls) > 0 || len(comp.script.DerivedDecls) > 0 || len(comp.script.EnvDecls) > 0)
	return &codegen.ComponentInfo{
		Name:         comp.name,
		ExportedName: comp.exported,
		Props:        props,
		HasState:     hasState,
		Doc:          comp.doc,
		Script:       comp.script,
		Stylesheet:   comp.stylesheet,
	}
}

// validateAllComponentRefs checks all component references resolve to the registry.
func validateAllComponentRefs(components []*parsedComponent, registry map[string]*codegen.ComponentInfo) error {
	for _, comp := range components {
		if err := validateComponentRefs(comp, registry); err != nil {
			return err
		}
	}
	return nil
}

// validateComponentRefs walks the document looking for unresolved ComponentElement nodes.
func validateComponentRefs(comp *parsedComponent, registry map[string]*codegen.ComponentInfo) error {
	return walkValidateNodes(comp.path, comp.doc.Children, registry)
}

// walkValidateNodes recursively validates component references in a node list.
func walkValidateNodes(path string, nodes []template.Node, registry map[string]*codegen.ComponentInfo) error {
	for _, node := range nodes {
		if err := walkValidateNode(path, node, registry); err != nil {
			return err
		}
	}
	return nil
}

// walkValidateNode validates a single node for component references.
func walkValidateNode(path string, node template.Node, registry map[string]*codegen.ComponentInfo) error {
	switch n := node.(type) {
	case *template.ComponentElement:
		key := template.TagRegistryKey(n.Name)
		if _, ok := registry[key]; !ok {
			return fmt.Errorf("%s: unknown component <%s />", path, n.Name)
		}
	case *template.BoxElement:
		return walkValidateNodes(path, n.Children, registry)
	case *template.IfNode:
		if err := walkValidateNodes(path, n.Then, registry); err != nil {
			return err
		}
		return walkValidateNodes(path, n.Else, registry)
	case *template.ForNode:
		return walkValidateNodes(path, n.Children, registry)
	}
	return nil
}

// generateAllComponents generates Go code for all parsed components.
// Child components are inlined into root templates and don't need separate output.
func generateAllComponents(components []*parsedComponent, registry map[string]*codegen.ComponentInfo) error {
	for _, comp := range components {
		if isChildComponent(comp) {
			continue // inlined at compile time, no separate output
		}
		if err := generateComponent(comp, registry); err != nil {
			return err
		}
	}
	return nil
}

// generateComponent generates Go code for a single parsed component.
func generateComponent(comp *parsedComponent, registry map[string]*codegen.ComponentInfo) error {
	opts := buildCodegenOptions(comp, registry)
	out, err := codegen.Generate(comp.doc, comp.script, comp.stylesheet, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", comp.path, err)
	}
	return writeOutput(comp.path, out)
}

// buildCodegenOptions builds codegen.Options for a component.
func buildCodegenOptions(comp *parsedComponent, registry map[string]*codegen.ComponentInfo) codegen.Options {
	return codegen.Options{
		PackageName: packageName(comp.path),
		Components:  registry,
	}
}

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
