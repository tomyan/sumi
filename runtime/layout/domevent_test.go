package layout

import "testing"

func TestDispatchDOMBubblesDeepestToRoot(t *testing.T) {
	// Given — a three-deep path with handlers recording call order
	var order []string
	root := &Input{Kind: KindBox, On: map[string]func(*DOMEvent){
		"click": func(e *DOMEvent) { order = append(order, "root") },
	}}
	mid := &Input{Kind: KindBox}
	leaf := &Input{Kind: KindBox, On: map[string]func(*DOMEvent){
		"click": func(e *DOMEvent) { order = append(order, "leaf") },
	}}

	// When
	DispatchDOM([]*Input{root, mid, leaf}, &DOMEvent{Type: "click"})

	// Then — leaf fires before root; mid (no handler) is skipped
	if len(order) != 2 || order[0] != "leaf" || order[1] != "root" {
		t.Errorf("dispatch order = %v, want [leaf root]", order)
	}
}

func TestDispatchDOMSetsTarget(t *testing.T) {
	// Given
	var target *Input
	root := &Input{Kind: KindBox, On: map[string]func(*DOMEvent){
		"click": func(e *DOMEvent) { target = e.Target },
	}}
	leaf := &Input{Kind: KindBox}

	// When
	DispatchDOM([]*Input{root, leaf}, &DOMEvent{Type: "click"})

	// Then — Target is the deepest node even when it has no handler
	if target != leaf {
		t.Errorf("Target = %v, want the leaf input", target)
	}
}

func TestDispatchDOMStopPropagation(t *testing.T) {
	// Given
	rootCalled := false
	root := &Input{Kind: KindBox, On: map[string]func(*DOMEvent){
		"click": func(e *DOMEvent) { rootCalled = true },
	}}
	leaf := &Input{Kind: KindBox, On: map[string]func(*DOMEvent){
		"click": func(e *DOMEvent) { e.StopPropagation() },
	}}

	// When
	DispatchDOM([]*Input{root, leaf}, &DOMEvent{Type: "click"})

	// Then
	if rootCalled {
		t.Error("root handler called after StopPropagation")
	}
}

func TestDispatchDOMEmptyPathIsNoOp(t *testing.T) {
	// When / Then — must not panic
	DispatchDOM(nil, &DOMEvent{Type: "click"})
}

func TestDOMEventPreventDefault(t *testing.T) {
	// Given
	evt := &DOMEvent{Type: "keydown"}

	// Then
	if evt.DefaultPrevented() {
		t.Error("new event should not be default-prevented")
	}

	// When
	evt.PreventDefault()

	// Then
	if !evt.DefaultPrevented() {
		t.Error("DefaultPrevented() = false after PreventDefault()")
	}
}
