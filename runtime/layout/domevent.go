package layout

import "github.com/tomyan/sumi/runtime/input"

// DOMEvent is an event dispatched through the element tree. Handlers
// receive the same event instance as it bubbles from the target toward
// the root, so StopPropagation and PreventDefault affect later handlers.
type DOMEvent struct {
	Type   string         // "click", "keydown", "focus", "blur", "input", "change", "paste", "toggle"
	Key    input.Event    // underlying terminal event, when there is one
	Data   map[string]any // event payload (e.g. value/cursor for "input")
	Target *Input         // deepest element on the dispatch path

	stopped          bool
	defaultPrevented bool
}

// StopPropagation prevents handlers further up the path from running.
func (e *DOMEvent) StopPropagation() { e.stopped = true }

// PreventDefault suppresses the default action that would follow dispatch.
func (e *DOMEvent) PreventDefault() { e.defaultPrevented = true }

// DefaultPrevented reports whether PreventDefault was called.
func (e *DOMEvent) DefaultPrevented() bool { return e.defaultPrevented }

// DispatchDOM bubbles evt from the deepest node in path toward the root,
// calling each node's handler for evt.Type. path runs root → deepest,
// as returned by HitTestPath.
func DispatchDOM(path []*Input, evt *DOMEvent) {
	if len(path) == 0 {
		return
	}
	evt.Target = path[len(path)-1]
	for i := len(path) - 1; i >= 0; i-- {
		if h := path[i].On[evt.Type]; h != nil {
			h(evt)
			if evt.stopped {
				return
			}
		}
	}
}
