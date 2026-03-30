package codegen

import (
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/template"
)

// contentExprSignals generates the Go expression for a TextElement's content,
// auto-unwrapping signal variables with .Get() in expressions.
func contentExprSignals(parts []template.Part, signals map[string]bool) string {
	if len(parts) == 0 {
		return `""`
	}
	if allStringParts(parts) {
		return fmt.Sprintf("%q", concatStringParts(parts))
	}
	return buildSprintfExprSignals(parts, signals)
}

// contentExpr generates the Go expression for a TextElement's content.
// Pure string parts produce a quoted string literal.
// Mixed parts with expressions produce a sumi.Sprintf call.
func contentExpr(parts []template.Part) string {
	if len(parts) == 0 {
		return `""`
	}
	if allStringParts(parts) {
		return fmt.Sprintf("%q", concatStringParts(parts))
	}
	return buildSprintfExpr(parts)
}

// allStringParts returns true if every part is a StringPart.
func allStringParts(parts []template.Part) bool {
	for _, p := range parts {
		if _, ok := p.(*template.ExprPart); ok {
			return false
		}
	}
	return true
}

// concatStringParts concatenates all StringPart values.
func concatStringParts(parts []template.Part) string {
	var sb strings.Builder
	for _, p := range parts {
		sb.WriteString(p.(*template.StringPart).Value)
	}
	return sb.String()
}

// buildSprintfExprSignals builds a sumi.Sprintf call, auto-unwrapping signal variables.
func buildSprintfExprSignals(parts []template.Part, signals map[string]bool) string {
	var fmtStr strings.Builder
	var args []string
	for _, p := range parts {
		switch pt := p.(type) {
		case *template.StringPart:
			fmtStr.WriteString(strings.ReplaceAll(pt.Value, "%", "%%"))
		case *template.ExprPart:
			fmtStr.WriteString("%v")
			args = append(args, unwrapSignals(pt.Expr, signals))
		}
	}
	return fmt.Sprintf("sumi.Sprintf(%q, %s)", fmtStr.String(), strings.Join(args, ", "))
}

// unwrapSignals replaces signal variable references in an expression with .Get() calls.
func unwrapSignals(expr string, signals map[string]bool) string {
	result := expr
	for name := range signals {
		result = replaceIdentifier(result, name, name+".Get()")
	}
	return result
}

// buildSprintfExpr builds a sumi.Sprintf call from mixed string/expr parts.
func buildSprintfExpr(parts []template.Part) string {
	var fmtStr strings.Builder
	var args []string
	for _, p := range parts {
		switch pt := p.(type) {
		case *template.StringPart:
			fmtStr.WriteString(strings.ReplaceAll(pt.Value, "%", "%%"))
		case *template.ExprPart:
			fmtStr.WriteString("%v")
			args = append(args, pt.Expr)
		}
	}
	return fmt.Sprintf("sumi.Sprintf(%q, %s)", fmtStr.String(), strings.Join(args, ", "))
}

// docHasExprs checks if any TextElement in the document has ExprParts.
func docHasExprs(doc *template.Document) bool {
	for _, child := range doc.Children {
		if nodeHasExprs(child) {
			return true
		}
	}
	return false
}

// docHasForKey checks if any ForNode in the document has a Key expression.
func docHasForKey(doc *template.Document) bool {
	for _, child := range doc.Children {
		if nodeHasForKey(child) {
			return true
		}
	}
	return false
}

// nodeHasForKey checks if a single node (or its children) contains a keyed ForNode.
func nodeHasForKey(node template.Node) bool {
	switch n := node.(type) {
	case *template.BoxElement:
		for _, child := range n.Children {
			if nodeHasForKey(child) {
				return true
			}
		}
	case *template.IfNode:
		for _, child := range n.Then {
			if nodeHasForKey(child) {
				return true
			}
		}
		for _, child := range n.Else {
			if nodeHasForKey(child) {
				return true
			}
		}
	case *template.ForNode:
		if n.Key != "" {
			return true
		}
		for _, child := range n.Children {
			if nodeHasForKey(child) {
				return true
			}
		}
	}
	return false
}

// nodeHasExprs checks if a single node (or its children) contains ExprParts.
func nodeHasExprs(node template.Node) bool {
	switch n := node.(type) {
	case *template.TextElement:
		for _, p := range n.Parts {
			if _, ok := p.(*template.ExprPart); ok {
				return true
			}
		}
	case *template.BoxElement:
		for _, child := range n.Children {
			if nodeHasExprs(child) {
				return true
			}
		}
	case *template.IfNode:
		for _, child := range n.Then {
			if nodeHasExprs(child) {
				return true
			}
		}
		for _, child := range n.Else {
			if nodeHasExprs(child) {
				return true
			}
		}
	case *template.ForNode:
		for _, child := range n.Children {
			if nodeHasExprs(child) {
				return true
			}
		}
	}
	return false
}
