package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// inlinedStateful tracks a stateful component instance that is inlined into the root.
type inlinedStateful struct {
	Prefix   string         // namespace prefix, e.g. "counter0_"
	Instance *componentInstance
}

// collectInlinedStateful returns inlined instances that have state (need closures + state decls).
func collectInlinedStateful(instances []componentInstance) []inlinedStateful {
	var result []inlinedStateful
	for i := range instances {
		inst := &instances[i]
		if inst.Info.Doc == nil || !inst.Info.HasState {
			continue
		}
		result = append(result, inlinedStateful{
			Prefix:   inst.VarName + "_",
			Instance: inst,
		})
	}
	return result
}

// writeInlinedStateDecls writes namespaced state and derived variable declarations for inlined components.
func writeInlinedStateDecls(buf *bytes.Buffer, inlined []inlinedStateful) {
	for _, is := range inlined {
		sc := is.Instance.Info.Script
		if sc == nil {
			continue
		}
		for _, sd := range sc.StateDecls {
			fmt.Fprintf(buf, "\t%s%s := %s\n", is.Prefix, sd.Name, sd.InitExpr)
		}
		for _, dd := range sc.DerivedDecls {
			expr := namespaceDerivedExpr(dd.Expr, sc, is.Prefix)
			fmt.Fprintf(buf, "\t%s%s := %s\n", is.Prefix, dd.Name, expr)
		}
	}
}

// namespaceDerivedExpr replaces state variable references in a derived expression with namespaced versions.
// Prepends a space so namespaceVarRef can match vars at the start of the expression.
func namespaceDerivedExpr(expr string, sc *script.Script, prefix string) string {
	result := " " + expr
	for _, sd := range sc.StateDecls {
		result = namespaceVarRef(result, sd.Name, prefix)
	}
	for _, dd := range sc.DerivedDecls {
		result = namespaceVarRef(result, dd.Name, prefix)
	}
	return strings.TrimPrefix(result, " ")
}

// writeInlinedFuncClosures writes namespaced function closures for inlined components.
func writeInlinedFuncClosures(buf *bytes.Buffer, inlined []inlinedStateful) {
	for _, is := range inlined {
		sc := is.Instance.Info.Script
		if sc == nil {
			continue
		}
		for _, fd := range sc.FuncDecls {
			if fd.Params != "" {
				fmt.Fprintf(buf, "\t%s%s := func(%s) {\n", is.Prefix, fd.Name, fd.Params)
			} else {
				fmt.Fprintf(buf, "\t%s%s := func() {\n", is.Prefix, fd.Name)
			}
			writeNamespacedFuncBody(buf, fd, is.Prefix)
			buf.WriteString("\t}\n")
		}
	}
}

// writeNamespacedFuncBody writes a function body with state vars namespaced.
func writeNamespacedFuncBody(buf *bytes.Buffer, funcDecl script.FuncDecl, prefix string) {
	stateLines := buildStateLinesSet(funcDecl.StateAssignments)
	for _, line := range strings.Split(funcDecl.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		namespaced := namespaceAssignment(trimmed, funcDecl.StateAssignments, prefix)
		fmt.Fprintf(buf, "\t\t%s\n", namespaced)
		if stateLines[trimmed] {
			buf.WriteString("\t\tapp.Dirty = true\n")
		}
	}
}

// namespaceAssignment rewrites a line, replacing state variable references with namespaced versions.
func namespaceAssignment(line string, assignments []script.StateAssignment, prefix string) string {
	for _, sa := range assignments {
		if line == sa.Line {
			return namespaceVarInLine(line, sa.VarName, prefix)
		}
	}
	return line
}

// namespaceVarInLine replaces all occurrences of a variable name with its namespaced version.
func namespaceVarInLine(line, varName, prefix string) string {
	// Replace "varName =" with "prefix_varName ="
	result := strings.Replace(line, varName+" =", prefix+varName+" =", 1)
	// Replace references on the right side
	eqIdx := strings.Index(result, "=")
	if eqIdx < 0 {
		return result
	}
	left := result[:eqIdx+1]
	right := result[eqIdx+1:]
	right = namespaceVarRef(right, varName, prefix)
	return left + right
}

// namespaceVarRef replaces standalone variable references with namespaced versions.
func namespaceVarRef(s, varName, prefix string) string {
	result := strings.ReplaceAll(s, " "+varName+" ", " "+prefix+varName+" ")
	result = strings.ReplaceAll(result, " "+varName+"\n", " "+prefix+varName+"\n")
	if strings.HasSuffix(result, " "+varName) {
		result = result[:len(result)-len(varName)] + prefix + varName
	}
	return result
}

// findChildOnkeyHandler finds the onkey handler name from a child component's document.
func findChildOnkeyHandler(doc *template.Document) string {
	return findRootOnkey(doc)
}

// writeSuppressInlinedFuncs writes _ = funcName for inlined functions not used as onkey handlers.
func writeSuppressInlinedFuncs(buf *bytes.Buffer, inlined []inlinedStateful) {
	for _, is := range inlined {
		sc := is.Instance.Info.Script
		if sc == nil {
			continue
		}
		for _, fd := range sc.FuncDecls {
			if !childDocHasOnkey(is.Instance.Info.Doc, fd.Name) {
				fmt.Fprintf(buf, "\t_ = %s%s\n", is.Prefix, fd.Name)
			}
		}
	}
}

// childDocHasOnkey checks if a child component's document references a function as an onkey handler.
func childDocHasOnkey(doc *template.Document, funcName string) bool {
	return docHasOnkey(doc, funcName)
}

// buildStateVarMap builds a map of state variable names for a component's script.
// This is used for namespacing expressions in inlined templates.
func buildStateVarMap(sc *script.Script) map[string]bool {
	if sc == nil {
		return nil
	}
	vars := make(map[string]bool)
	for _, sd := range sc.StateDecls {
		vars[sd.Name] = true
	}
	return vars
}
