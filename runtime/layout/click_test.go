package layout

import "testing"

func TestHitTestPathReturnsRootToDeepest(t *testing.T) {
	// Given a parent box with a nested child
	child := &Input{Kind: KindBox}
	parent := &Input{Kind: KindBox, Children: []*Input{child}}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 2, Y: 2, Width: 10, Height: 5},
		},
	}

	// When hitting inside the child box
	path := HitTestPath(parent, parentBox, 5, 3)

	// Then the path runs root → deepest
	if len(path) != 2 || path[0] != parent || path[1] != child {
		t.Fatalf("path = %v, want [parent child]", path)
	}
}

func TestHitTestPathParentOnlyOutsideChild(t *testing.T) {
	// Given
	child := &Input{Kind: KindBox}
	parent := &Input{Kind: KindBox, Children: []*Input{child}}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 2, Y: 2, Width: 10, Height: 5},
		},
	}

	// When hitting inside the parent but outside the child
	path := HitTestPath(parent, parentBox, 15, 9)

	// Then only the parent is on the path
	if len(path) != 1 || path[0] != parent {
		t.Fatalf("path = %v, want [parent]", path)
	}
}

func TestHitTestPathEmptyOutsideBounds(t *testing.T) {
	// Given
	parent := &Input{Kind: KindBox}
	parentBox := &Box{X: 5, Y: 5, Width: 10, Height: 5}

	// When / Then
	if path := HitTestPath(parent, parentBox, 0, 0); path != nil {
		t.Errorf("path = %v, want nil for miss", path)
	}
}

func TestHitTestPathBubbleFallsBackToParentHandler(t *testing.T) {
	// Given a parent with a click handler and a child without
	parentClicked := false
	child := &Input{Kind: KindBox}
	parent := &Input{
		Kind: KindBox,
		On: map[string]func(*DOMEvent){
			"click": func(e *DOMEvent) { parentClicked = true },
		},
		Children: []*Input{child},
	}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 2, Y: 2, Width: 10, Height: 5},
		},
	}

	// When clicking inside the child and dispatching
	path := HitTestPath(parent, parentBox, 5, 3)
	DispatchDOM(path, &DOMEvent{Type: "click"})

	// Then the click bubbles to the parent handler
	if !parentClicked {
		t.Error("expected parent handler to fire via bubbling")
	}
}

func TestHitTestPathReachesFixedChildInZeroSizedParent(t *testing.T) {
	// Given a zero-sized parent with a fixed-position child covering the viewport
	child := &Input{Kind: KindBox, Position: "fixed"}
	parent := &Input{Kind: KindBox, Children: []*Input{child}}
	parentBox := &Box{
		X: 0, Y: 0, Width: 0, Height: 0, // zero-sized (fixed child removed from flow)
		Children: []*Box{
			{X: 0, Y: 0, Width: 80, Height: 24},
		},
	}

	// When clicking anywhere in the viewport
	path := HitTestPath(parent, parentBox, 40, 12)

	// Then the ancestry path still includes the parent
	if len(path) != 2 || path[0] != parent || path[1] != child {
		t.Fatalf("path = %v, want [parent child]", path)
	}
}

func TestHitTestPathSkipsNilChildren(t *testing.T) {
	// Given a parent with nil children (display:none placeholders)
	parent := &Input{
		Kind: KindBox,
		Children: []*Input{
			nil,
			{Kind: KindBox},
		},
	}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			nil,
			{X: 0, Y: 0, Width: 10, Height: 5},
		},
	}

	// When / Then — no crash, child on path
	path := HitTestPath(parent, parentBox, 5, 3)
	if len(path) != 2 {
		t.Fatalf("path length = %d, want 2", len(path))
	}
}

func TestHasClickHandlersDetectsNestedHandler(t *testing.T) {
	// Given a tree with a nested click handler
	tree := &Input{
		Kind: KindBox,
		Children: []*Input{
			{
				Kind: KindBox,
				On: map[string]func(*DOMEvent){
					"click": func(e *DOMEvent) {},
				},
			},
		},
	}

	// Then
	if !HasClickHandlers(tree) {
		t.Error("expected HasClickHandlers to return true")
	}
}

func TestHasClickHandlersReturnsFalseWhenNone(t *testing.T) {
	// Given a tree with handlers for other event types only
	tree := &Input{
		Kind: KindBox,
		Children: []*Input{
			{
				Kind: KindBox,
				On: map[string]func(*DOMEvent){
					"keydown": func(e *DOMEvent) {},
				},
			},
		},
	}

	// Then
	if HasClickHandlers(tree) {
		t.Error("expected HasClickHandlers to return false")
	}
}
