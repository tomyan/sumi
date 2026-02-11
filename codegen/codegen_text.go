package codegen

import (
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/template"
)

// contentExpr generates the Go expression for a TextElement's content.
// Pure string parts produce a quoted string literal.
// Mixed parts with expressions produce a fmt.Sprintf call.
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

// buildSprintfExpr builds a fmt.Sprintf call from mixed string/expr parts.
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
	return fmt.Sprintf("fmt.Sprintf(%q, %s)", fmtStr.String(), strings.Join(args, ", "))
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
