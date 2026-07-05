package layout

import "testing"

func TestCollectFocusablesReturnsTreeOrder(t *testing.T) {
	// Given — focusables at different depths, with a nil child (display:none)
	first := &Input{Kind: KindBox, Focusable: true}
	second := &Input{Kind: KindBox, Focusable: true}
	root := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox, Children: []*Input{first}},
			nil,
			{Kind: KindBox},
			second,
		},
	}

	// When
	got := CollectFocusables(root)

	// Then
	if len(got) != 2 {
		t.Fatalf("CollectFocusables returned %d inputs, want 2", len(got))
	}
	if got[0] != first || got[1] != second {
		t.Errorf("CollectFocusables order wrong: got [%p %p], want [%p %p]", got[0], got[1], first, second)
	}
}

func TestFocusablePathReturnsAncestryOfIndexedFocusable(t *testing.T) {
	// Given — two focusables at different depths
	first := &Input{Kind: KindBox, Focusable: true}
	second := &Input{Kind: KindBox, Focusable: true}
	mid := &Input{Kind: KindBox, Children: []*Input{first}}
	root := &Input{Kind: KindBox, Children: []*Input{mid, second}}

	// When / Then — path runs root → focusable
	path := FocusablePath(root, 0)
	if len(path) != 3 || path[0] != root || path[1] != mid || path[2] != first {
		t.Fatalf("FocusablePath(root, 0) = %v, want [root mid first]", path)
	}
	path = FocusablePath(root, 1)
	if len(path) != 2 || path[0] != root || path[1] != second {
		t.Fatalf("FocusablePath(root, 1) = %v, want [root second]", path)
	}
}

func TestFocusablePathOutOfRangeReturnsNil(t *testing.T) {
	// Given
	root := &Input{Kind: KindBox, Children: []*Input{{Kind: KindBox, Focusable: true}}}

	// When / Then
	if path := FocusablePath(root, 5); path != nil {
		t.Errorf("FocusablePath out of range = %v, want nil", path)
	}
	if path := FocusablePath(root, -1); path != nil {
		t.Errorf("FocusablePath(-1) = %v, want nil", path)
	}
}

func TestButtonElementIsFocusableByDefault(t *testing.T) {
	// Given — a button element with no focusable attribute
	button := &Input{Kind: KindText, Tag: "button", Content: "Save"}
	root := &Input{Kind: KindBox, Children: []*Input{button}}

	// When
	got := CollectFocusables(root)

	// Then
	if len(got) != 1 || got[0] != button {
		t.Errorf("CollectFocusables = %v, want the button element", got)
	}
	if path := FocusablePath(root, 0); len(path) != 2 || path[1] != button {
		t.Errorf("FocusablePath = %v, want [root button]", path)
	}
}

func TestDisabledButtonSkippedByFocusTraversal(t *testing.T) {
	// Given
	disabled := &Input{Kind: KindText, Tag: "button", Content: "Off",
		Attrs: map[string]string{"disabled": "true"}}
	enabled := &Input{Kind: KindText, Tag: "button", Content: "On"}
	root := &Input{Kind: KindBox, Children: []*Input{disabled, enabled}}

	// When
	got := CollectFocusables(root)

	// Then
	if len(got) != 1 || got[0] != enabled {
		t.Errorf("CollectFocusables = %v, want only the enabled button", got)
	}
}

func TestCollectFocusablesWithoutFocusables(t *testing.T) {
	// Given
	root := &Input{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "hi"}}}

	// When / Then
	if got := CollectFocusables(root); len(got) != 0 {
		t.Errorf("expected no focusables, got %d", len(got))
	}
	if got := CollectFocusables(nil); len(got) != 0 {
		t.Errorf("expected no focusables for nil root, got %d", len(got))
	}
}

func TestFocusCycleForward(t *testing.T) {
	// Given
	tests := []struct {
		name   string
		focus  int
		count  int
		expect int
	}{
		{"0 of 3 → 1", 0, 3, 1},
		{"1 of 3 → 2", 1, 3, 2},
		{"2 of 3 → 0 (wrap)", 2, 3, 0},
		{"0 of 1 → 0 (single)", 0, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			got := CycleFocus(tt.focus, tt.count)

			// Then
			if got != tt.expect {
				t.Errorf("CycleFocus(%d, %d) = %d, want %d", tt.focus, tt.count, got, tt.expect)
			}
		})
	}
}

func TestFocusCycleBackward(t *testing.T) {
	// Given
	tests := []struct {
		name   string
		focus  int
		count  int
		expect int
	}{
		{"1 of 3 → 0", 1, 3, 0},
		{"0 of 3 → 2 (wrap)", 0, 3, 2},
		{"2 of 3 → 1", 2, 3, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			got := CycleFocusBackward(tt.focus, tt.count)

			// Then
			if got != tt.expect {
				t.Errorf("CycleFocusBackward(%d, %d) = %d, want %d", tt.focus, tt.count, got, tt.expect)
			}
		})
	}
}
