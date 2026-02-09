package render

import (
	"bytes"
	"testing"
)

func TestEnterAlternateScreenWritesCorrectSequence(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	EnterAlternateScreen(&buf)

	// Then
	want := "\x1b[?1049h\x1b[?25l"
	got := buf.String()
	if got != want {
		t.Errorf("EnterAlternateScreen wrote %q, want %q", got, want)
	}
}

func TestExitAlternateScreenWritesCorrectSequence(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	ExitAlternateScreen(&buf)

	// Then
	want := "\x1b[?25h\x1b[?1049l"
	got := buf.String()
	if got != want {
		t.Errorf("ExitAlternateScreen wrote %q, want %q", got, want)
	}
}

func TestAlternateScreenWithBytesBuffer(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	EnterAlternateScreen(&buf)
	ExitAlternateScreen(&buf)

	// Then
	want := "\x1b[?1049h\x1b[?25l\x1b[?25h\x1b[?1049l"
	got := buf.String()
	if got != want {
		t.Errorf("combined output = %q, want %q", got, want)
	}
}
