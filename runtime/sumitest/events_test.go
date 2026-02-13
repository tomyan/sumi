package sumitest

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestKeyEvent(t *testing.T) {
	evt := KeyEvent('a')
	if evt.Kind != input.EventKey {
		t.Errorf("Kind = %d, want EventKey", evt.Kind)
	}
	if evt.Rune != 'a' {
		t.Errorf("Rune = %c, want 'a'", evt.Rune)
	}
}

func TestCtrlEvent(t *testing.T) {
	evt := CtrlEvent('c')
	if evt.Kind != input.EventKey {
		t.Errorf("Kind = %d, want EventKey", evt.Kind)
	}
	if evt.Rune != 'c' {
		t.Errorf("Rune = %c, want 'c'", evt.Rune)
	}
	if !evt.Ctrl {
		t.Error("Ctrl not set")
	}
}

func TestSpecialEvent(t *testing.T) {
	evt := SpecialEvent(input.KeyUp)
	if evt.Kind != input.EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", evt.Kind)
	}
	if evt.Special != input.KeyUp {
		t.Errorf("Special = %q, want %q", evt.Special, input.KeyUp)
	}
}

func TestPasteEvent(t *testing.T) {
	evt := PasteEvent("hello world")
	if evt.Kind != input.EventPaste {
		t.Errorf("Kind = %d, want EventPaste", evt.Kind)
	}
	if evt.PasteText != "hello world" {
		t.Errorf("PasteText = %q, want %q", evt.PasteText, "hello world")
	}
}

func TestEnterEvent(t *testing.T) {
	evt := EnterEvent()
	if evt.Kind != input.EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", evt.Kind)
	}
	if evt.Special != input.KeyEnter {
		t.Errorf("Special = %q, want %q", evt.Special, input.KeyEnter)
	}
}

func TestEscapeEvent(t *testing.T) {
	evt := EscapeEvent()
	if evt.Kind != input.EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", evt.Kind)
	}
	if evt.Special != input.KeyEscape {
		t.Errorf("Special = %q, want %q", evt.Special, input.KeyEscape)
	}
}

func TestBackspaceEvent(t *testing.T) {
	evt := BackspaceEvent()
	if evt.Kind != input.EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", evt.Kind)
	}
	if evt.Special != input.KeyBackspace {
		t.Errorf("Special = %q, want %q", evt.Special, input.KeyBackspace)
	}
}

func TestTabEvent(t *testing.T) {
	evt := TabEvent()
	if evt.Kind != input.EventSpecial {
		t.Errorf("Kind = %d, want EventSpecial", evt.Kind)
	}
	if evt.Special != input.KeyTab {
		t.Errorf("Special = %q, want %q", evt.Special, input.KeyTab)
	}
}

func TestClickEvent(t *testing.T) {
	evt := ClickEvent(5, 10)
	if evt.Kind != input.EventMouse {
		t.Errorf("Kind = %d, want EventMouse", evt.Kind)
	}
	if evt.Mouse.Action != input.MousePress {
		t.Errorf("Action = %d, want MousePress", evt.Mouse.Action)
	}
	if evt.Mouse.Button != input.ButtonLeft {
		t.Errorf("Button = %d, want ButtonLeft", evt.Mouse.Button)
	}
	if evt.Mouse.Y != 5 {
		t.Errorf("Y = %d, want 5", evt.Mouse.Y)
	}
	if evt.Mouse.X != 10 {
		t.Errorf("X = %d, want 10", evt.Mouse.X)
	}
}

func TestScrollUpEvent(t *testing.T) {
	evt := ScrollUpEvent(3, 7)
	if evt.Kind != input.EventMouse {
		t.Errorf("Kind = %d, want EventMouse", evt.Kind)
	}
	if evt.Mouse.Action != input.MouseScroll {
		t.Errorf("Action = %d, want MouseScroll", evt.Mouse.Action)
	}
	if evt.Mouse.Button != input.ScrollUp {
		t.Errorf("Button = %d, want ScrollUp", evt.Mouse.Button)
	}
	if evt.Mouse.Y != 3 || evt.Mouse.X != 7 {
		t.Errorf("position = (%d,%d), want (3,7)", evt.Mouse.Y, evt.Mouse.X)
	}
}

func TestScrollDownEvent(t *testing.T) {
	evt := ScrollDownEvent(2, 4)
	if evt.Kind != input.EventMouse {
		t.Errorf("Kind = %d, want EventMouse", evt.Kind)
	}
	if evt.Mouse.Action != input.MouseScroll {
		t.Errorf("Action = %d, want MouseScroll", evt.Mouse.Action)
	}
	if evt.Mouse.Button != input.ScrollDown {
		t.Errorf("Button = %d, want ScrollDown", evt.Mouse.Button)
	}
}

func TestDragEvent(t *testing.T) {
	evt := DragEvent(1, 2)
	if evt.Kind != input.EventMouse {
		t.Errorf("Kind = %d, want EventMouse", evt.Kind)
	}
	if evt.Mouse.Action != input.MouseMotion {
		t.Errorf("Action = %d, want MouseMotion", evt.Mouse.Action)
	}
	if evt.Mouse.Y != 1 || evt.Mouse.X != 2 {
		t.Errorf("position = (%d,%d), want (1,2)", evt.Mouse.Y, evt.Mouse.X)
	}
}
