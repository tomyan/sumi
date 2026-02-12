package input

import (
	"strings"
	"syscall"
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

	// Then — should return escape as a special key
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

func TestEventSignalKind(t *testing.T) {
	// Given — an event with the signal kind
	evt := Event{Kind: EventSignal, Signal: syscall.SIGINT}

	// Then
	if evt.Kind != EventSignal {
		t.Errorf("Kind = %d, want EventSignal (%d)", evt.Kind, EventSignal)
	}
	if evt.Signal != syscall.SIGINT {
		t.Errorf("Signal = %v, want SIGINT", evt.Signal)
	}
}

func TestEventSignalPreservesIdentity(t *testing.T) {
	// Given — different signals should be distinguishable
	sigint := Event{Kind: EventSignal, Signal: syscall.SIGINT}
	sigterm := Event{Kind: EventSignal, Signal: syscall.SIGTERM}

	// Then
	if sigint.Signal == sigterm.Signal {
		t.Error("SIGINT and SIGTERM should be different signals")
	}
}

func TestReadEventEnter(t *testing.T) {
	// Given — \r (carriage return) should be Enter
	r := strings.NewReader("\r")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyEnter {
		t.Errorf("Special = %q, want %q", ev.Special, KeyEnter)
	}
}

func TestReadEventEnterNewline(t *testing.T) {
	// Given — \n (newline) should also be Enter
	r := strings.NewReader("\n")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyEnter {
		t.Errorf("Special = %q, want %q", ev.Special, KeyEnter)
	}
}

func TestReadEventBackspace(t *testing.T) {
	// Given — 127 (DEL) should be Backspace
	r := strings.NewReader("\x7f")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyBackspace {
		t.Errorf("Special = %q, want %q", ev.Special, KeyBackspace)
	}
}

func TestReadEventBackspaceCtrlH(t *testing.T) {
	// Given — 8 (Ctrl+H / BS) should also be Backspace
	r := strings.NewReader("\x08")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyBackspace {
		t.Errorf("Special = %q, want %q", ev.Special, KeyBackspace)
	}
}

func TestReadEventDelete(t *testing.T) {
	// Given — ESC[3~ is Delete
	r := strings.NewReader("\x1b[3~")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Special != KeyDelete {
		t.Errorf("Special = %q, want %q", ev.Special, KeyDelete)
	}
}

func TestReadEventEscapeAsSpecial(t *testing.T) {
	// Given — bare ESC should be the Escape special key
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

func TestReadEventCtrlC(t *testing.T) {
	// Given — Ctrl+C is byte 3
	r := strings.NewReader("\x03")

	// When
	ev, err := ReadEvent(r)

	// Then — should be EventKey with Ctrl flag and rune 'c'
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", ev.Kind)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
	if ev.Rune != 'c' {
		t.Errorf("Rune = %c (%d), want 'c'", ev.Rune, ev.Rune)
	}
}

func TestReadEventCtrlA(t *testing.T) {
	// Given — Ctrl+A is byte 1
	r := strings.NewReader("\x01")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
	if ev.Rune != 'a' {
		t.Errorf("Rune = %c (%d), want 'a'", ev.Rune, ev.Rune)
	}
}

func TestReadEventCtrlZ(t *testing.T) {
	// Given — Ctrl+Z is byte 26
	r := strings.NewReader("\x1a")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl = true")
	}
	if ev.Rune != 'z' {
		t.Errorf("Rune = %c (%d), want 'z'", ev.Rune, ev.Rune)
	}
}

func TestReadEventTabNotCtrl(t *testing.T) {
	// Given — Tab (0x09/Ctrl+I) should remain KeyTab, not Ctrl+i
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
	if ev.Ctrl {
		t.Error("Tab should not have Ctrl flag")
	}
}

func TestEventFrameKind(t *testing.T) {
	// Given — EventFrame is a distinct event kind for animation ticks
	evt := Event{Kind: EventFrame}

	// Then
	if evt.Kind != EventFrame {
		t.Errorf("Kind = %d, want EventFrame (%d)", evt.Kind, EventFrame)
	}
	// EventFrame should be distinct from all other kinds
	if EventFrame == EventKey || EventFrame == EventSpecial || EventFrame == EventMouse || EventFrame == EventSignal {
		t.Error("EventFrame should be a unique kind value")
	}
}

func TestReadEventEnterNotCtrl(t *testing.T) {
	// Given — Enter (0x0d/Ctrl+M) should remain KeyEnter, not Ctrl+m
	r := strings.NewReader("\r")

	// When
	ev, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Kind != EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", ev.Kind)
	}
	if ev.Ctrl {
		t.Error("Enter should not have Ctrl flag")
	}
}
