package preview

import "github.com/tomyan/sumi/runtime/input"

// FocusState tracks which UI element currently receives keyboard input.
type FocusState int

const (
	FocusControls    FocusState = iota // preview commands (h/l/u/q/i)
	FocusEditor1                       // source .sumi file
	FocusEditor2                       // snapshot file
	FocusEditor3                       // scenario.go file
	FocusInteractive                   // component subprocess
)

// Name returns a human-readable name for the focus state.
func (f FocusState) Name() string {
	switch f {
	case FocusControls:
		return "controls"
	case FocusEditor1:
		return "source"
	case FocusEditor2:
		return "snapshot"
	case FocusEditor3:
		return "scenario"
	case FocusInteractive:
		return "interactive"
	default:
		return "unknown"
	}
}

// Package-level focus state.
var (
	pvFocus         FocusState
	pvPrefixPending bool
)

// focusForDigit returns the editor focus state for a digit key.
func focusForDigit(r rune) FocusState {
	switch r {
	case '1':
		return FocusEditor1
	case '2':
		return FocusEditor2
	case '3':
		return FocusEditor3
	default:
		return FocusControls
	}
}

// prefixCommand interprets a key event as a Ctrl+\ prefix command.
// Returns the command name or "" if unrecognized.
func prefixCommand(evt input.Event) string {
	if isCtrlBackslash(evt) {
		return "exit"
	}
	if evt.Kind != input.EventKey {
		return ""
	}
	switch evt.Rune {
	case 'q':
		return "quit"
	case 'l':
		return "next"
	case 'h':
		return "prev"
	case 'u':
		return "update"
	case 'i':
		return "interactive"
	default:
		return ""
	}
}

// isCtrlBackslash checks if an event is Ctrl+\ (byte 0x1c).
func isCtrlBackslash(evt input.Event) bool {
	return evt.Kind == input.EventKey && evt.Rune == 0x1c
}

// editorIndex returns the 0-based editor index for a focus state, or -1.
func editorIndex(f FocusState) int {
	switch f {
	case FocusEditor1:
		return 0
	case FocusEditor2:
		return 1
	case FocusEditor3:
		return 2
	default:
		return -1
	}
}
