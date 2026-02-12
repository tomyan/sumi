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

// buildStaticTitleString returns the concatenated string content of a static title.
func buildStaticTitleString(title *template.TitleElement) string {
	var content string
	for _, part := range title.Parts {
		if sp, ok := part.(*template.StringPart); ok {
			content += sp.Value
		}
	}
	return content
}

// writeTitleSet writes the OSC escape sequence to set the terminal title.
// Called inside doRender for dynamic titles that update when state changes.
// Static titles are handled by App.Title, so only dynamic titles emit here.
func writeTitleSet(buf *bytes.Buffer, title *template.TitleElement) {
	if title == nil || isStaticTitle(title) {
		return
	}
	writeDynamicTitle(buf, title)
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
