package input

import (
	"strings"
	"testing"
)

func TestCtrlRight(t *testing.T) {
	// Given — ESC[1;5C = Ctrl+Right
	r := strings.NewReader("\x1b[1;5C")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyRight {
		t.Errorf("Special = %q, want %q", ev.Special, KeyRight)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
	if ev.Shift {
		t.Error("expected Shift = false")
	}
}

func TestShiftLeft(t *testing.T) {
	// Given — ESC[1;2D = Shift+Left
	r := strings.NewReader("\x1b[1;2D")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyLeft {
		t.Errorf("Special = %q, want %q", ev.Special, KeyLeft)
	}
	if !ev.Shift {
		t.Error("expected Shift = true")
	}
	if ev.Ctrl {
		t.Error("expected Ctrl = false")
	}
}

func TestShiftHome(t *testing.T) {
	// Given — ESC[1;2H = Shift+Home
	r := strings.NewReader("\x1b[1;2H")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyHome {
		t.Errorf("Special = %q, want %q", ev.Special, KeyHome)
	}
	if !ev.Shift {
		t.Error("expected Shift = true")
	}
}

func TestCtrlShiftRight(t *testing.T) {
	// Given — ESC[1;6C = Ctrl+Shift+Right (modifier 6: bits shift+ctrl)
	r := strings.NewReader("\x1b[1;6C")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyRight {
		t.Errorf("Special = %q, want %q", ev.Special, KeyRight)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
	if !ev.Shift {
		t.Error("expected Shift = true")
	}
}

func TestCtrlDelete(t *testing.T) {
	// Given — ESC[3;5~ = Ctrl+Delete
	r := strings.NewReader("\x1b[3;5~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyDelete {
		t.Errorf("Special = %q, want %q", ev.Special, KeyDelete)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
}

func TestAltRight(t *testing.T) {
	// Given — ESC[1;3C = Alt+Right (modifier 3: bit alt)
	r := strings.NewReader("\x1b[1;3C")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Special != KeyRight {
		t.Errorf("Special = %q, want %q", ev.Special, KeyRight)
	}
	if !ev.Alt {
		t.Error("expected Alt = true")
	}
	if ev.Ctrl {
		t.Error("expected Ctrl = false")
	}
}
