package layout

import "testing"

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
