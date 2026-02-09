package codegen

import (
	"bytes"
	"fmt"
	"go/format"

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
	buf.WriteString("\t\"github.com/tomyan/sumi/runtime/render\"\n")
	buf.WriteString(")\n\n")

	buf.WriteString("func Run() {\n")
	buf.WriteString("\tbuf := render.NewBuffer(80, 24)\n")
	buf.WriteString("\trender.EnterAlternateScreen(os.Stdout)\n")

	row := 0
	for _, child := range doc.Children {
		switch n := child.(type) {
		case *template.TextElement:
			fmt.Fprintf(&buf, "\tbuf.WriteText(%d, 0, %q)\n", row, n.Content)
			row++
		}
	}

	buf.WriteString("\tbuf.RenderTo(os.Stdout)\n")
	buf.WriteString("\tbufio.NewScanner(os.Stdin).Scan()\n")
	buf.WriteString("\trender.ExitAlternateScreen(os.Stdout)\n")
	buf.WriteString("}\n")

	return format.Source(buf.Bytes())
}
