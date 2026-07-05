package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func inputElementApp(attrs map[string]string, on map[string]func(*layout.DOMEvent)) (*tui.Component, *layout.Input) {
	field := &layout.Input{Kind: layout.KindBox, Tag: "input", Attrs: attrs, On: on,
		CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{field},
		},
	}
	return comp, field
}

func valueChild(t *testing.T, field *layout.Input) *layout.Input {
	t.Helper()
	if len(field.Children) == 0 {
		t.Fatal("input element has no value child")
	}
	return field.Children[0]
}

func TestInputElementShowsInitialValue(t *testing.T) {
	// Given
	comp, field := inputElementApp(map[string]string{"value": "hi"}, nil)

	// When
	tui.TestApp(comp, 30, 3)

	// Then — the value renders and the cursor sits at the end
	if got := valueChild(t, field).Content; got != "hi" {
		t.Errorf("value child content = %q, want \"hi\"", got)
	}
	if field.CursorCol != 2 || field.CursorRow != 0 {
		t.Errorf("cursor = (%d,%d), want (2,0)", field.CursorCol, field.CursorRow)
	}
}

func TestInputElementTypingUpdatesValue(t *testing.T) {
	// Given
	comp, field := inputElementApp(map[string]string{"value": "hi"}, nil)
	app := tui.TestApp(comp, 30, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyBackspace})
	app.Step(input.Event{Kind: input.EventKey, Rune: '!'})

	// Then
	if got := valueChild(t, field).Content; got != "hi!" {
		t.Errorf("value = %q, want \"hi!\"", got)
	}
}

func TestInputElementDispatchesInputEvents(t *testing.T) {
	// Given
	var events []*layout.DOMEvent
	comp, _ := inputElementApp(nil, map[string]func(*layout.DOMEvent){
		"input": func(evt *layout.DOMEvent) { events = append(events, evt) },
	})
	app := tui.TestApp(comp, 30, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'a'})

	// Then — input event carries value and cursor
	if len(events) != 1 {
		t.Fatalf("input events = %d, want 1", len(events))
	}
	if v, _ := events[0].Data["value"].(string); v != "a" {
		t.Errorf("event value = %v, want \"a\"", events[0].Data["value"])
	}
	if c, _ := events[0].Data["cursor"].(int); c != 1 {
		t.Errorf("event cursor = %v, want 1", events[0].Data["cursor"])
	}
}

func TestInputElementConsumesEditingKeys(t *testing.T) {
	// Given
	var rootEvents []input.Event
	comp, _ := inputElementApp(nil, nil)
	comp.OnEvent = func(evt input.Event) { rootEvents = append(rootEvents, evt) }
	app := tui.TestApp(comp, 30, 3)

	// When — a printable key edits; Escape is not an editing key
	app.Step(input.Event{Kind: input.EventKey, Rune: 'a'})
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEscape})

	// Then
	for _, e := range rootEvents {
		if e.Kind == input.EventKey {
			t.Error("root handler saw a consumed editing key")
		}
	}
	if len(rootEvents) != 1 || rootEvents[0].Special != input.KeyEscape {
		t.Errorf("root events = %v, want just Escape", rootEvents)
	}
}

func TestInputElementCursorFollowsFocus(t *testing.T) {
	// Given — two inputs stacked
	first := &layout.Input{Kind: layout.KindBox, Tag: "input",
		Attrs: map[string]string{"value": "one"}, CursorCol: -1, CursorRow: -1}
	second := &layout.Input{Kind: layout.KindBox, Tag: "input",
		Attrs: map[string]string{"value": "two"}, CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{first, second},
	}}
	app := tui.TestApp(comp, 30, 4)

	// Then — only the focused input shows a cursor
	if first.CursorCol != 3 || second.CursorCol != -1 {
		t.Fatalf("cursors = (%d, %d), want (3, -1)", first.CursorCol, second.CursorCol)
	}

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then
	if first.CursorCol != -1 || second.CursorCol != 3 {
		t.Errorf("cursors after Tab = (%d, %d), want (-1, 3)", first.CursorCol, second.CursorCol)
	}
}

func TestPreventDefaultBlocksEditing(t *testing.T) {
	// Given — keydown handler vetoes editing
	comp, field := inputElementApp(map[string]string{"value": "ro"},
		map[string]func(*layout.DOMEvent){
			"keydown": func(evt *layout.DOMEvent) { evt.PreventDefault() },
		})
	app := tui.TestApp(comp, 30, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then
	if got := valueChild(t, field).Content; got != "ro" {
		t.Errorf("value = %q, want unchanged \"ro\"", got)
	}
}
