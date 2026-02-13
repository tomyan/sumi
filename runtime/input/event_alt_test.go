package input

import (
	"strings"
	"testing"
)

func TestAltF(t *testing.T) {
	// Given — ESC followed by 'f' is Alt+F
	r := strings.NewReader("\x1bf")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 'f' {
		t.Errorf("Rune = %c, want 'f'", ev.Rune)
	}
	if !ev.Alt {
		t.Error("expected Alt = true")
	}
	if ev.Ctrl {
		t.Error("expected Ctrl = false")
	}
}

func TestAltB(t *testing.T) {
	// Given — ESC followed by 'b' is Alt+B
	r := strings.NewReader("\x1bb")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 'b' {
		t.Errorf("Rune = %c, want 'b'", ev.Rune)
	}
	if !ev.Alt {
		t.Error("expected Alt = true")
	}
}

func TestAltD(t *testing.T) {
	// Given — ESC followed by 'd' is Alt+D
	r := strings.NewReader("\x1bd")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 'd' {
		t.Errorf("Rune = %c, want 'd'", ev.Rune)
	}
	if !ev.Alt {
		t.Error("expected Alt = true")
	}
}

func TestAltUppercaseF(t *testing.T) {
	// Given — ESC followed by 'F' is Alt+Shift+F
	r := strings.NewReader("\x1bF")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 'F' {
		t.Errorf("Rune = %c, want 'F'", ev.Rune)
	}
	if !ev.Alt {
		t.Error("expected Alt = true")
	}
}

func TestCtrlSlash(t *testing.T) {
	// Given — Ctrl+/ sends byte 0x1F in terminals
	r := strings.NewReader("\x1f")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != '/' {
		t.Errorf("Rune = %c (%d), want '/'", ev.Rune, ev.Rune)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
}

func TestBareEscapeStillWorks(t *testing.T) {
	// Given — bare ESC with no following bytes should still be Escape
	r := strings.NewReader("\x1b")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyEscape {
		t.Errorf("Special = %q, want %q", ev.Special, KeyEscape)
	}
}
