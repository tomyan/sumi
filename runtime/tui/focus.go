package tui

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// initFocus gives focus to the first focusable element in the tree,
// stamping Focused flags and dispatching EventFocus to its handler.
// Called once before the initial render.
func initFocus(comp *Component) {
	focusables := layout.CollectFocusables(comp.Tree)
	if len(focusables) == 0 {
		return
	}
	comp.FocusIndex = 0
	stampFocus(focusables, 0)
	dispatchFocusEvent(focusables, 0, "focus")
}

// syncFocus re-stamps Focused flags from the component's FocusIndex.
// Called before each render so dynamically rebuilt subtrees stay correct.
func syncFocus(comp *Component) {
	focusables := layout.CollectFocusables(comp.Tree)
	if len(focusables) == 0 {
		return
	}
	if comp.FocusIndex >= len(focusables) {
		comp.FocusIndex = len(focusables) - 1
	}
	stampFocus(focusables, comp.FocusIndex)
}

// handleFocusCycle consumes Tab/Shift-Tab when the tree has focusable
// elements: blurs the current focusable, moves FocusIndex, and focuses
// the next one. Returns true if the event was consumed.
func handleFocusCycle(comp *Component, evt input.Event) bool {
	if evt.Kind != input.EventSpecial {
		return false
	}
	if evt.Special != input.KeyTab && evt.Special != input.KeyShiftTab {
		return false
	}
	focusables := layout.CollectFocusables(comp.Tree)
	if len(focusables) == 0 {
		return false
	}
	current := comp.FocusIndex
	var next int
	if evt.Special == input.KeyShiftTab {
		next = layout.CycleFocusBackward(current, len(focusables))
	} else {
		next = layout.CycleFocus(current, len(focusables))
	}
	if next != current {
		dispatchFocusEvent(focusables, current, "blur")
	}
	comp.FocusIndex = next
	stampFocus(focusables, next)
	if next != current {
		dispatchFocusEvent(focusables, next, "focus")
	}
	return true
}

// stampFocus sets Focused on each focusable according to the active index.
func stampFocus(focusables []*layout.Input, active int) {
	for i, f := range focusables {
		f.Focused = i == active
	}
}

// dispatchFocusEvent delivers a focus or blur DOM event to the focusable
// at idx. Focus and blur target the element directly — they do not bubble.
func dispatchFocusEvent(focusables []*layout.Input, idx int, eventType string) {
	if idx < 0 || idx >= len(focusables) {
		return
	}
	layout.DispatchDOM([]*layout.Input{focusables[idx]}, &layout.DOMEvent{Type: eventType})
}

// dispatchKeyToFocused bubbles a keydown/paste DOM event along the path to
// the focused element. Returns true if a handler stopped propagation, in
// which case the root component handler should not see the event.
func dispatchKeyToFocused(comp *Component, evt input.Event) bool {
	var eventType string
	switch evt.Kind {
	case input.EventKey, input.EventSpecial:
		eventType = "keydown"
	case input.EventPaste:
		eventType = "paste"
	default:
		return false
	}
	path := layout.FocusablePath(comp.Tree, comp.FocusIndex)
	if len(path) == 0 {
		return false
	}
	dom := &layout.DOMEvent{Type: eventType, Key: evt}
	layout.DispatchDOM(path, dom)
	return dom.Stopped()
}

// componentEventHandler builds the app OnEvent callback shared by
// TestApp, Run, and RunWithOptions: mouse scroll routing, click
// dispatch, focus cycling, then the component's own handler.
func componentEventHandler(app *App, comp *Component) func(input.Event) {
	return func(evt input.Event) {
		dispatchMouseScroll(evt, comp)
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			path := layout.HitTestPath(comp.Tree, comp.LayoutResult, evt.Mouse.X, evt.Mouse.Y)
			layout.DispatchDOM(path, &layout.DOMEvent{Type: "click", Key: evt})
		}
		if handleFocusCycle(comp, evt) {
			app.Dirty = true
			return
		}
		if dispatchKeyToFocused(comp, evt) {
			app.Dirty = true
			return
		}
		if comp.OnEvent != nil {
			comp.OnEvent(evt)
		}
		app.Dirty = true
	}
}
