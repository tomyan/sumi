package sumitest

import "github.com/tomyan/sumi/runtime/input"

// KeyEvent creates a regular key event.
func KeyEvent(r rune) input.Event {
	return input.Event{Kind: input.EventKey, Rune: r}
}

// CtrlEvent creates a Ctrl+key event.
func CtrlEvent(r rune) input.Event {
	return input.Event{Kind: input.EventKey, Rune: r, Ctrl: true}
}

// SpecialEvent creates a special key event.
func SpecialEvent(k input.SpecialKey) input.Event {
	return input.Event{Kind: input.EventSpecial, Special: k}
}

// PasteEvent creates a bracketed paste event.
func PasteEvent(text string) input.Event {
	return input.Event{Kind: input.EventPaste, PasteText: text}
}

// EnterEvent creates an Enter key event.
func EnterEvent() input.Event {
	return input.Event{Kind: input.EventSpecial, Special: input.KeyEnter}
}

// EscapeEvent creates an Escape key event.
func EscapeEvent() input.Event {
	return input.Event{Kind: input.EventSpecial, Special: input.KeyEscape}
}

// BackspaceEvent creates a Backspace key event.
func BackspaceEvent() input.Event {
	return input.Event{Kind: input.EventSpecial, Special: input.KeyBackspace}
}

// TabEvent creates a Tab key event.
func TabEvent() input.Event {
	return input.Event{Kind: input.EventSpecial, Special: input.KeyTab}
}

// ShiftTabEvent creates a Shift+Tab key event.
func ShiftTabEvent() input.Event {
	return input.Event{Kind: input.EventSpecial, Special: input.KeyShiftTab}
}

// ClickEvent creates a left mouse button press at (row, col).
func ClickEvent(row, col int) input.Event {
	return input.Event{
		Kind: input.EventMouse,
		Mouse: input.MouseEvent{
			Action: input.MousePress,
			Button: input.ButtonLeft,
			X:      col,
			Y:      row,
		},
	}
}

// DragEvent creates a mouse motion event at (row, col).
func DragEvent(row, col int) input.Event {
	return input.Event{
		Kind: input.EventMouse,
		Mouse: input.MouseEvent{
			Action: input.MouseMotion,
			X:      col,
			Y:      row,
		},
	}
}

// ScrollUpEvent creates a scroll-up event at (row, col).
func ScrollUpEvent(row, col int) input.Event {
	return input.Event{
		Kind: input.EventMouse,
		Mouse: input.MouseEvent{
			Action: input.MouseScroll,
			Button: input.ScrollUp,
			X:      col,
			Y:      row,
		},
	}
}

// ScrollDownEvent creates a scroll-down event at (row, col).
func ScrollDownEvent(row, col int) input.Event {
	return input.Event{
		Kind: input.EventMouse,
		Mouse: input.MouseEvent{
			Action: input.MouseScroll,
			Button: input.ScrollDown,
			X:      col,
			Y:      row,
		},
	}
}
