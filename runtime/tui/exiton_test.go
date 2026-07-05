package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func plainApp() *tui.Component {
	return &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{{Kind: layout.KindText, Content: "hi"}},
	}}
}

func TestExitOnDefaultsToCtrlC(t *testing.T) {
	// Given
	app := tui.TestApp(plainApp(), 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'c', Ctrl: true})

	// Then
	if !app.QuitRequested() {
		t.Error("Ctrl+C should request quit by default")
	}
}

func TestExitOnCustomChords(t *testing.T) {
	// Given — quit on q or ctrl+d, and NOT on ctrl+c
	comp := plainApp()
	app := tui.TestApp(comp, 20, 3)
	app.ExitOn = []string{"q", "ctrl+d"}

	// When / Then
	app.Step(input.Event{Kind: input.EventKey, Rune: 'c', Ctrl: true})
	if app.QuitRequested() {
		t.Fatal("ctrl+c should not quit when exitOn overrides it")
	}
	app.Step(input.Event{Kind: input.EventKey, Rune: 'q'})
	if !app.QuitRequested() {
		t.Error("q should quit")
	}
}

func TestExitOnEscapeChord(t *testing.T) {
	// Given
	app := tui.TestApp(plainApp(), 20, 3)
	app.ExitOn = []string{"escape"}

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEscape})

	// Then
	if !app.QuitRequested() {
		t.Error("escape should quit")
	}
}

func TestExitChordSuppressedWhenConsumed(t *testing.T) {
	// Given — a focused input consumes printable keys
	field := &layout.Input{Kind: layout.KindBox, Tag: "input",
		CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{field},
	}}
	app := tui.TestApp(comp, 20, 3)
	app.ExitOn = []string{"q"}

	// When — typing q goes into the input, not the quit chord
	app.Step(input.Event{Kind: input.EventKey, Rune: 'q'})

	// Then
	if app.QuitRequested() {
		t.Error("q consumed by the focused input must not quit")
	}
}
