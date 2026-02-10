package term

import (
	"testing"
)

func TestGetSizeReturnsPositiveDimensions(t *testing.T) {
	// Given a file descriptor (stdin)
	// When we get the terminal size
	w, h := GetSize(0)

	// Then dimensions should be positive (even if fallback)
	if w <= 0 {
		t.Errorf("width = %d, want > 0", w)
	}
	if h <= 0 {
		t.Errorf("height = %d, want > 0", h)
	}
}

func TestGetSizeInvalidFdReturnsFallback(t *testing.T) {
	// Given an invalid file descriptor
	// When we get the terminal size
	w, h := GetSize(-1)

	// Then it should return the fallback (80x24)
	if w != 80 {
		t.Errorf("width = %d, want 80 (fallback)", w)
	}
	if h != 24 {
		t.Errorf("height = %d, want 24 (fallback)", h)
	}
}
