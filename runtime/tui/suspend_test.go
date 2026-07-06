package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// D4b: Ctrl+Z suspends — terminal restored, process stopped, terminal
// re-entered and repainted on resume. Hooks are injectable; the real
// implementation raises SIGTSTP.

func suspendApp(handler func(*layout.DOMEvent)) (*tui.App, *[]string) {
	var calls []string
	tree := &layout.Input{Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindText, Content: "hi", CursorCol: -1, CursorRow: -1,
				Attrs: map[string]string{"focusable": "true"}, Focusable: true,
				On: onMap(handler)},
		}}
	comp := &tui.Component{Tree: tree}
	app := tui.TestApp(comp, 10, 3)
	app.SuspendHooks = &tui.SuspendHooks{
		ExitTerminal:  func() { calls = append(calls, "exit") },
		Stop:          func() { calls = append(calls, "stop") },
		EnterTerminal: func() { calls = append(calls, "enter") },
	}
	return app, &calls
}

func onMap(handler func(*layout.DOMEvent)) map[string]func(*layout.DOMEvent) {
	if handler == nil {
		return nil
	}
	return map[string]func(*layout.DOMEvent){"keydown": handler}
}

func ctrlZ() input.Event {
	return input.Event{Kind: input.EventKey, Rune: 'z', Ctrl: true}
}

func TestCtrlZSuspendsInOrder(t *testing.T) {
	// Given
	app, calls := suspendApp(nil)

	// When
	app.Step(ctrlZ())

	// Then: terminal exits before the stop, re-enters after.
	want := []string{"exit", "stop", "enter"}
	if len(*calls) != 3 || (*calls)[0] != want[0] || (*calls)[1] != want[1] || (*calls)[2] != want[2] {
		t.Errorf("calls = %v, want %v", *calls, want)
	}
}

func TestCtrlZPreventDefaultSkipsSuspend(t *testing.T) {
	// Given: a keydown handler that claims Ctrl+Z (e.g. undo).
	app, calls := suspendApp(func(evt *layout.DOMEvent) {
		if evt.Key.Ctrl && evt.Key.Rune == 'z' {
			evt.PreventDefault()
		}
	})

	// When
	app.Step(ctrlZ())

	// Then
	if len(*calls) != 0 {
		t.Errorf("calls = %v, want none (default prevented)", *calls)
	}
}

func TestOtherKeysDoNotSuspend(t *testing.T) {
	// Given
	app, calls := suspendApp(nil)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'z'})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'c', Ctrl: true})

	// Then
	if len(*calls) != 0 {
		t.Errorf("calls = %v, want none", *calls)
	}
}
