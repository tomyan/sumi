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

// ComponentOptions configures component code generation.
type ComponentOptions struct {
	PackageName   string
	ComponentName string // e.g. "Counter" → generates NewCounter, CounterProps
	Components    map[string]ComponentChildInfo // child components available in templates
}

// ComponentChildInfo describes a child component available for use in templates.
type ComponentChildInfo struct {
	ImportPath string // e.g. "github.com/example/greeting"
	Package    string // e.g. "greeting"
}

// GenerateComponent produces Go source for a signal-based component.
// Returns a file containing the props struct and NewFoo constructor.
func GenerateComponent(doc *template.Document, scriptSrc string, stylesheet *style.Stylesheet, opts ComponentOptions) ([]byte, error) {
	// Parse script with go/ast.
	info, err := script.ParseGoAST(scriptSrc)
	if err != nil {
		return nil, fmt.Errorf("parse script: %w", err)
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", opts.PackageName)

	// Imports.
	writeComponentImports(&buf, info, doc, opts.Components)

	// Collect slot definitions from template for props.
	slots := collectSlots(doc)

	// Props struct (always generated, even if empty).
	writePropsStruct(&buf, opts.ComponentName, info.Props, slots)

	// Constructor function.
	writeConstructor(&buf, opts.ComponentName, info, doc, scriptSrc, stylesheet, opts)

	return format.Source(buf.Bytes())
}

// writeComponentImports emits import declarations for a component.
func writeComponentImports(buf *bytes.Buffer, info *script.ScriptInfo, doc *template.Document, components map[string]ComponentChildInfo) {
	buf.WriteString("import (\n")
	if docHasExprs(doc) {
		buf.WriteString("\t\"fmt\"\n")
	}
	buf.WriteString("\n")
	// Check if any function has input.Event parameter.
	needsInput := false
	for _, f := range info.Funcs {
		if strings.Contains(f.Params, "input.Event") {
			needsInput = true
			break
		}
	}
	if needsInput {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/input\"\n")
	}
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/layout\"\n")
	if hasStyles(doc) {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	}
	if len(info.Signals) > 0 {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/signal\"\n")
	}
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/tui\"\n")
	// Child component imports.
	if len(components) > 0 {
		buf.WriteString("\n")
		imported := make(map[string]bool)
		for _, ci := range components {
			if !imported[ci.ImportPath] {
				fmt.Fprintf(buf, "\t\"%s\"\n", ci.ImportPath)
				imported[ci.ImportPath] = true
			}
		}
	}
	buf.WriteString(")\n\n")
}

// writePropsStruct emits the props struct type.
func writePropsStruct(buf *bytes.Buffer, name string, props []script.PropInfo, slots []slotInfo) {
	fmt.Fprintf(buf, "type %sProps struct {\n", name)
	for _, p := range props {
		fieldName := exportedName(p.Name)
		fmt.Fprintf(buf, "\t%s %s\n", fieldName, p.TypeStr)
	}
	for _, s := range slots {
		fieldName := exportedName(s.name)
		fmt.Fprintf(buf, "\t%s []*layout.Input\n", fieldName)
	}
	buf.WriteString("}\n\n")
}

type slotInfo struct {
	name string
}

// collectSlots finds all <slot:name> elements in the template.
func collectSlots(doc *template.Document) []slotInfo {
	var slots []slotInfo
	walkSlots(doc.Children, func(s *template.SlotElement) {
		slots = append(slots, slotInfo{name: s.Name})
	})
	return slots
}

func walkSlots(children []template.Node, fn func(*template.SlotElement)) {
	for _, child := range children {
		switch n := child.(type) {
		case *template.SlotElement:
			fn(n)
		case *template.BoxElement:
			walkSlots(n.Children, fn)
		}
	}
}

// writeConstructor emits the NewFoo function.
func writeConstructor(buf *bytes.Buffer, name string, info *script.ScriptInfo, doc *template.Document, scriptSrc string, stylesheet *style.Stylesheet, opts ComponentOptions) {
	if len(info.Props) > 0 {
		fmt.Fprintf(buf, "func New%s(props %sProps) *tui.Component {\n", name, name)
		// Extract props into local variables.
		for _, p := range info.Props {
			field := exportedName(p.Name)
			if p.Default != "" {
				fmt.Fprintf(buf, "\t%s := props.%s\n", p.Name, field)
				fmt.Fprintf(buf, "\tif %s == %s {\n", p.Name, zeroValue(p.TypeStr))
				fmt.Fprintf(buf, "\t\t%s = %s\n", p.Name, p.Default)
				fmt.Fprintf(buf, "\t}\n")
			} else {
				fmt.Fprintf(buf, "\t%s := props.%s\n", p.Name, field)
			}
		}
		buf.WriteString("\n")
	} else {
		fmt.Fprintf(buf, "func New%s(props %sProps) *tui.Component {\n", name, name)
	}

	// Emit signal/variable declarations (non-var, non-func lines from script).
	writeScriptDeclarations(buf, info, scriptSrc)

	// Emit function closures.
	for _, f := range info.Funcs {
		writeComponentFunc(buf, f)
	}

	// Instantiate child components before building the tree.
	writeChildComponentInstances(buf, doc, opts.Components)

	// Build layout tree with signal-aware expressions.
	ext := newExtractionCtx("")
	ext.signals = info.Signals
	ext.componentChildren = opts.Components
	var treeBuf bytes.Buffer
	writeLayoutTree(&treeBuf, doc, stylesheet, false, nil, ext)
	buf.Write(ext.declBuf.Bytes())
	buf.Write(treeBuf.Bytes())

	// Signal effect for syncing expression nodes.
	if len(ext.nodes) > 0 {
		buf.WriteString("\n\tsignal.Effect(func() {\n")
		for _, n := range ext.nodes {
			fmt.Fprintf(buf, "\t\t%s.Content = %s\n", n.varName, n.syncExpr)
		}
		buf.WriteString("\t})\n")
	}

	// Build and return Component.
	writeComponentReturn(buf, info)

	buf.WriteString("}\n")
}

