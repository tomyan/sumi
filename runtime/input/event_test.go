package input

import (
	"strings"
	"testing"
)

func TestReadEventPlainRune(t *testing.T) {
	// Given
	r := strings.NewReader("a")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 'a' {
		t.Errorf("Rune = %c, want 'a'", ev.Rune)
	}
}

func TestReadEventArrowUp(t *testing.T) {
	// Given — \x1b[A is arrow up
	r := strings.NewReader("\x1b[A")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyUp {
		t.Errorf("Special = %q, want %q", ev.Special, KeyUp)
	}
}

func TestReadEventArrowDown(t *testing.T) {
	// Given
	r := strings.NewReader("\x1b[B")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyDown {
		t.Errorf("Special = %q, want %q", ev.Special, KeyDown)
	}
}

func TestReadEventArrowRight(t *testing.T) {
	// Given
	r := strings.NewReader("\x1b[C")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyRight {
		t.Errorf("Special = %q, want %q", ev.Special, KeyRight)
	}
}

func TestReadEventArrowLeft(t *testing.T) {
	// Given
	r := strings.NewReader("\x1b[D")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyLeft {
		t.Errorf("Special = %q, want %q", ev.Special, KeyLeft)
	}
}

func TestReadEventPageUp(t *testing.T) {
	// Given — \x1b[5~ is page up
	r := strings.NewReader("\x1b[5~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyPgUp {
		t.Errorf("Special = %q, want %q", ev.Special, KeyPgUp)
	}
}

func TestReadEventPageDown(t *testing.T) {
	// Given
	r := strings.NewReader("\x1b[6~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyPgDn {
		t.Errorf("Special = %q, want %q", ev.Special, KeyPgDn)
	}
}

func TestReadEventHome(t *testing.T) {
	// Given
	r := strings.NewReader("\x1b[H")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyHome {
		t.Errorf("Special = %q, want %q", ev.Special, KeyHome)
	}
}

func TestReadEventEnd(t *testing.T) {
	// Given
	r := strings.NewReader("\x1b[F")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyEnd {
		t.Errorf("Special = %q, want %q", ev.Special, KeyEnd)
	}
}

func TestReadEventEscapeAlone(t *testing.T) {
	// Given — bare escape with no following [ (just ESC key)
	r := strings.NewReader("\x1b")

	// When
	ev, err := ReadEvent(r)

	// Then — should return escape as a rune
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 0x1b {
		t.Errorf("Rune = %d, want 0x1b (ESC)", ev.Rune)
	}
}

func TestReadEventTab(t *testing.T) {
	// Given
	r := strings.NewReader("\t")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyTab {
		t.Errorf("Special = %q, want %q", ev.Special, KeyTab)
	}
}

func TestReadEventShiftTab(t *testing.T) {
	// Given — \x1b[Z is shift-tab
	r := strings.NewReader("\x1b[Z")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyShiftTab {
		t.Errorf("Special = %q, want %q", ev.Special, KeyShiftTab)
	}
}

func TestReadEventUTF8TwoByte(t *testing.T) {
	// Given — "é" is a 2-byte UTF-8 character (0xC3 0xA9)
	r := strings.NewReader("é")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != 'é' {
		t.Errorf("Rune = %U, want %U ('é')", ev.Rune, 'é')
	}
}

func TestReadEventUTF8ThreeByte(t *testing.T) {
	// Given — "日" is a 3-byte UTF-8 character (0xE6 0x97 0xA5)
	r := strings.NewReader("日")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != '日' {
		t.Errorf("Rune = %U, want %U ('日')", ev.Rune, '日')
	}
}

func TestReadEventUTF8FourByte(t *testing.T) {
	// Given — "🎉" is a 4-byte UTF-8 character (0xF0 0x9F 0x8E 0x89)
	r := strings.NewReader("🎉")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if ev.Rune != '🎉' {
		t.Errorf("Rune = %U, want %U ('🎉')", ev.Rune, '🎉')
	}
}
