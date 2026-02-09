package render

import (
	"bytes"
	"testing"
)

func TestEnterAlternateScreenWritesCorrectSequence(t *testing.T) {
	var buf bytes.Buffer
	EnterAlternateScreen(&buf)

	want := "\x1b[?1049h\x1b[?25l"
	got := buf.String()
	if got != want {
		t.Errorf("EnterAlternateScreen wrote %q, want %q", got, want)
	}
}

func TestExitAlternateScreenWritesCorrectSequence(t *testing.T) {
	var buf bytes.Buffer
	ExitAlternateScreen(&buf)

	want := "\x1b[?25h\x1b[?1049l"
	got := buf.String()
	if got != want {
		t.Errorf("ExitAlternateScreen wrote %q, want %q", got, want)
	}
}

func TestAlternateScreenWithBytesBuffer(t *testing.T) {
	var buf bytes.Buffer
	EnterAlternateScreen(&buf)
	ExitAlternateScreen(&buf)

	want := "\x1b[?1049h\x1b[?25l\x1b[?25h\x1b[?1049l"
	got := buf.String()
	if got != want {
		t.Errorf("combined output = %q, want %q", got, want)
	}
}
