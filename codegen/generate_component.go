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
	writeComponentImports(&buf, info, doc)

	// Props struct (always generated, even if empty).
	writePropsStruct(&buf, opts.ComponentName, info.Props)

	// Constructor function.
	writeConstructor(&buf, opts.ComponentName, info, doc, scriptSrc, stylesheet)

	return format.Source(buf.Bytes())
}

// writeComponentImports emits import declarations for a component.
func writeComponentImports(buf *bytes.Buffer, info *script.ScriptInfo, doc *template.Document) {
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
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	if len(info.Signals) > 0 {
		buf.WriteString("\t\"github.com/tomyan/sumi/runtime/signal\"\n")
	}
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/tui\"\n")
	buf.WriteString(")\n\n")
}

// writePropsStruct emits the props struct type.
func writePropsStruct(buf *bytes.Buffer, name string, props []script.PropInfo) {
	fmt.Fprintf(buf, "type %sProps struct {\n", name)
	for _, p := range props {
		fieldName := exportedName(p.Name)
		fmt.Fprintf(buf, "\t%s %s\n", fieldName, p.TypeStr)
	}
	buf.WriteString("}\n\n")
}

// writeConstructor emits the NewFoo function.
func writeConstructor(buf *bytes.Buffer, name string, info *script.ScriptInfo, doc *template.Document, scriptSrc string, stylesheet *style.Stylesheet) {
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

	// Build layout tree with signal-aware expressions.
	ext := newExtractionCtx("")
	ext.signals = info.Signals
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
