package edit

import "github.com/tomyan/sumi/runtime/input"

// HandleKey applies a terminal input event to the editing state using
// readline-style bindings. Returns true when the event was handled;
// events the editor does not own (Tab, Enter, Escape, unbound chords)
// are left for the caller.
func HandleKey(s *State, evt input.Event) bool {
	switch evt.Kind {
	case input.EventKey:
		return handleKeyRune(s, evt)
	case input.EventSpecial:
		return handleKeySpecial(s, evt.Special)
	case input.EventPaste:
		s.InsertString(evt.PasteText)
		return true
	}
	return false
}

func handleKeyRune(s *State, evt input.Event) bool {
	if !evt.Ctrl {
		s.Insert(evt.Rune)
		return true
	}
	switch evt.Rune {
	case 'a':
		s.Home()
	case 'e':
		s.End()
	case 'b':
		s.Left()
	case 'f':
		s.Right()
	case 'h':
		s.Backspace()
	case 'd':
		s.Delete()
	case 'k':
		s.KillToEnd()
	case 'u':
		s.KillToStart()
	case 'w':
		s.KillWord()
	case 'y':
		s.Yank()
	case 't':
		s.TransposeChars()
	default:
		return false
	}
	return true
}

func handleKeySpecial(s *State, key input.SpecialKey) bool {
	switch key {
	case input.KeyLeft:
		s.Left()
	case input.KeyRight:
		s.Right()
	case input.KeyHome:
		s.Home()
	case input.KeyEnd:
		s.End()
	case input.KeyBackspace:
		s.Backspace()
	case input.KeyDelete:
		s.Delete()
	default:
		return false
	}
	return true
}
