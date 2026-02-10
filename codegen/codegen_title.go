package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/template"
)

// findTitleElement returns the first TitleElement in the document, or nil.
func findTitleElement(doc *template.Document) *template.TitleElement {
	for _, child := range doc.Children {
		if title, ok := child.(*template.TitleElement); ok {
			return title
		}
	}
	return nil
}

// writeTitleSave writes the xterm title save sequence if a title element exists.
func writeTitleSave(buf *bytes.Buffer, title *template.TitleElement) {
	if title == nil {
		return
	}
	buf.WriteString("\tfmt.Fprint(os.Stdout, \"\\033[22;2t\")\n")
}

// writeTitleRestore writes a deferred xterm title restore sequence.
func writeTitleRestore(buf *bytes.Buffer, title *template.TitleElement) {
	if title == nil {
		return
	}
	buf.WriteString("\tdefer fmt.Fprint(os.Stdout, \"\\033[23;2t\")\n")
}

// writeTitleSet writes the OSC escape sequence to set the terminal title.
// Called inside doRender so the title updates when state changes.
func writeTitleSet(buf *bytes.Buffer, title *template.TitleElement) {
	if title == nil {
		return
	}
	if isStaticTitle(title) {
		writeStaticTitle(buf, title)
	} else {
		writeDynamicTitle(buf, title)
	}
}

// isStaticTitle returns true if all parts are StringParts.
func isStaticTitle(title *template.TitleElement) bool {
	for _, part := range title.Parts {
		if _, ok := part.(*template.ExprPart); ok {
			return false
		}
	}
	return true
}

// writeStaticTitle writes an OSC title set with a static string.
func writeStaticTitle(buf *bytes.Buffer, title *template.TitleElement) {
	var content string
	for _, part := range title.Parts {
		if sp, ok := part.(*template.StringPart); ok {
			content += sp.Value
		}
	}
	fmt.Fprintf(buf, "\t\tfmt.Fprint(os.Stdout, \"\\033]2;%s\\007\")\n", content)
}

// writeDynamicTitle writes an OSC title set using fmt.Fprintf for expressions.
func writeDynamicTitle(buf *bytes.Buffer, title *template.TitleElement) {
	format, args := buildTitleFormatArgs(title)
	if len(args) == 0 {
		fmt.Fprintf(buf, "\t\tfmt.Fprint(os.Stdout, \"\\033]2;%s\\007\")\n", format)
	} else {
		fmt.Fprintf(buf, "\t\tfmt.Fprintf(os.Stdout, \"\\033]2;%s\\007\", %s)\n", format, joinArgs(args))
	}
}

// buildTitleFormatArgs builds a format string and argument list for a dynamic title.
func buildTitleFormatArgs(title *template.TitleElement) (string, []string) {
	var format string
	var args []string
	for _, part := range title.Parts {
		switch p := part.(type) {
		case *template.StringPart:
			format += p.Value
		case *template.ExprPart:
			format += "%v"
			args = append(args, p.Expr)
		}
	}
	return format, args
}

// joinArgs joins argument expressions with ", ".
func joinArgs(args []string) string {
	result := args[0]
	for _, a := range args[1:] {
		result += ", " + a
	}
	return result
}
