package layout

import "testing"

func TestCollectScrollStatesFindsScrollable(t *testing.T) {
	// Given a tree with one scrollable input
	scroll := &ScrollState{}
	root := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox, Scroll: scroll, Overflow: "auto"},
		},
	}

	// When
	states := CollectScrollStates(root)

	// Then
	if len(states) != 1 {
		t.Fatalf("got %d states, want 1", len(states))
	}
	if states[0] != scroll {
		t.Error("expected the scroll state pointer to match")
	}
}

func TestCollectScrollStatesDepthFirstOrder(t *testing.T) {
	// Given two scrollable inputs in depth-first order
	s0 := &ScrollState{}
	s1 := &ScrollState{}
	root := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox, Scroll: s0, Overflow: "auto"},
			{Kind: KindBox, Scroll: s1, Overflow: "auto"},
		},
	}

	// When
	states := CollectScrollStates(root)

	// Then
	if len(states) != 2 {
		t.Fatalf("got %d states, want 2", len(states))
	}
	if states[0] != s0 || states[1] != s1 {
		t.Error("states not in depth-first order")
	}
}

func TestCollectScrollStatesSkipsNonScrollable(t *testing.T) {
	// Given a tree with no scrollable inputs
	root := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox},
			{Kind: KindText},
		},
	}

	// When
	states := CollectScrollStates(root)

	// Then
	if len(states) != 0 {
		t.Errorf("got %d states, want 0", len(states))
	}
}

func TestCollectScrollStatesSkipsNil(t *testing.T) {
	// Given a tree with nil children
	scroll := &ScrollState{}
	root := &Input{
		Kind: KindBox,
		Children: []*Input{
			nil,
			{Kind: KindBox, Scroll: scroll, Overflow: "auto"},
		},
	}

	// When
	states := CollectScrollStates(root)

	// Then
	if len(states) != 1 {
		t.Fatalf("got %d states, want 1", len(states))
	}
}
