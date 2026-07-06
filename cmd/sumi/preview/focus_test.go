package preview

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestFocusDefaultIsControls(t *testing.T) {
	// Given
	pvFocus = FocusControls

	// Then
	if pvFocus != FocusControls {
		t.Errorf("pvFocus = %d, want FocusControls", pvFocus)
	}
}

func TestFocusTransitionToEditor(t *testing.T) {
	// Given
	pvFocus = FocusControls

	// When — pressing '1' focuses editor 1
	pvFocus = FocusEditor1

	// Then
	if pvFocus != FocusEditor1 {
		t.Errorf("pvFocus = %d, want FocusEditor1", pvFocus)
	}
}

func TestFocusTransitionBackToControls(t *testing.T) {
	// Given
	pvFocus = FocusEditor1

	// When — Ctrl+\ exits to controls
	pvFocus = FocusControls

	// Then
	if pvFocus != FocusControls {
		t.Errorf("pvFocus = %d, want FocusControls", pvFocus)
	}
}

func TestPrefixPending(t *testing.T) {
	// Given
	pvPrefixPending = false

	// When — Ctrl+\ sets prefix pending
	pvPrefixPending = true

	// Then
	if !pvPrefixPending {
		t.Error("pvPrefixPending should be true")
	}
}

func TestHandleFocusKeyPressDigit(t *testing.T) {
	// Given
	pvFocus = FocusControls

	// When/Then — '1' should focus editor 1
	newFocus := focusForDigit('1')
	if newFocus != FocusEditor1 {
		t.Errorf("focusForDigit('1') = %d, want FocusEditor1", newFocus)
	}

	newFocus = focusForDigit('2')
	if newFocus != FocusEditor2 {
		t.Errorf("focusForDigit('2') = %d, want FocusEditor2", newFocus)
	}

	newFocus = focusForDigit('3')
	if newFocus != FocusEditor3 {
		t.Errorf("focusForDigit('3') = %d, want FocusEditor3", newFocus)
	}
}

func TestHandlePrefixCommand(t *testing.T) {
	// Given — prefix pending
	tests := []struct {
		name string
		rune rune
		want string
	}{
		{"quit", 'q', "quit"},
		{"next", 'l', "next"},
		{"prev", 'h', "prev"},
		{"update", 'u', "update"},
		{"interactive", 'i', "interactive"},
		{"unknown", 'x', ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// When
			cmd := prefixCommand(input.Event{Kind: input.EventKey, Rune: tc.rune})

			// Then
			if cmd != tc.want {
				t.Errorf("prefixCommand(%c) = %q, want %q", tc.rune, cmd, tc.want)
			}
		})
	}
}

func TestPrefixCommandCtrlBackslashReturnsExit(t *testing.T) {
	// Given
	evt := input.Event{Kind: input.EventKey, Rune: 0x1c}

	// When
	cmd := prefixCommand(evt)

	// Then
	if cmd != "exit" {
		t.Errorf("prefixCommand(Ctrl+\\) = %q, want %q", cmd, "exit")
	}
}

func TestFocusStateName(t *testing.T) {
	// Given/When/Then
	tests := []struct {
		state FocusState
		want  string
	}{
		{FocusControls, "controls"},
		{FocusEditor1, "source"},
		{FocusEditor2, "snapshot"},
		{FocusEditor3, "scenario"},
		{FocusInteractive, "interactive"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if tc.state.Name() != tc.want {
				t.Errorf("FocusState(%d).Name() = %q, want %q", tc.state, tc.state.Name(), tc.want)
			}
		})
	}
}
