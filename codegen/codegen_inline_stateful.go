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
// Bound variables (via bind:) are skipped since the parent owns them.
func writeInlinedStateDecls(buf *bytes.Buffer, inlined []inlinedStateful) {
	for _, is := range inlined {
		sc := is.Instance.Info.Script
		if sc == nil {
			continue
		}
		bindings := extractBindings(is.Instance.Attrs)
		for _, sd := range sc.StateDecls {
			if _, bound := bindings[sd.Name]; bound {
				continue // parent owns this variable
			}
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
		propMap := buildPropMap(is.Instance)
		stateNameMap := buildStateNameMap(is.Instance)
		for _, fd := range sc.FuncDecls {
			if fd.Params != "" {
				fmt.Fprintf(buf, "\t%s%s := func(%s) {\n", is.Prefix, fd.Name, fd.Params)
			} else {
				fmt.Fprintf(buf, "\t%s%s := func() {\n", is.Prefix, fd.Name)
			}
			writeNamespacedFuncBody(buf, fd, stateNameMap, propMap)
			buf.WriteString("\t}\n")
		}
	}
}

// writeNamespacedFuncBody writes a function body with state vars namespaced and callback props resolved.
// stateNameMap maps child variable names to their resolved names (namespaced or parent-bound).
func writeNamespacedFuncBody(buf *bytes.Buffer, funcDecl script.FuncDecl, stateNameMap map[string]string, propMap map[string]string) {
	stateLines := buildStateLinesSet(funcDecl.StateAssignments)
	for _, line := range strings.Split(funcDecl.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		namespaced := namespaceAssignmentMapped(trimmed, funcDecl.StateAssignments, stateNameMap)
		namespaced = resolveCallbackProps(namespaced, propMap)
		fmt.Fprintf(buf, "\t\t%s\n", namespaced)
		if stateLines[trimmed] {
			buf.WriteString("\t\tapp.Dirty = true\n")
		}
	}
}

// resolveCallbackProps replaces prop name references in a line with their resolved values.
// Handles function calls: propName(...) → resolvedValue(...)
// Expression prop values ({expr}) have curlies stripped before substitution.
func resolveCallbackProps(line string, propMap map[string]string) string {
	for propName, propValue := range propMap {
		resolved := propValue
		if isExprValue(resolved) {
			resolved = extractExprValue(resolved)
		}
		line = strings.ReplaceAll(line, propName+"(", resolved+"(")
	}
	return line
}

// namespaceAssignmentMapped rewrites a line, replacing state variable references using the stateNameMap.
// The map translates child var names to their resolved names (namespaced or parent-bound).
func namespaceAssignmentMapped(line string, assignments []script.StateAssignment, stateNameMap map[string]string) string {
	for _, sa := range assignments {
		if line == sa.Line {
			return renameVarInLine(line, sa.VarName, stateNameMap)
		}
	}
	return line
}

// renameVarInLine replaces all occurrences of a variable name with its mapped name from stateNameMap.
func renameVarInLine(line, varName string, stateNameMap map[string]string) string {
	newName, ok := stateNameMap[varName]
	if !ok {
		return line
	}
	// Replace "varName =" with "newName ="
	result := strings.Replace(line, varName+" =", newName+" =", 1)
	// Replace references on the right side
	eqIdx := strings.Index(result, "=")
	if eqIdx < 0 {
		return result
	}
	left := result[:eqIdx+1]
	right := result[eqIdx+1:]
	right = renameVarRef(right, varName, newName)
	return left + right
}

// renameVarRef replaces standalone variable references with a new name.
func renameVarRef(s, oldName, newName string) string {
	result := strings.ReplaceAll(s, " "+oldName+" ", " "+newName+" ")
	result = strings.ReplaceAll(result, " "+oldName+"\n", " "+newName+"\n")
	if strings.HasSuffix(result, " "+oldName) {
		result = result[:len(result)-len(oldName)] + newName
	}
	return result
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
