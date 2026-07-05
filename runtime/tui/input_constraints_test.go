package tui_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestInputMaxLengthAttr(t *testing.T) {
	// Given
	comp, field := inputElementApp(map[string]string{"value": "abcd", "maxlength": "5"}, nil)
	app := tui.TestApp(comp, 30, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'e'})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'f'})

	// Then
	if got := valueChild(t, field).Content; got != "abcde" {
		t.Errorf("value = %q, want capped \"abcde\"", got)
	}
}

func TestInputReadonlyAttr(t *testing.T) {
	// Given
	comp, field := inputElementApp(map[string]string{"value": "ro", "readonly": "true"}, nil)
	app := tui.TestApp(comp, 30, 3)

	// When — typing is swallowed, caret still moves
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyLeft})

	// Then
	if got := valueChild(t, field).Content; got != "ro" {
		t.Errorf("value = %q, want unchanged \"ro\"", got)
	}
	if field.CursorCol != 1 {
		t.Errorf("cursor = %d, want 1 (caret movement allowed)", field.CursorCol)
	}
}

func TestInputPasswordMasksDisplayNotEvents(t *testing.T) {
	// Given
	var values []string
	comp, field := inputElementApp(map[string]string{"type": "password"},
		map[string]func(*layout.DOMEvent){
			"input": func(evt *layout.DOMEvent) {
				values = append(values, evt.Data["value"].(string))
			},
		})
	app := tui.TestApp(comp, 30, 3)

	// When
	for _, r := range "hunter2" {
		app.Step(input.Event{Kind: input.EventKey, Rune: r})
	}

	// Then — bullets on screen, real value in events
	if got := valueChild(t, field).Content; got != strings.Repeat("•", 7) {
		t.Errorf("display = %q, want 7 bullets", got)
	}
	if len(values) == 0 || values[len(values)-1] != "hunter2" {
		t.Errorf("last input event value = %v, want hunter2", values)
	}
}

func TestInputEventOnlyFiresOnValueChange(t *testing.T) {
	// Given
	count := 0
	comp, _ := inputElementApp(map[string]string{"value": "ab"},
		map[string]func(*layout.DOMEvent){
			"input": func(evt *layout.DOMEvent) { count++ },
		})
	app := tui.TestApp(comp, 30, 3)

	// When — cursor moves don't change the value
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyLeft})
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyHome})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then
	if count != 1 {
		t.Errorf("input events = %d, want 1 (only the edit)", count)
	}
}

func TestInputWindowsLongValueAroundCursor(t *testing.T) {
	// Given — a value longer than the UA default width of 20
	comp, field := inputElementApp(map[string]string{"value": "abcdefghijklmnopqrstuvwxy"}, nil)
	tui.TestApp(comp, 40, 3)

	// Then — display shows the tail so the cursor (at end) is visible
	got := valueChild(t, field).Content
	if utf8.RuneCountInString(got) > 20 {
		t.Errorf("display %q exceeds width 20", got)
	}
	if !strings.HasSuffix(got, "y") {
		t.Errorf("display %q should end at the cursor tail", got)
	}
	if field.CursorCol < 0 || field.CursorCol > 19 {
		t.Errorf("cursor col %d outside the visible window", field.CursorCol)
	}
}
