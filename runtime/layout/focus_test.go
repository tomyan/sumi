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