// writeScriptDeclarations emits the non-func, non-var declarations from the script.
func writeScriptDeclarations(buf *bytes.Buffer, info *script.ScriptInfo, src string) {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "var ") {
			continue // props, already handled
		}
		if strings.HasPrefix(trimmed, "func ") {
			break // functions come later via writeComponentFunc
		}
		fmt.Fprintf(buf, "\t%s\n", trimmed)
	}
	buf.WriteString("\n")
}

// writeChildComponentInstances generates constructor calls for child components used in templates.
func writeChildComponentInstances(buf *bytes.Buffer, doc *template.Document, components map[string]ComponentChildInfo) {
	if len(components) == 0 {
		return
	}
	idx := 0
	walkComponentElements(doc.Children, func(comp *template.ComponentElement) {
		info, ok := components[comp.Name]
		if !ok {
			return
		}
		varName := fmt.Sprintf("%s%d", strings.ToLower(comp.Name[:1])+comp.Name[1:], idx)
		fmt.Fprintf(buf, "\t%s := %s.New%s(%s.%sProps{\n", varName, info.Package, comp.Name, info.Package, comp.Name)
		for k, v := range comp.Attributes {
			if strings.HasPrefix(k, "bind:") {
				// bind:value={name} → Value: name (pass signal reference)
				propName := strings.TrimPrefix(k, "bind:")
				fieldName := exportedName(propName)
				expr := v
				if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
					expr = expr[1 : len(expr)-1]
				}
				fmt.Fprintf(buf, "\t\t%s: %s,\n", fieldName, expr)
			} else if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
				// Expression prop: value={expr}
				fieldName := exportedName(k)
				expr := v[1 : len(v)-1]
				fmt.Fprintf(buf, "\t\t%s: %s,\n", fieldName, expr)
			} else {
				// Literal prop: value="text"
				fieldName := exportedName(k)
				fmt.Fprintf(buf, "\t\t%s: %q,\n", fieldName, v)
			}
		}
		fmt.Fprintf(buf, "\t})\n")
		idx++
	})
	buf.WriteString("\n")
}

// walkComponentElements finds all ComponentElement nodes in a template tree.
func walkComponentElements(children []template.Node, fn func(*template.ComponentElement)) {
	for _, child := range children {
		switch n := child.(type) {
		case *template.ComponentElement:
			fn(n)
		case *template.BoxElement:
			walkComponentElements(n.Children, fn)
		}
	}
}

// rewriteAppCalls replaces app.Quit() with tui.Quit() in function bodies.
func rewriteAppCalls(body string) string {
	return strings.ReplaceAll(body, "app.Quit()", "tui.Quit()")
}

// writeComponentFunc emits a function closure.
func writeComponentFunc(buf *bytes.Buffer, f script.FuncInfo) {
	if f.Params == "" {
		fmt.Fprintf(buf, "\t%s := func() {\n", f.Name)
	} else {
		fmt.Fprintf(buf, "\t%s := func(%s) {\n", f.Name, f.Params)
	}
	body := rewriteAppCalls(f.Body)
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fmt.Fprintf(buf, "\t\t%s\n", trimmed)
	}
	buf.WriteString("\t}\n\n")
}

// writeComponentReturn emits the return statement.
func writeComponentReturn(buf *bytes.Buffer, info *script.ScriptInfo) {
	// Find the handleKey function if present.
	var handler string
	for _, f := range info.Funcs {
		if f.Name == "handleKey" {
			handler = f.Name
			break
		}
	}

	buf.WriteString("\n\treturn &tui.Component{\n")
	buf.WriteString("\t\tTree: root,\n")
	if handler != "" {
		fmt.Fprintf(buf, "\t\tOnEvent: %s,\n", handler)
	}
	buf.WriteString("\t}\n")
}

// hasStyles checks if any node in the document has style attributes.
func hasStyles(doc *template.Document) bool {
	for _, child := range doc.Children {
		if nodeHasStyle(child) {
			return true
		}
	}
	return false
}

func nodeHasStyle(node template.Node) bool {
	switch n := node.(type) {
	case *template.TextElement:
		return len(n.Attributes) > 0
	case *template.BoxElement:
		if _, ok := n.Attributes["class"]; ok {
			return true
		}
		for _, child := range n.Children {
			if nodeHasStyle(child) {
				return true
			}
		}
	}
	return false
}

// exportedName converts a Go identifier to exported (PascalCase).
func exportedName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// zeroValue returns the zero value literal for a type string.
func zeroValue(typeStr string) string {
	switch typeStr {
	case "string":
		return `""`
	case "int", "int64", "float64":
		return "0"
	case "bool":
		return "false"
	default:
		return "nil"
	}
}
