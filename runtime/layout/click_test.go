package layout

import "testing"

func TestFindClickHandlerHitsDeepestBox(t *testing.T) {
	// Given a parent box with an onclick and a child with its own onclick
	parentClicked := false
	childClicked := false

	parent := &Input{
		Kind:    KindBox,
		OnClick: func() { parentClicked = true },
		Children: []*Input{
			{
				Kind:    KindBox,
				OnClick: func() { childClicked = true },
			},
		},
	}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 2, Y: 2, Width: 10, Height: 5},
		},
	}

	// When clicking inside the child box
	handler := FindClickHandler(parent, parentBox, 5, 3)

	// Then the deepest handler (child) is returned
	if handler == nil {
		t.Fatal("expected a handler, got nil")
	}
	handler()
	if !childClicked {
		t.Error("expected child handler to be called")
	}
	if parentClicked {
		t.Error("expected parent handler NOT to be called")
	}
}

func TestFindClickHandlerFallsBackToParent(t *testing.T) {
	// Given a parent with onclick and a child without
	parentClicked := false

	parent := &Input{
		Kind:    KindBox,
		OnClick: func() { parentClicked = true },
		Children: []*Input{
			{Kind: KindBox},
		},
	}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 2, Y: 2, Width: 10, Height: 5},
		},
	}

	// When clicking inside the child box (which has no handler)
	handler := FindClickHandler(parent, parentBox, 5, 3)

	// Then the parent handler is returned
	if handler == nil {
		t.Fatal("expected a handler, got nil")
	}
	handler()
	if !parentClicked {
		t.Error("expected parent handler to be called")
	}
}

func TestFindClickHandlerReturnsNilOutsideBounds(t *testing.T) {
	// Given a box with onclick
	parent := &Input{
		Kind:    KindBox,
		OnClick: func() {},
	}
	parentBox := &Box{
		X: 5, Y: 5, Width: 10, Height: 5,
	}

	// When clicking outside the box
	handler := FindClickHandler(parent, parentBox, 0, 0)

	// Then no handler is returned
	if handler != nil {
		t.Error("expected nil handler for click outside bounds")
	}
}

func TestFindClickHandlerReturnsNilNoHandlers(t *testing.T) {
	// Given boxes without onclick
	parent := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox},
		},
	}
	parentBox := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 5},
		},
	}

	// When clicking inside
	handler := FindClickHandler(parent, parentBox, 5, 3)

	// Then nil is returned
	if handler != nil {
		t.Error("expected nil handler when no onclick handlers exist")
	}
}

func TestHasClickHandlersDetectsNestedHandler(t *testing.T) {
	// Given a tree with a nested onclick
	tree := &Input{
		Kind: KindBox,
		Children: []*Input{
			{
				Kind:    KindBox,
				OnClick: func() {},
			},
		},
	}

	// Then
	if !HasClickHandlers(tree) {
		t.Error("expected HasClickHandlers to return true")
	}
}

func TestHasClickHandlersReturnsFalseWhenNone(t *testing.T) {
	// Given a tree with no onclick
	tree := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox},
		},
	}

	// Then
	if HasClickHandlers(tree) {
		t.Error("expected HasClickHandlers to return false")
	}
}

func TestFindClickHandlerSkipsNilChildren(t *testing.T) {
	// Given a parent with nil children (display:none placeholders)
	clicked := false
	parent := &Input{
		Kind:    KindBox,
		OnClick: func() { clicked = true },
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

	// When clicking inside
	handler := FindClickHandler(parent, parentBox, 5, 3)

	// Then parent handler is returned (no crash on nil)
	if handler == nil {
		t.Fatal("expected handler, got nil")
	}
	handler()
	if !clicked {
		t.Error("expected parent handler to be called")
	}
}
