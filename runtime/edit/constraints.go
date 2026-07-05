package edit

import (
	"unicode/utf8"

	"github.com/tomyan/sumi/runtime/input"
)

// Constraints limit what editing keys may do to a State.
type Constraints struct {
	MaxLength int  // > 0 caps the value's rune length (typing and paste)
	ReadOnly  bool // block edits; caret movement stays allowed
	Multiline bool // Enter inserts a newline; Up/Down move between lines
}

// HandleKeyWith applies an event under constraints. Edits blocked by a
// constraint are still consumed (returns true) so they don't leak to
// other handlers; events the editor doesn't own return false.
func HandleKeyWith(s *State, evt input.Event, c Constraints) bool {
	if c.Multiline && evt.Kind == input.EventSpecial {
		switch evt.Special {
		case input.KeyUp:
			s.CursorUp()
			return true
		case input.KeyDown:
			s.CursorDown()
			return true
		case input.KeyEnter:
			if c.ReadOnly {
				return true
			}
			if c.MaxLength > 0 && len([]rune(s.Value)) >= c.MaxLength {
				return true
			}
			s.InsertNewline()
			return true
		}
	}
	if c.ReadOnly && !isNavigation(evt) {
		// Consume the event iff it would have edited; probe on a scratch
		// copy so the real state is untouched.
		scratch := &State{Value: s.Value, Cursor: s.Cursor}
		return HandleKey(scratch, evt)
	}
	if c.MaxLength > 0 {
		if handled, done := applyMaxLength(s, evt, c.MaxLength); done {
			return handled
		}
	}
	return HandleKey(s, evt)
}

// applyMaxLength intercepts growing edits at the cap. The second return
// reports whether the event was fully dealt with here.
func applyMaxLength(s *State, evt input.Event, max int) (handled, done bool) {
	length := utf8.RuneCountInString(s.Value)
	switch {
	case evt.Kind == input.EventKey && !evt.Ctrl:
		if length >= max {
			return true, true
		}
	case evt.Kind == input.EventPaste:
		room := max - length
		if room <= 0 {
			return true, true
		}
		runes := []rune(evt.PasteText)
		if len(runes) > room {
			s.InsertString(string(runes[:room]))
			return true, true
		}
	}
	return false, false
}

// isNavigation reports whether the event only moves the caret.
func isNavigation(evt input.Event) bool {
	if evt.Kind == input.EventSpecial {
		switch evt.Special {
		case input.KeyLeft, input.KeyRight, input.KeyHome, input.KeyEnd:
			return true
		}
		return false
	}
	if evt.Kind == input.EventKey && evt.Ctrl {
		switch evt.Rune {
		case 'a', 'e', 'b', 'f':
			return true
		}
	}
	return false
}
