package input

import (
	"strings"
	"testing"
)

func TestReadEventMousePress(t *testing.T) {
	// Given — SGR mouse press: ESC[<0;10;5M (left button press at col=10, row=5)
	r := strings.NewReader("\x1b[<0;10;5M")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventMouse {
		t.Fatalf("Kind = %d, want EventMouse", ev.Kind)
	}
	if ev.Mouse.Action != MousePress {
		t.Errorf("Action = %d, want MousePress", ev.Mouse.Action)
	}
	if ev.Mouse.Button != ButtonLeft {
		t.Errorf("Button = %d, want ButtonLeft", ev.Mouse.Button)
	}
	if ev.Mouse.X != 9 { // 0-indexed from 1-indexed terminal coords
		t.Errorf("X = %d, want 9", ev.Mouse.X)
	}
	if ev.Mouse.Y != 4 { // 0-indexed
		t.Errorf("Y = %d, want 4", ev.Mouse.Y)
	}
}

func TestReadEventMouseRelease(t *testing.T) {
	// Given — SGR mouse release: ESC[<0;10;5m (lowercase m = release)
	r := strings.NewReader("\x1b[<0;10;5m")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Mouse.Action != MouseRelease {
		t.Errorf("Action = %d, want MouseRelease", ev.Mouse.Action)
	}
}

func TestReadEventScrollWheelUp(t *testing.T) {
	// Given — scroll up: button code 64
	r := strings.NewReader("\x1b[<64;5;3M")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventMouse {
		t.Fatalf("Kind = %d, want EventMouse", ev.Kind)
	}
	if ev.Mouse.Action != MouseScroll {
		t.Errorf("Action = %d, want MouseScroll", ev.Mouse.Action)
	}
	if ev.Mouse.Button != ScrollUp {
		t.Errorf("Button = %d, want ScrollUp", ev.Mouse.Button)
	}
}

func TestReadEventScrollWheelDown(t *testing.T) {
	// Given — scroll down: button code 65
	r := strings.NewReader("\x1b[<65;5;3M")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Mouse.Button != ScrollDown {
		t.Errorf("Button = %d, want ScrollDown", ev.Mouse.Button)
	}
}

func TestReadEventRightClick(t *testing.T) {
	// Given — right button press: button code 2
	r := strings.NewReader("\x1b[<2;5;3M")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Mouse.Button != ButtonRight {
		t.Errorf("Button = %d, want ButtonRight", ev.Mouse.Button)
	}
}

func TestReadEventMiddleClick(t *testing.T) {
	// Given — middle button press: button code 1
	r := strings.NewReader("\x1b[<1;5;3M")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Mouse.Button != ButtonMiddle {
		t.Errorf("Button = %d, want ButtonMiddle", ev.Mouse.Button)
	}
}

func TestReadEventMouseMotion(t *testing.T) {
	// Given — motion with left button held: button code 32 (0 + 32 motion flag)
	r := strings.NewReader("\x1b[<32;15;8M")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Mouse.Action != MouseMotion {
		t.Errorf("Action = %d, want MouseMotion", ev.Mouse.Action)
	}
}

func TestMouseEnableDisableSequences(t *testing.T) {
	// Then — verify the escape sequences are correct constants
	if MouseEnableSeq != "\x1b[?1006h\x1b[?1003h" {
		t.Errorf("MouseEnableSeq = %q, unexpected", MouseEnableSeq)
	}
	if MouseDisableSeq != "\x1b[?1003l\x1b[?1006l" {
		t.Errorf("MouseDisableSeq = %q, unexpected", MouseDisableSeq)
	}
}
