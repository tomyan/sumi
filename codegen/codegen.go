package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strconv"

	"github.com/tomyan/sumi/parser/template"
)

// Options configures code generation.
type Options struct {
	PackageName string
}

// Generate produces Go source code from a template AST.
func Generate(doc *template.Document, opts Options) ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "package %s\n\n", opts.PackageName)

	buf.WriteString("import (\n")
	buf.WriteString("\t\"bufio\"\n")
	buf.WriteString("\t\"os\"\n\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/layout\"\n")
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	buf.WriteString(")\n\n")

	buf.WriteString("func Run() {\n")

	// Build the layout input tree from the template AST.
	// Wrap document children in a root column box.
	buf.WriteString("\troot := &layout.Input{\n")
	buf.WriteString("\t\tKind:      layout.KindBox,\n")
	buf.WriteString("\t\tDirection: \"column\",\n")
	buf.WriteString("\t\tChildren:  []*layout.Input{\n")
	for _, child := range doc.Children {
		writeInputNode(&buf, child, 3)
	}
	buf.WriteString("\t\t},\n")
	buf.WriteString("\t}\n")

	buf.WriteString("\ttree := layout.Layout(root, 80, 24)\n")
	buf.WriteString("\tbuf := render.NewBuffer(80, 24)\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")
	buf.WriteString("\trenderTree(buf, tree)\n")
	buf.WriteString("\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\tbufio.NewScanner(os.Stdin).Scan()\n")
	buf.WriteString("\trender.ExitAlternateScreen(os.Stdout)\n")
	buf.WriteString("}\n\n")

	// Generate the renderTree helper function.
	buf.WriteString("func renderTree(buf *render.Buffer, box *layout.Box) {\n")
	buf.WriteString("\tif box.Border != \"\" && box.Border != \"none\" {\n")
	buf.WriteString("\t\tbuf.DrawBorder(box.Y, box.X, box.Width, box.Height, box.Border)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif box.Content != \"\" {\n")
	buf.WriteString("\t\tbuf.WriteText(box.Y, box.X, box.Content)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tfor _, child := range box.Children {\n")
	buf.WriteString("\t\trenderTree(buf, child)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")

	return format.Source(buf.Bytes())
}

// writeInputNode writes a layout.Input literal for a template AST node.
func writeInputNode(buf *bytes.Buffer, node template.Node, indent int) {
	tabs := indentStr(indent)

	switch n := node.(type) {
	case *template.TextElement:
		fmt.Fprintf(buf, "%s{\n", tabs)
		fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
		fmt.Fprintf(buf, "%s\tContent: %q,\n", tabs, n.Content)
		fmt.Fprintf(buf, "%s},\n", tabs)

	case *template.BoxElement:
		fmt.Fprintf(buf, "%s{\n", tabs)
		fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)

		if dir, ok := n.Attributes["direction"]; ok {
			fmt.Fprintf(buf, "%s\tDirection: %q,\n", tabs, dir)
		}
		if w, ok := n.Attributes["width"]; ok {
			if v, err := strconv.Atoi(w); err == nil {
				fmt.Fprintf(buf, "%s\tFixedWidth:  %d,\n", tabs, v)
			}
		}
		if h, ok := n.Attributes["height"]; ok {
			if v, err := strconv.Atoi(h); err == nil {
				fmt.Fprintf(buf, "%s\tFixedHeight: %d,\n", tabs, v)
			}
		}
		if p, ok := n.Attributes["padding"]; ok {
			fmt.Fprintf(buf, "%s\tPadding: layout.ParsePadding(%q),\n", tabs, p)
		}
		if b, ok := n.Attributes["border"]; ok {
			fmt.Fprintf(buf, "%s\tBorder: %q,\n", tabs, b)
		}

		if len(n.Children) > 0 {
			fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
			for _, child := range n.Children {
				writeInputNode(buf, child, indent+2)
			}
			fmt.Fprintf(buf, "%s\t},\n", tabs)
		}

		fmt.Fprintf(buf, "%s},\n", tabs)
	}
}

func indentStr(n int) string {
	s := ""
	for range n {
		s += "\t"
	}
	return s
}
