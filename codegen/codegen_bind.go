package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/template"
)

// bindInfo describes a native two-way binding declared with bind:value or
// bind:checked on a form control. The update half is wired as a DOM event
// handler that pushes the control's new value into the signal; the display
// half re-projects the signal onto the control in the sync effect.
type bindInfo struct {
	signal    string // bound signal expression (e.g. "count")
	eventType string // DOM event that carries the new value ("input" or "change")
	dataKey   string // event Data key holding the value ("value" or "checked")
	dataType  string // Go type of the value ("string" or "bool")
}

// isNativeControlTag reports whether a tag is a form control that supports
// native bindings.
func isNativeControlTag(tag string) bool {
	switch tag {
	case "input", "textarea", "select":
		return true
	}
	return false
}

// bindOf returns the binding declared on a native control, or nil. The bind
// attribute name selects the semantics: bind:checked drives a checkbox/radio
// via the change event's bool "checked"; bind:value drives text controls via
// the input event and selects via the change event, both with a string
// "value".
func bindOf(tag string, attrs map[string]string) *bindInfo {
	if !isNativeControlTag(tag) {
		return nil
	}
	if v, ok := attrs["bind:checked"]; ok {
		return &bindInfo{signal: extractExprValue(v),
			eventType: "change", dataKey: "checked", dataType: "bool"}
	}
	if v, ok := attrs["bind:value"]; ok {
		event := "input"
		if tag == "select" {
			event = "change"
		}
		return &bindInfo{signal: extractExprValue(v),
			eventType: event, dataKey: "value", dataType: "string"}
	}
	return nil
}

// hasBind reports whether the element carries a native binding, which forces
// extraction as a named variable so the display half can patch it in sync.
func hasBind(tag string, attrs map[string]string) bool {
	return bindOf(tag, attrs) != nil
}

// writeBindHandler emits the update-half handler entry inside the On map.
func writeBindHandler(buf *bytes.Buffer, tabs string, b *bindInfo) {
	fmt.Fprintf(buf, "%s\t\t%q: func(evt *sumi.DOMEvent) {\n", tabs, b.eventType)
	fmt.Fprintf(buf, "%s\t\t\t%s.Set(evt.Data[%q].(%s))\n", tabs, b.signal, b.dataKey, b.dataType)
	fmt.Fprintf(buf, "%s\t\t},\n", tabs)
}

// writeBindSync emits the display-half statement into the sync effect: the
// control re-projects the signal each time it changes, so an external Set is
// reflected. Typing already keeps the two equal, making this a no-op then.
func writeBindSync(buf *bytes.Buffer, name, tag string, attrs map[string]string) {
	b := bindOf(tag, attrs)
	if b == nil {
		return
	}
	switch {
	case b.dataKey == "checked":
		fmt.Fprintf(buf, "\t\tsumi.BindChecked(%s, %s.Get())\n", name, b.signal)
	case tag == "select":
		fmt.Fprintf(buf, "\t\tsumi.BindSelectValue(%s, %s.Get())\n", name, b.signal)
	default:
		fmt.Fprintf(buf, "\t\tsumi.BindInputValue(%s, %s.Get())\n", name, b.signal)
	}
}

// validateBindConflicts reports the first element that declares both a
// binding and an author handler for the binding's own event — the two would
// collide on the On map, and silently preferring either hides code.
func validateBindConflicts(doc *template.Document) error {
	var walkErr error
	var walk func(nodes []template.Node)
	walk = func(nodes []template.Node) {
		for _, n := range nodes {
			if walkErr != nil {
				return
			}
			switch node := n.(type) {
			case *template.BoxElement:
				if err := bindConflict(node.Tag, node.Attributes); err != nil {
					walkErr = err
					return
				}
				walk(node.Children)
			case *template.TextElement:
				if err := bindConflict(node.Tag, node.Attributes); err != nil {
					walkErr = err
					return
				}
			case *template.SnippetNode:
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

func bindConflict(tag string, attrs map[string]string) error {
	b := bindOf(tag, attrs)
	if b == nil {
		return nil
	}
	if _, ok := attrs["on"+b.eventType]; ok {
		return fmt.Errorf("on%s conflicts with bind:%s on <%s>: the binding wires %s; remove the handler or the binding",
			b.eventType, b.dataKey, tag, b.eventType)
	}
	return nil
}
