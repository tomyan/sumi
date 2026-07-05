package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func selectApp(on map[string]func(*layout.DOMEvent)) (*tui.Component, *layout.Input) {
	sel := &layout.Input{Kind: layout.KindBox, Tag: "select", On: on,
		CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindText, Tag: "option", Content: "Small",
				Attrs: map[string]string{"value": "s"}},
			{Kind: layout.KindText, Tag: "option", Content: "Medium",
				Attrs: map[string]string{"value": "m", "selected": "true"}},
			{Kind: layout.KindText, Tag: "option", Content: "Enormous",
				Attrs: map[string]string{"value": "xl"}},
		}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{sel},
	}}
	return comp, sel
}

// selectDisplay finds the implicit projection child (untagged text).
func selectDisplay(t *testing.T, sel *layout.Input) *layout.Input {
	t.Helper()
	for _, c := range sel.Children {
		if c != nil && c.Kind == layout.KindText && c.Tag == "" {
			return c
		}
	}
	t.Fatal("select has no projection child")
	return nil
}

func TestSelectShowsSelectedOptionWithMarker(t *testing.T) {
	// Given / When
	comp, sel := selectApp(nil)
	tui.TestApp(comp, 30, 3)

	// Then — the selected attr picks the initial option
	if got := selectDisplay(t, sel).Content; got != "Medium ▾" {
		t.Errorf("display = %q, want \"Medium ▾\"", got)
	}
	// Options are hidden from layout.
	for _, c := range sel.Children {
		if c.Tag == "option" && c.Display != "none" {
			t.Errorf("option %q visible in layout", c.Content)
		}
	}
	// Sized to the longest option plus the marker.
	if sel.FixedWidth != len("Enormous")+2 {
		t.Errorf("FixedWidth = %d, want %d", sel.FixedWidth, len("Enormous")+2)
	}
}

func TestSelectArrowsMoveWithWraparound(t *testing.T) {
	// Given
	var changes []string
	comp, sel := selectApp(map[string]func(*layout.DOMEvent){
		"change": func(evt *layout.DOMEvent) {
			changes = append(changes, evt.Data["value"].(string))
		},
	})
	app := tui.TestApp(comp, 30, 3)

	// When / Then — down from Medium → Enormous, down again wraps to Small
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyDown})
	if got := selectDisplay(t, sel).Content; got != "Enormous ▾" {
		t.Errorf("after Down: %q", got)
	}
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyDown})
	if got := selectDisplay(t, sel).Content; got != "Small ▾" {
		t.Errorf("after wrap: %q", got)
	}
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyUp})
	if got := selectDisplay(t, sel).Content; got != "Enormous ▾" {
		t.Errorf("after Up: %q", got)
	}

	if len(changes) != 3 || changes[0] != "xl" || changes[1] != "s" || changes[2] != "xl" {
		t.Errorf("change values = %v, want [xl s xl]", changes)
	}
}

func TestSelectSpaceEnterAndClickAdvance(t *testing.T) {
	// Given
	comp, sel := selectApp(nil)
	app := tui.TestApp(comp, 30, 3)

	// When / Then
	app.Step(input.Event{Kind: input.EventKey, Rune: ' '})
	if got := selectDisplay(t, sel).Content; got != "Enormous ▾" {
		t.Errorf("after Space: %q", got)
	}
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})
	if got := selectDisplay(t, sel).Content; got != "Small ▾" {
		t.Errorf("after Enter: %q", got)
	}
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 1, Y: 0,
	}})
	if got := selectDisplay(t, sel).Content; got != "Medium ▾" {
		t.Errorf("after click: %q", got)
	}
}
