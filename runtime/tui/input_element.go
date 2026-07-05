package tui

import (
	"strconv"
	"strings"
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

// syncInputElement projects element state onto the tree. Checkables
// render their glyph with no caret; text inputs render the value —
// masked for password inputs, windowed to the laid-out width — in an
// implicit text child, with the cursor shown only while focused.
func syncInputElement(n *layout.Input, focused bool) {
	if isCheckable(n) {
		ensureValueChild(n).Content = checkableGlyph(n)
		n.CursorCol = -1
		return
	}
	state := ensureEditState(n)
	display := state.Value
	if n.Attrs["type"] == "password" {
		display = strings.Repeat("•", utf8.RuneCountInString(state.Value))
	}
	display = windowDisplay(n, state, display)
	ensureValueChild(n).Content = display
	if focused {
		n.CursorCol = state.Cursor - state.ViewOffset
		n.CursorRow = 0
	} else {
		n.CursorCol = -1
	}
}

// windowDisplay slides the view offset so the cursor stays visible when
// the value exceeds the element's laid-out content width.
func windowDisplay(n *layout.Input, state *edit.State, display string) string {
	contentW := inputContentWidth(n)
	if contentW <= 0 {
		state.ViewOffset = 0
		return display
	}
	runes := []rune(display)
	maxOffset := 0
	if len(runes)+1 > contentW {
		maxOffset = len(runes) + 1 - contentW // +1 keeps a cell for the caret
	}
	if state.ViewOffset > state.Cursor {
		state.ViewOffset = state.Cursor
	}
	if state.Cursor-state.ViewOffset > contentW-1 {
		state.ViewOffset = state.Cursor - (contentW - 1)
	}
	if state.ViewOffset > maxOffset {
		state.ViewOffset = maxOffset
	}
	if state.ViewOffset < 0 {
		state.ViewOffset = 0
	}
	end := state.ViewOffset + contentW
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[state.ViewOffset:end])
}

// inputContentWidth returns the width available for the value inside the
// element's borders and padding, from the previous layout pass (falling
// back to the resolved fixed width before the first layout).
func inputContentWidth(n *layout.Input) int {
	w := n.LastW
	if w <= 0 {
		w = n.FixedWidth
	}
	if w <= 0 {
		return 0
	}
	if n.Border != "" && n.Border != "none" {
		w -= 2
	}
	return w - n.Padding.Left - n.Padding.Right
}

// ensureValueChild returns the implicit untagged text child that displays
// the element's projected value (input text, select label).
func ensureValueChild(n *layout.Input) *layout.Input {
	for _, c := range n.Children {
		if c != nil && c.Kind == layout.KindText && c.Tag == "" {
			return c
		}
	}
	child := &layout.Input{Kind: layout.KindText, CursorCol: -1, CursorRow: -1}
	n.Children = append(n.Children, child)
	return child
}

// syncProjection projects UA element state (input value, select label,
// progress/meter bar) into the element's implicit child.
func syncProjection(n *layout.Input) {
	switch n.Tag {
	case "input":
		syncInputElement(n, n.Focused)
	case "select":
		syncSelectElement(n)
	case "progress", "meter":
		syncBarElement(n)
	case "details":
		syncDetailsElement(n)
	}
}

// syncProjections walks the tree projecting all UA elements.
func syncProjections(root *layout.Input) {
	if root == nil {
		return
	}
	syncProjection(root)
	for _, c := range root.Children {
		syncProjections(c)
	}
}

// resyncInputElements re-projects UA elements after layout, when the
// laid-out width (LastW) is fresh — value windowing and bar widths
// depend on it. Returns true when any projection changed, requesting
// another converge pass.
func resyncInputElements(comp *Component) bool {
	changed := false
	var walk func(n *layout.Input)
	walk = func(n *layout.Input) {
		if n == nil {
			return
		}
		switch n.Tag {
		case "input", "select", "progress", "meter":
			child := ensureValueChild(n)
			beforeContent, beforeCursor := child.Content, n.CursorCol
			syncProjection(n)
			if child.Content != beforeContent || n.CursorCol != beforeCursor {
				changed = true
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(comp.Tree)
	return changed
}

// inputConstraints reads maxlength/readonly attributes.
func inputConstraints(n *layout.Input) edit.Constraints {
	c := edit.Constraints{}
	if v, ok := n.Attrs["maxlength"]; ok {
		if max, err := strconv.Atoi(v); err == nil {
			c.MaxLength = max
		}
	}
	if v, ok := n.Attrs["readonly"]; ok && v != "false" {
		c.ReadOnly = true
	}
	return c
}

// editFocusedInput applies an editing key to the focused input element as
// its keydown default action. An "input" DOM event fires only when the
// value actually changed. Returns true when the key was consumed.
func editFocusedInput(comp *Component, evt input.Event) bool {
	path := layout.FocusablePath(comp.Tree, comp.FocusIndex)
	if len(path) == 0 {
		return false
	}
	target := path[len(path)-1]
	if target.Tag != "input" || isCheckable(target) {
		return false
	}
	state := ensureEditState(target)
	before := state.Value
	if !edit.HandleKeyWith(state, evt, inputConstraints(target)) {
		return false
	}
	syncInputElement(target, true)
	if state.Value != before {
		layout.DispatchDOM(path, &layout.DOMEvent{Type: "input", Key: evt,
			Data: map[string]any{"value": state.Value, "cursor": state.Cursor}})
	}
	return true
}
