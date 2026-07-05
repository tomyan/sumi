package tui

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// syncDialogElement hides a dialog that lacks the open attribute.
// Uses the runtime Hidden flag: display belongs to the cascade, which
// would overwrite a projection-written value on the next resolve pass.
func syncDialogElement(n *layout.Input) {
	n.Hidden = !boolAttr(n, "open")
}

// findOpenDialog returns the first open dialog in tree order, or nil.
func findOpenDialog(n *layout.Input) *layout.Input {
	if n == nil {
		return nil
	}
	if n.Tag == "dialog" && boolAttr(n, "open") {
		return n
	}
	for _, c := range n.Children {
		if d := findOpenDialog(c); d != nil {
			return d
		}
	}
	return nil
}

// focusScope returns the root of focus traversal: an open dialog traps
// focus inside its subtree; otherwise the whole tree.
func focusScope(comp *Component) *layout.Input {
	if d := findOpenDialog(comp.Tree); d != nil {
		return d
	}
	return comp.Tree
}

// scopedFocusables lists the focusables within the current focus scope.
func scopedFocusables(comp *Component) []*layout.Input {
	return layout.CollectFocusables(focusScope(comp))
}

// focusedPath returns the full root→focused-element path, crossing the
// scope boundary so events still bubble to the tree root.
func focusedPath(comp *Component) []*layout.Input {
	scope := focusScope(comp)
	inner := layout.FocusablePath(scope, comp.FocusIndex)
	if len(inner) == 0 {
		return nil
	}
	if scope == comp.Tree {
		return inner
	}
	outer := layout.PathTo(comp.Tree, scope)
	if len(outer) == 0 {
		return inner
	}
	return append(outer[:len(outer)-1:len(outer)-1], inner...)
}

// pathInFocusScope reports whether a hit path reaches into the current
// focus scope. A modal dialog captures interactions outside it.
func pathInFocusScope(comp *Component, path []*layout.Input) bool {
	scope := focusScope(comp)
	if scope == comp.Tree {
		return true
	}
	for _, n := range path {
		if n == scope {
			return true
		}
	}
	return false
}

// closeDialog removes the open attribute, hides the dialog, dispatches a
// close event on it, and returns focus to the page.
func closeDialog(comp *Component, dialog *layout.Input, evt input.Event) {
	delete(dialog.Attrs, "open")
	syncDialogElement(dialog)
	path := layout.PathTo(comp.Tree, dialog)
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "close", Key: evt})
	comp.FocusIndex = 0
}

// dialogEscape closes the trapping dialog on Escape.
func dialogEscape(comp *Component, evt input.Event) bool {
	if evt.Kind != input.EventSpecial || evt.Special != input.KeyEscape {
		return false
	}
	dialog := findOpenDialog(comp.Tree)
	if dialog == nil {
		return false
	}
	closeDialog(comp, dialog, evt)
	return true
}
