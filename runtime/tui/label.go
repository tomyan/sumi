package tui

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// labellable reports whether a node is a form control a label can
// activate (matching svelterm's LABELLABLE set).
func labellable(n *layout.Input) bool {
	switch n.Tag {
	case "input", "button", "textarea", "select":
		return true
	}
	return false
}

// labelControl resolves the control a label is associated with: the
// element referenced by for="id", or the first control the label wraps.
func labelControl(root, label *layout.Input) *layout.Input {
	if id := label.Attrs["for"]; id != "" {
		return findByID(root, id)
	}
	return firstControl(label)
}

func findByID(n *layout.Input, id string) *layout.Input {
	if n == nil {
		return nil
	}
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := findByID(c, id); found != nil {
			return found
		}
	}
	return nil
}

func firstControl(n *layout.Input) *layout.Input {
	for _, c := range n.Children {
		if c == nil {
			continue
		}
		if labellable(c) {
			return c
		}
		if found := firstControl(c); found != nil {
			return found
		}
	}
	return nil
}

// activateLabel focuses the label's control and synthesizes a click on
// it, running the control's own default action (toggle, open) — but not
// label-following again, which would recurse on wrapping labels.
func activateLabel(comp *Component, label *layout.Input, evt input.Event) bool {
	control := labelControl(comp.Tree, label)
	if control == nil {
		return false
	}
	path := layout.PathTo(comp.Tree, control)
	if len(path) == 0 {
		return false
	}
	focusClickedElement(comp, path)
	dom := &layout.DOMEvent{Type: "click", Key: evt}
	layout.DispatchDOM(path, dom)
	if !dom.DefaultPrevented() {
		clickDefault(comp, path, evt, false)
	}
	return true
}
