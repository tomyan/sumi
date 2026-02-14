package main

import (
	"strings"

	"github.com/tomyan/sumi/runtime/input"
)

func isQuitKey(evt input.Event) bool {
	return evt.Kind == input.EventKey && evt.Rune == 'q'
}

func isNextKey(evt input.Event) bool {
	if evt.Kind == input.EventKey && (evt.Rune == '\r' || evt.Rune == '\n' || evt.Rune == 'l') {
		return true
	}
	return evt.Kind == input.EventSpecial && evt.Special == input.KeyRight
}

func isPrevKey(evt input.Event) bool {
	if evt.Kind == input.EventKey && evt.Rune == 'h' {
		return true
	}
	return evt.Kind == input.EventSpecial && evt.Special == input.KeyLeft
}

func isScrollDown(evt input.Event) bool {
	if evt.Kind == input.EventKey && evt.Rune == 'j' {
		return true
	}
	if evt.Kind == input.EventSpecial && evt.Special == input.KeyDown {
		return true
	}
	return evt.Kind == input.EventSpecial && evt.Special == input.KeyPgDn
}

func isScrollUp(evt input.Event) bool {
	if evt.Kind == input.EventKey && evt.Rune == 'k' {
		return true
	}
	if evt.Kind == input.EventSpecial && evt.Special == input.KeyUp {
		return true
	}
	return evt.Kind == input.EventSpecial && evt.Special == input.KeyPgUp
}

func scrollAmount(evt input.Event) int {
	if evt.Kind == input.EventSpecial &&
		(evt.Special == input.KeyPgDn || evt.Special == input.KeyPgUp) {
		return 10
	}
	return 1
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline")
}
