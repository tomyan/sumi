package tui

import (
	"unicode/utf8"

	"github.com/tomyan/sumi/runtime/edit"
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// ensureEditState lazily initializes an input element's editing state
// from its value attribute (literals only — expression values are wired
// by author handlers) with the cursor at the end.
func ensureEditState(n *layout.Input) *edit.State {
	if n.Edit == nil {
		value := n.Attrs["value"]
		if len(value) > 0 && value[0] == '{' {
			value = ""
		}
		n.Edit = &edit.State{Value: value, Cursor: utf8.RuneCountInString(value)}
	}
	return n.Edit
}

// syncInputElement projects editing state onto the element: the value
// renders in an implicit text child, and the cursor shows only while
// the element is focused.
func syncInputElement(n *layout.Input, focused bool) {
	state := ensureEditState(n)
	child := ensureValueChild(n)
	child.Content = state.Value
	if focused {
		n.CursorCol = state.Cursor
		n.CursorRow = 0
	} else {
		n.CursorCol = -1
	}
}

// ensureValueChild returns the implicit text child that displays the value.
func ensureValueChild(n *layout.Input) *layout.Input {
	for _, c := range n.Children {
		if c != nil && c.Kind == layout.KindText {
			return c
		}
	}
	child := &layout.Input{Kind: layout.KindText, CursorCol: -1, CursorRow: -1}
	n.Children = append(n.Children, child)
	return child
}

// editFocusedInput applies an editing key to the focused input element as
// its keydown default action, then dispatches an "input" DOM event with
// the new value and cursor. Returns true when the key was consumed.
func editFocusedInput(comp *Component, evt input.Event) bool {
	path := layout.FocusablePath(comp.Tree, comp.FocusIndex)
	if len(path) == 0 {
		return false
	}
	target := path[len(path)-1]
	if target.Tag != "input" {
		return false
	}
	state := ensureEditState(target)
	if !edit.HandleKey(state, evt) {
		return false
	}
	syncInputElement(target, true)
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "input", Key: evt,
		Data: map[string]any{"value": state.Value, "cursor": state.Cursor}})
	return true
}
