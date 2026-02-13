package input

import (
	"strings"
	"testing"
)

func TestReadEventBracketedPaste(t *testing.T) {
	// Given — bracketed paste: ESC[200~ ... ESC[201~
	r := strings.NewReader("\x1b[200~hello world\x1b[201~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventPaste {
		t.Errorf("Kind = %d, want EventPaste (%d)", ev.Kind, EventPaste)
	}
	if ev.PasteText != "hello world" {
		t.Errorf("PasteText = %q, want %q", ev.PasteText, "hello world")
	}
}

func TestReadEventBracketedPasteEmpty(t *testing.T) {
	// Given — empty paste
	r := strings.NewReader("\x1b[200~\x1b[201~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventPaste {
		t.Errorf("Kind = %d, want EventPaste (%d)", ev.Kind, EventPaste)
	}
	if ev.PasteText != "" {
		t.Errorf("PasteText = %q, want empty string", ev.PasteText)
	}
}

func TestReadEventBracketedPasteWithSpecialChars(t *testing.T) {
	// Given — paste containing newlines and tabs
	r := strings.NewReader("\x1b[200~line1\nline2\ttab\x1b[201~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventPaste {
		t.Errorf("Kind = %d, want EventPaste (%d)", ev.Kind, EventPaste)
	}
	if ev.PasteText != "line1\nline2\ttab" {
		t.Errorf("PasteText = %q, want %q", ev.PasteText, "line1\nline2\ttab")
	}
}

func TestPasteEnableDisableSequences(t *testing.T) {
	// Then — constants should be the correct escape sequences
	if PasteEnableSeq != "\x1b[?2004h" {
		t.Errorf("PasteEnableSeq = %q, want %q", PasteEnableSeq, "\x1b[?2004h")
	}
	if PasteDisableSeq != "\x1b[?2004l" {
		t.Errorf("PasteDisableSeq = %q, want %q", PasteDisableSeq, "\x1b[?2004l")
	}
}
