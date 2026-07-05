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
	for i, f := range focusables {
		switch f.Tag {
		case "input":
			syncInputElement(f, i == comp.FocusIndex)
		case "select":
			syncSelectElement(f)
		}
	}
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

// focusClickedElement moves focus to the deepest focusable element on a
// clicked path (click-to-focus), dispatching blur and focus events.
func focusClickedElement(comp *Component, path []*layout.Input) {
	var target *layout.Input
	for i := len(path) - 1; i >= 0; i-- {
		if layout.IsFocusable(path[i]) {
			target = path[i]
			break
		}
	}
	if target == nil {
		return
	}
	focusables := layout.CollectFocusables(comp.Tree)
	idx := -1
	for i, f := range focusables {
		if f == target {
			idx = i
			break
		}
	}
	if idx < 0 || idx == comp.FocusIndex {
		return
	}
	dispatchFocusEvent(focusables, comp.FocusIndex, "blur")
	comp.FocusIndex = idx
	stampFocus(focusables, idx)
	dispatchFocusEvent(focusables, idx, "focus")
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
// the focused element. Returns the dispatched event so callers can inspect
// Stopped/DefaultPrevented, or nil when nothing was dispatched.
func dispatchKeyToFocused(comp *Component, evt input.Event) *layout.DOMEvent {
	var eventType string
	switch evt.Kind {
	case input.EventKey, input.EventSpecial:
		eventType = "keydown"
	case input.EventPaste:
		eventType = "paste"
	default:
		return nil
	}
	path := layout.FocusablePath(comp.Tree, comp.FocusIndex)
	if len(path) == 0 {
		return nil
	}
	dom := &layout.DOMEvent{Type: eventType, Key: evt}
	layout.DispatchDOM(path, dom)
	return dom
}

// applyDefaultActions runs the built-in behavior for an event after DOM
// dispatch, unless a handler called PreventDefault. Returns true when the
// event is consumed by a default action.
func applyDefaultActions(comp *Component, evt input.Event, dom *layout.DOMEvent) bool {
	if dom != nil && dom.DefaultPrevented() {
		return false
	}
	if handleFocusCycle(comp, evt) {
		return true
	}
	if editFocusedInput(comp, evt) {
		return true
	}
	if selectKeydown(comp, evt) {
		return true
	}
	return activateFocused(comp, evt)
}

// activateFocused synthesizes a bubbling click on the focused element.
// Enter activates anything activatable (click handler, anchor with href,
// or checkable); Space activates checkables only — it types into text
// inputs and passes through everything else.
func activateFocused(comp *Component, evt input.Event) bool {
	isEnter := evt.Kind == input.EventSpecial && evt.Special == input.KeyEnter
	isSpace := evt.Kind == input.EventKey && !evt.Ctrl && evt.Rune == ' '
	if !isEnter && !isSpace {
		return false
	}
	path := layout.FocusablePath(comp.Tree, comp.FocusIndex)
	if len(path) == 0 {
		return false
	}
	target := path[len(path)-1]
	checkable := isCheckable(target)
	if isSpace && !checkable {
		return false
	}
	isAnchor := target.Tag == "a" && target.Attrs["href"] != ""
	if isEnter && target.On["click"] == nil && !isAnchor && !checkable {
		return false
	}
	dom := &layout.DOMEvent{Type: "click", Key: evt}
	layout.DispatchDOM(path, dom)
	if !dom.DefaultPrevented() {
		runClickDefault(comp, path, evt)
	}
	return true
}

// runClickDefault performs the default action for a click, walking from
// the target upward: toggle the first checkable, open the first anchor's
// href, or activate the first label's control.
func runClickDefault(comp *Component, path []*layout.Input, evt input.Event) {
	clickDefault(comp, path, evt, true)
}

func clickDefault(comp *Component, path []*layout.Input, evt input.Event, followLabels bool) {
	for i := len(path) - 1; i >= 0; i-- {
		n := path[i]
		if isCheckable(n) {
			toggleCheckable(comp, path[:i+1], n, evt)
			return
		}
		if n.Tag == "select" {
			moveSelect(comp, path[:i+1], n, 1, evt)
			return
		}
		if n.Tag == "a" {
			if href := n.Attrs["href"]; href != "" {
				OpenURL(href)
				return
			}
		}
		if followLabels && n.Tag == "label" && activateLabel(comp, n, evt) {
			return
		}
	}
}

// componentEventHandler builds the app OnEvent callback shared by
// TestApp, Run, and RunWithOptions: mouse scroll routing, click
// dispatch, focus cycling, then the component's own handler.
func componentEventHandler(app *App, comp *Component) func(input.Event) {
	return func(evt input.Event) {
		dispatchMouseScroll(evt, comp)
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			path := layout.HitTestPath(comp.Tree, comp.LayoutResult, evt.Mouse.X, evt.Mouse.Y)
			focusClickedElement(comp, path)
			dom := &layout.DOMEvent{Type: "click", Key: evt}
			layout.DispatchDOM(path, dom)
			if !dom.DefaultPrevented() {
				runClickDefault(comp, path, evt)
			}
		}
		dom := dispatchKeyToFocused(comp, evt)
		if applyDefaultActions(comp, evt, dom) {
			app.Dirty = true
			return
		}
		if dom != nil && dom.Stopped() {
			app.Dirty = true
			return
		}
		if comp.OnEvent != nil {
			comp.OnEvent(evt)
		}
		app.Dirty = true
	}
}
