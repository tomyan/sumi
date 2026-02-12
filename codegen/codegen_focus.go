package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/template"
)

// focusableHandler records a focusable box's onkey handler and its focus index.
type focusableHandler struct {
	FocusIndex int
	HandlerName string // the onkey handler function name
}

// collectFocusableHandlers walks the document and returns all focusable boxes with handlers.
// Handlers are collected in tree order, which determines focus index.
func collectFocusableHandlers(doc *template.Document, inlined []inlinedStateful) []focusableHandler {
	var handlers []focusableHandler
	for _, child := range doc.Children {
		handlers = collectFocusableFromNode(child, handlers, inlined)
	}
	return handlers
}

// collectFocusableFromNode recursively collects focusable handlers from a node.
func collectFocusableFromNode(node template.Node, handlers []focusableHandler, inlined []inlinedStateful) []focusableHandler {
	switch n := node.(type) {
	case *template.BoxElement:
		if n.Attributes["focusable"] == "true" {
			if handler, ok := n.Attributes["onkey"]; ok {
				handlers = append(handlers, focusableHandler{
					FocusIndex:  len(handlers),
					HandlerName: handler,
				})
			}
		}
		for _, child := range n.Children {
			handlers = collectFocusableFromNode(child, handlers, inlined)
		}
	case *template.ComponentElement:
		// Check inlined components for focusable boxes
		for _, is := range inlined {
			if is.Instance.Info.Doc != nil {
				for _, child := range is.Instance.Info.Doc.Children {
					handlers = collectInlinedFocusable(child, handlers, is.Prefix)
				}
			}
		}
	case *template.IfNode:
		for _, child := range n.Then {
			handlers = collectFocusableFromNode(child, handlers, inlined)
		}
		for _, child := range n.Else {
			handlers = collectFocusableFromNode(child, handlers, inlined)
		}
	case *template.ForNode:
		for _, child := range n.Children {
			handlers = collectFocusableFromNode(child, handlers, inlined)
		}
	}
	return handlers
}

// collectInlinedFocusable collects focusable handlers from inlined component templates.
func collectInlinedFocusable(node template.Node, handlers []focusableHandler, prefix string) []focusableHandler {
	switch n := node.(type) {
	case *template.BoxElement:
		if n.Attributes["focusable"] == "true" {
			if handler, ok := n.Attributes["onkey"]; ok {
				handlers = append(handlers, focusableHandler{
					FocusIndex:  len(handlers),
					HandlerName: prefix + handler,
				})
			}
		}
		for _, child := range n.Children {
			handlers = collectInlinedFocusable(child, handlers, prefix)
		}
	}
	return handlers
}

// findNonFocusableHandlers returns handler names that are on non-focusable boxes (for bubbling).
func findNonFocusableHandlers(doc *template.Document, inlined []inlinedStateful) []string {
	var handlers []string
	for _, child := range doc.Children {
		handlers = findNonFocusableFromNode(child, handlers, inlined)
	}
	return handlers
}

func findNonFocusableFromNode(node template.Node, handlers []string, inlined []inlinedStateful) []string {
	switch n := node.(type) {
	case *template.BoxElement:
		if n.Attributes["focusable"] != "true" {
			if handler, ok := n.Attributes["onkey"]; ok {
				handlers = append(handlers, handler)
			}
		}
		for _, child := range n.Children {
			handlers = findNonFocusableFromNode(child, handlers, inlined)
		}
	case *template.IfNode:
		for _, child := range n.Then {
			handlers = findNonFocusableFromNode(child, handlers, inlined)
		}
		for _, child := range n.Else {
			handlers = findNonFocusableFromNode(child, handlers, inlined)
		}
	case *template.ForNode:
		for _, child := range n.Children {
			handlers = findNonFocusableFromNode(child, handlers, inlined)
		}
	}
	return handlers
}

// writeFocusStateDecls writes focus state variables if focusable boxes exist.
func writeFocusStateDecls(buf *bytes.Buffer, focusHandlers []focusableHandler) {
	if len(focusHandlers) == 0 {
		return
	}
	fmt.Fprintf(buf, "\tfocusIndex := -1\n")
	fmt.Fprintf(buf, "\tfocusCount := %d\n", len(focusHandlers))
	buf.WriteString("\tpropagationStopped := false\n")
	buf.WriteString("\tstopPropagation := func() { propagationStopped = true }\n")
	buf.WriteString("\n")
}

// writeSuppressFocusVars writes _ = var for focus variables not referenced in closures.
func writeSuppressFocusVars(buf *bytes.Buffer, focusHandlers []focusableHandler) {
	if len(focusHandlers) == 0 {
		return
	}
	buf.WriteString("\t_ = stopPropagation\n")
}

// writeFocusDispatch writes focus-directed event dispatch in the OnEvent closure.
func writeFocusDispatch(buf *bytes.Buffer, focusHandlers []focusableHandler,
	bubblingHandlers []string, eventAware map[string]bool) {

	// Tab cycling — includes -1 (unfocused) as a valid position
	buf.WriteString("\t\t\tif evt.Kind == input.EventSpecial {\n")
	buf.WriteString("\t\t\t\tif evt.Special == input.KeyTab {\n")
	buf.WriteString("\t\t\t\t\tfocusIndex = (focusIndex + 2) % (focusCount + 1) - 1\n")
	buf.WriteString("\t\t\t\t\tapp.Dirty = true\n")
	buf.WriteString("\t\t\t\t\treturn\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tif evt.Special == input.KeyShiftTab {\n")
	buf.WriteString("\t\t\t\t\tfocusIndex = (focusIndex + focusCount + 1) % (focusCount + 1) - 1\n")
	buf.WriteString("\t\t\t\t\tapp.Dirty = true\n")
	buf.WriteString("\t\t\t\t\treturn\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t}\n")

	// Focus-directed dispatch
	buf.WriteString("\t\t\tpropagationStopped = false\n")
	buf.WriteString("\t\t\tswitch focusIndex {\n")
	for _, fh := range focusHandlers {
		fmt.Fprintf(buf, "\t\t\tcase %d:\n", fh.FocusIndex)
		if eventAware[fh.HandlerName] {
			fmt.Fprintf(buf, "\t\t\t\t%s(evt)\n", fh.HandlerName)
		} else {
			fmt.Fprintf(buf, "\t\t\t\t%s()\n", fh.HandlerName)
		}
	}
	buf.WriteString("\t\t\t}\n")

	// Bubbling to non-focusable handlers
	for _, handler := range bubblingHandlers {
		buf.WriteString("\t\t\tif !propagationStopped {\n")
		if eventAware[handler] {
			fmt.Fprintf(buf, "\t\t\t\t%s(evt)\n", handler)
		} else {
			fmt.Fprintf(buf, "\t\t\t\t%s()\n", handler)
		}
		buf.WriteString("\t\t\t}\n")
	}
}
