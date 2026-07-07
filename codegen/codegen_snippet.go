package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// collectLocalSnippets walks the template collecting {snippet} declarations
// that belong to THIS component. Snippets inside a child component tag are
// consumer snippets (props of that child), so the walk does not descend into
// ComponentElement bodies. Local snippets are hoisted to component scope.
func collectLocalSnippets(children []template.Node) []*template.SnippetNode {
	var snippets []*template.SnippetNode
	var walk func(nodes []template.Node)
	walk = func(nodes []template.Node) {
		for _, n := range nodes {
			switch node := n.(type) {
			case *template.SnippetNode:
				snippets = append(snippets, node)
				walk(node.Children)
			case *template.BoxElement:
				walk(node.Children)
			case *template.IfNode:
				walk(node.Then)
				walk(node.Else)
			case *template.ForNode:
				walk(node.Children)
			}
		}
	}
	walk(children)
	return snippets
}

// writeSnippetClosures emits each local snippet as a closure returning a
// child slice. A blank assignment guards against "declared and not used" for
// snippets that are declared but never rendered.
func writeSnippetClosures(buf *bytes.Buffer, snippets []*template.SnippetNode, stylesheet *style.Stylesheet, signals map[string]bool) {
	for _, s := range snippets {
		fmt.Fprintf(buf, "\t%s := func%s []*sumi.Input {\n", s.Name, snippetParamList(s.Params))
		buf.WriteString("\t\tvar cs []*sumi.Input\n")
		for _, child := range s.Children {
			writeDynamicChild(buf, child, stylesheet, 2, "\t\t", signals)
		}
		buf.WriteString("\t\treturn cs\n")
		buf.WriteString("\t}\n")
		fmt.Fprintf(buf, "\t_ = %s\n\n", s.Name)
	}
}

// snippetParamList normalises a snippet parameter clause to a Go signature
// fragment: "" becomes "()", "(name string)" is passed through.
func snippetParamList(params string) string {
	if params == "" {
		return "()"
	}
	return params
}

// writeRenderAppend emits a {render name(args)} invocation, spreading the
// snippet's result into the surrounding children slice.
func writeRenderAppend(buf *bytes.Buffer, n *template.RenderNode, tabs string) {
	fmt.Fprintf(buf, "%scs = append(cs, %s(%s)...)\n", tabs, n.Name, n.Args)
}

// validateSnippetRenders reports the first {render name} whose name resolves to
// neither a local snippet nor a snippet-typed prop. The walk does not descend
// into child component bodies — snippets there are props of the child.
func validateSnippetRenders(doc *template.Document, props []script.PropInfo) error {
	allowed := snippetNameSet(collectLocalSnippets(doc.Children))
	for _, p := range props {
		if isSnippetPropType(p.TypeStr) {
			allowed[p.Name] = true
		}
	}
	var walkErr error
	var walk func(nodes []template.Node)
	walk = func(nodes []template.Node) {
		for _, n := range nodes {
			if walkErr != nil {
				return
			}
			switch node := n.(type) {
			case *template.RenderNode:
				if !allowed[node.Name] {
					walkErr = fmt.Errorf("{render %s} names an unknown snippet: declare a {snippet %s()} or a snippet prop", node.Name, node.Name)
				}
			case *template.SnippetNode:
				walk(node.Children)
			case *template.BoxElement:
				walk(node.Children)
			case *template.IfNode:
				walk(node.Then)
				walk(node.Else)
			case *template.ForNode:
				walk(node.Children)
			}
		}
	}
	walk(doc.Children)
	return walkErr
}

// snippetNameSet returns the set of local snippet names for collision checks.
func snippetNameSet(snippets []*template.SnippetNode) map[string]bool {
	names := make(map[string]bool, len(snippets))
	for _, s := range snippets {
		names[s.Name] = true
	}
	return names
}

// isSnippetPropType reports whether a prop's Go type is a snippet closure:
// a func returning []*sumi.Input.
func isSnippetPropType(typeStr string) bool {
	return strings.HasPrefix(typeStr, "func") && strings.HasSuffix(typeStr, "[]*sumi.Input")
}

// writePropExtraction copies props into local variables. Snippet-shaped props
// are nil-defaulted to a no-op closure so {render} of an unpassed prop renders
// nothing. A prop whose name is shadowed by a local snippet is skipped (local
// snippets win resolution).
func writePropExtraction(buf *bytes.Buffer, props []script.PropInfo, localSnippets map[string]bool) {
	for _, p := range props {
		if localSnippets[p.Name] {
			continue
		}
		field := exportedName(p.Name)
		if p.Default != "" {
			fmt.Fprintf(buf, "\t%s := props.%s\n", p.Name, field)
			fmt.Fprintf(buf, "\tif %s == %s {\n", p.Name, zeroValue(p.TypeStr))
			fmt.Fprintf(buf, "\t\t%s = %s\n", p.Name, p.Default)
			fmt.Fprintf(buf, "\t}\n")
		} else {
			fmt.Fprintf(buf, "\t%s := props.%s\n", p.Name, field)
		}
		if isSnippetPropType(p.TypeStr) {
			fmt.Fprintf(buf, "\tif %s == nil {\n", p.Name)
			fmt.Fprintf(buf, "\t\t%s = %s { return nil }\n", p.Name, p.TypeStr)
			fmt.Fprintf(buf, "\t}\n")
		}
	}
	buf.WriteString("\n")
}

// writeConsumerSnippetProps emits snippet props for a mounted child component.
// {snippet name()} blocks in the tag body become named props; the remaining
// body content becomes the implicit Children snippet prop.
func writeConsumerSnippetProps(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, signals map[string]bool) {
	var rest []template.Node
	for _, c := range children {
		if snip, ok := c.(*template.SnippetNode); ok {
			writeSnippetPropField(buf, exportedName(snip.Name), snip.Params, snip.Children, stylesheet, signals)
		} else {
			rest = append(rest, c)
		}
	}
	if len(rest) > 0 {
		writeSnippetPropField(buf, "Children", "", rest, stylesheet, signals)
	}
}

// writeSnippetPropField emits a single "Name: func(params) []*sumi.Input {...}"
// prop field inside a child component's props literal.
func writeSnippetPropField(buf *bytes.Buffer, field, params string, body []template.Node, stylesheet *style.Stylesheet, signals map[string]bool) {
	fmt.Fprintf(buf, "\t\t%s: func%s []*sumi.Input {\n", field, snippetParamList(params))
	buf.WriteString("\t\t\tvar cs []*sumi.Input\n")
	for _, child := range body {
		writeDynamicChild(buf, child, stylesheet, 3, "\t\t\t", signals)
	}
	buf.WriteString("\t\t\treturn cs\n")
	buf.WriteString("\t\t},\n")
}
