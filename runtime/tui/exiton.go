package tui

import (
	"strings"

	"github.com/tomyan/sumi/runtime/input"
)

// defaultExitOn is the quit chord when App.ExitOn is unset.
var defaultExitOn = []string{"ctrl+c"}

// matchesExitChord reports whether the event matches any configured quit
// chord. Chords are "ctrl+<letter>", a special key name ("escape",
// "enter", ...), or a single character.
func matchesExitChord(evt input.Event, chords []string) bool {
	if len(chords) == 0 {
		chords = defaultExitOn
	}
	for _, chord := range chords {
		if matchesChord(evt, chord) {
			return true
		}
	}
	return false
}

func matchesChord(evt input.Event, chord string) bool {
	if ctrl, ok := strings.CutPrefix(chord, "ctrl+"); ok {
		return evt.Kind == input.EventKey && evt.Ctrl && len(ctrl) == 1 && evt.Rune == rune(ctrl[0])
	}
	if len([]rune(chord)) == 1 {
		return evt.Kind == input.EventKey && !evt.Ctrl && evt.Rune == []rune(chord)[0]
	}
	return evt.Kind == input.EventSpecial && string(evt.Special) == chord
}
