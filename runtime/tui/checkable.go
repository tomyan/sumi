package tui

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// isCheckable reports whether an input element toggles rather than edits.
func isCheckable(n *layout.Input) bool {
	t := n.Attrs["type"]
	return n.Tag == "input" && (t == "checkbox" || t == "radio")
}

// isChecked reads the checked state attribute.
func isChecked(n *layout.Input) bool {
	v, ok := n.Attrs["checked"]
	return ok && v != "false"
}

// setChecked writes the checked state attribute (attribute-backed state
// drives :checked selectors).
func setChecked(n *layout.Input, checked bool) {
	if n.Attrs == nil {
		n.Attrs = map[string]string{}
	}
	if checked {
		n.Attrs["checked"] = "true"
	} else {
		delete(n.Attrs, "checked")
	}
}

// checkableGlyph renders the checkbox/radio state.
func checkableGlyph(n *layout.Input) string {
	if n.Attrs["type"] == "radio" {
		if isChecked(n) {
			return "(•)"
		}
		return "( )"
	}
	if isChecked(n) {
		return "[x]"
	}
	return "[ ]"
}

// toggleCheckable flips (or, for radios, sets) the checked state and
// dispatches change and input events with {checked, value}. A radio
// checks itself, unchecks same-name radios across the tree, and never
// untoggles itself.
func toggleCheckable(comp *Component, path []*layout.Input, target *layout.Input, key input.Event) {
	if target.Attrs["type"] == "radio" {
		if isChecked(target) {
			return
		}
		uncheckRadioGroup(comp.Tree, target)
		setChecked(target, true)
	} else {
		setChecked(target, !isChecked(target))
	}
	syncInputElement(target, target.Focused)
	data := map[string]any{"checked": isChecked(target), "value": target.Attrs["value"]}
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "change", Key: key, Data: data})
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "input", Key: key, Data: data})
}

// uncheckRadioGroup clears every other same-name radio in the tree.
func uncheckRadioGroup(root, target *layout.Input) {
	name := target.Attrs["name"]
	var walk func(n *layout.Input)
	walk = func(n *layout.Input) {
		if n == nil {
			return
		}
		if n != target && n.Tag == "input" && n.Attrs["type"] == "radio" && n.Attrs["name"] == name {
			setChecked(n, false)
			syncInputElement(n, n.Focused)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
}
