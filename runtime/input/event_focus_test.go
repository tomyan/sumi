package input

import "testing"

func TestEventFocusKind(t *testing.T) {
	// Given
	evt := Event{Kind: EventFocus}

	// Then
	if evt.Kind != EventFocus {
		t.Errorf("Kind = %d, want EventFocus (%d)", evt.Kind, EventFocus)
	}
}

func TestEventBlurKind(t *testing.T) {
	// Given
	evt := Event{Kind: EventBlur}

	// Then
	if evt.Kind != EventBlur {
		t.Errorf("Kind = %d, want EventBlur (%d)", evt.Kind, EventBlur)
	}
}

func TestEventFocusBlurDistinct(t *testing.T) {
	// Then — EventFocus and EventBlur should be distinct from each other and all other kinds
	kinds := []EventKind{EventKey, EventSpecial, EventMouse, EventSignal, EventFrame, EventPaste, EventFocus, EventBlur}
	for i := 0; i < len(kinds); i++ {
		for j := i + 1; j < len(kinds); j++ {
			if kinds[i] == kinds[j] {
				t.Errorf("EventKind %d and %d should be distinct, both = %d", i, j, kinds[i])
			}
		}
	}
}
