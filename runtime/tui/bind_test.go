package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

// boundInputApp mirrors what codegen emits for bind:value on a text input:
// an "input" handler pushing the value into the signal, plus an effect that
// re-projects the signal onto the control.
func boundInputApp(sig *signal.Signal[string]) (*tui.Component, *layout.Input) {
	field := &layout.Input{
		Kind: layout.KindBox, Tag: "input",
		Attrs: map[string]string{"type": "text"},
		On: map[string]func(*layout.DOMEvent){
			"input": func(evt *layout.DOMEvent) { sig.Set(evt.Data["value"].(string)) },
		},
		CursorCol: -1, CursorRow: -1,
	}
	signal.Effect(func() { tui.BindInputValue(field, sig.Get()) })
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{field},
		},
	}
	return comp, field
}

func TestBindValueTypingUpdatesSignal(t *testing.T) {
	// Given — a text input two-way bound to a signal
	sig := signal.New("")
	comp, _ := boundInputApp(sig)
	app := tui.TestApp(comp, 30, 3)

	// When — the user types
	app.Step(input.Event{Kind: input.EventKey, Rune: 'h'})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'i'})

	// Then — the signal reflects the typed value
	if got := sig.Get(); got != "hi" {
		t.Errorf("signal = %q, want \"hi\"", got)
	}
}

func TestBindValueExternalSetUpdatesControl(t *testing.T) {
	// Given
	sig := signal.New("")
	comp, field := boundInputApp(sig)
	app := tui.TestApp(comp, 30, 3)

	// When — external code sets the signal
	sig.Set("hello")
	app.Render()

	// Then — the rendered control shows the new value
	if got := field.Children[0].Content; got != "hello" {
		t.Errorf("control value = %q, want \"hello\"", got)
	}
	// And the cursor sits at the end of the adopted value.
	if field.CursorCol != 5 {
		t.Errorf("cursor = %d, want 5", field.CursorCol)
	}
}

func TestBindCheckedReflectsSignal(t *testing.T) {
	// Given — a checkbox two-way bound to a bool signal
	sig := signal.New(false)
	field := &layout.Input{
		Kind: layout.KindBox, Tag: "input",
		Attrs: map[string]string{"type": "checkbox"},
		On: map[string]func(*layout.DOMEvent){
			"change": func(evt *layout.DOMEvent) { sig.Set(evt.Data["checked"].(bool)) },
		},
		CursorCol: -1, CursorRow: -1,
	}
	signal.Effect(func() { tui.BindChecked(field, sig.Get()) })
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{field},
	}}
	app := tui.TestApp(comp, 30, 3)

	// When — external Set checks the box
	sig.Set(true)
	app.Render()

	// Then — the glyph reflects the checked state
	if got := field.Children[0].Content; got != "[x]" {
		t.Errorf("checkbox glyph = %q, want \"[x]\"", got)
	}

	// When — the user toggles via Space (change event fires)
	app.Step(input.Event{Kind: input.EventKey, Rune: ' '})

	// Then — the signal reflects the new state
	if sig.Get() {
		t.Errorf("signal = true, want false after toggle")
	}
}
