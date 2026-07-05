package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func checkboxApp(attrs map[string]string, on map[string]func(*layout.DOMEvent)) (*tui.Component, *layout.Input) {
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["type"] = "checkbox"
	field := &layout.Input{Kind: layout.KindBox, Tag: "input", Attrs: attrs, On: on,
		CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{field},
	}}
	return comp, field
}

func TestCheckboxRendersGlyphs(t *testing.T) {
	// Given / When
	comp, field := checkboxApp(map[string]string{"checked": "true"}, nil)
	tui.TestApp(comp, 20, 3)

	// Then
	if got := valueChild(t, field).Content; got != "[x]" {
		t.Errorf("checked checkbox renders %q, want \"[x]\"", got)
	}
	if field.CursorCol != -1 {
		t.Errorf("checkbox has a caret (CursorCol %d)", field.CursorCol)
	}
}

func TestSpaceTogglesCheckbox(t *testing.T) {
	// Given
	var changes []*layout.DOMEvent
	comp, field := checkboxApp(map[string]string{"value": "opt1"},
		map[string]func(*layout.DOMEvent){
			"change": func(evt *layout.DOMEvent) { changes = append(changes, evt) },
		})
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: ' '})

	// Then
	if got := valueChild(t, field).Content; got != "[x]" {
		t.Errorf("after Space: %q, want \"[x]\"", got)
	}
	if len(changes) != 1 {
		t.Fatalf("change events = %d, want 1", len(changes))
	}
	if c, _ := changes[0].Data["checked"].(bool); !c {
		t.Errorf("change event checked = %v, want true", changes[0].Data["checked"])
	}
	if v, _ := changes[0].Data["value"].(string); v != "opt1" {
		t.Errorf("change event value = %v, want opt1", changes[0].Data["value"])
	}

	// When — Space again unchecks
	app.Step(input.Event{Kind: input.EventKey, Rune: ' '})

	// Then
	if got := valueChild(t, field).Content; got != "[ ]" {
		t.Errorf("after second Space: %q, want \"[ ]\"", got)
	}
}

func TestClickAndEnterToggleCheckbox(t *testing.T) {
	// Given
	comp, field := checkboxApp(nil, nil)
	app := tui.TestApp(comp, 20, 3)

	// When — click on the glyph
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 1, Y: 0,
	}})

	// Then
	if got := valueChild(t, field).Content; got != "[x]" {
		t.Errorf("after click: %q, want \"[x]\"", got)
	}

	// When — Enter activates too
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then
	if got := valueChild(t, field).Content; got != "[ ]" {
		t.Errorf("after Enter: %q, want \"[ ]\"", got)
	}
}

func TestTypingDoesNotEditCheckbox(t *testing.T) {
	// Given
	comp, field := checkboxApp(nil, nil)
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then — no text editing on checkables
	if got := valueChild(t, field).Content; got != "[ ]" {
		t.Errorf("after typing: %q, want \"[ ]\"", got)
	}
}

func radioGroupApp() (*tui.Component, *layout.Input, *layout.Input) {
	a := &layout.Input{Kind: layout.KindBox, Tag: "input", CursorCol: -1, CursorRow: -1,
		Attrs: map[string]string{"type": "radio", "name": "size", "value": "s", "checked": "true"}}
	b := &layout.Input{Kind: layout.KindBox, Tag: "input", CursorCol: -1, CursorRow: -1,
		Attrs: map[string]string{"type": "radio", "name": "size", "value": "m"}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{a, b},
	}}
	return comp, a, b
}

func TestRadioSelectionMovesWithinGroup(t *testing.T) {
	// Given
	comp, a, b := radioGroupApp()
	app := tui.TestApp(comp, 20, 4)

	if got := valueChild(t, a).Content; got != "(•)" {
		t.Fatalf("initial first radio %q, want \"(•)\"", got)
	}

	// When — Tab to the second radio and select it
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})
	app.Step(input.Event{Kind: input.EventKey, Rune: ' '})

	// Then — selection moved: first unchecked, second checked
	if got := valueChild(t, a).Content; got != "( )" {
		t.Errorf("first radio = %q, want \"( )\"", got)
	}
	if got := valueChild(t, b).Content; got != "(•)" {
		t.Errorf("second radio = %q, want \"(•)\"", got)
	}
}

func TestRadioNeverUntogglesItself(t *testing.T) {
	// Given — first radio checked and focused
	comp, a, _ := radioGroupApp()
	app := tui.TestApp(comp, 20, 4)

	// When — Space on the already-checked radio
	app.Step(input.Event{Kind: input.EventKey, Rune: ' '})

	// Then
	if got := valueChild(t, a).Content; got != "(•)" {
		t.Errorf("radio = %q, want still \"(•)\"", got)
	}
}
