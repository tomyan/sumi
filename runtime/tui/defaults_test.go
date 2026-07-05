package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func twoFocusables(firstOn, secondOn map[string]func(*layout.DOMEvent)) (*tui.Component, *layout.Input, *layout.Input) {
	first := &layout.Input{Kind: layout.KindBox, Focusable: true, On: firstOn}
	second := &layout.Input{Kind: layout.KindBox, Focusable: true, On: secondOn}
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{first, second},
		},
	}
	return comp, first, second
}

func TestFocusedElementSeesTabAsKeydownBeforeCycling(t *testing.T) {
	// Given — the focused field records keydowns
	var keys []*layout.DOMEvent
	comp, _, _ := twoFocusables(map[string]func(*layout.DOMEvent){
		"keydown": func(evt *layout.DOMEvent) { keys = append(keys, evt) },
	}, nil)
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then — the field saw the Tab keydown, and focus still moved
	if len(keys) != 1 || keys[0].Key.Special != input.KeyTab {
		t.Errorf("focused field keydowns = %v, want the Tab keydown", keys)
	}
	if comp.FocusIndex != 1 {
		t.Errorf("FocusIndex = %d, want 1 (default action ran)", comp.FocusIndex)
	}
}

func TestPreventDefaultTrapsTab(t *testing.T) {
	// Given — the focused field traps Tab
	comp, first, _ := twoFocusables(map[string]func(*layout.DOMEvent){
		"keydown": func(evt *layout.DOMEvent) {
			if evt.Key.Special == input.KeyTab {
				evt.PreventDefault()
			}
		},
	}, nil)
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then — focus did not move
	if comp.FocusIndex != 0 || !first.Focused {
		t.Errorf("FocusIndex = %d, Focused = %v; want focus trapped on first", comp.FocusIndex, first.Focused)
	}
}

func TestEnterSynthesizesClickOnFocusedElement(t *testing.T) {
	// Given — the focused element has a click handler
	clicks := 0
	comp, _, _ := twoFocusables(map[string]func(*layout.DOMEvent){
		"click": func(evt *layout.DOMEvent) { clicks++ },
	}, nil)
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then
	if clicks != 1 {
		t.Errorf("clicks = %d, want 1 (Enter activates the focused element)", clicks)
	}
}

func TestEnterActivationConsumedWhenHandled(t *testing.T) {
	// Given — a click handler on the focused element and a root handler
	var rootEvents []input.Event
	comp, _, _ := twoFocusables(map[string]func(*layout.DOMEvent){
		"click": func(evt *layout.DOMEvent) {},
	}, nil)
	comp.OnEvent = func(evt input.Event) { rootEvents = append(rootEvents, evt) }
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then — Enter was consumed by the activation
	for _, e := range rootEvents {
		if e.Special == input.KeyEnter {
			t.Error("root handler received Enter although activation handled it")
		}
	}
}

func TestEnterPassesThroughWithoutClickHandler(t *testing.T) {
	// Given — focusable without a click handler
	var rootEvents []input.Event
	comp, _, _ := twoFocusables(nil, nil)
	comp.OnEvent = func(evt input.Event) { rootEvents = append(rootEvents, evt) }
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then — root still sees Enter
	if len(rootEvents) != 1 || rootEvents[0].Special != input.KeyEnter {
		t.Errorf("root events = %v, want the Enter event", rootEvents)
	}
}

func TestClickFocusesTheClickedFocusable(t *testing.T) {
	// Given — two focusable text rows; the first starts focused
	var secondEvents []*layout.DOMEvent
	first := &layout.Input{Kind: layout.KindText, Content: "one", Focusable: true}
	second := &layout.Input{Kind: layout.KindText, Content: "two", Focusable: true,
		On: map[string]func(*layout.DOMEvent){
			"focus": func(evt *layout.DOMEvent) { secondEvents = append(secondEvents, evt) },
		}}
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{first, second},
		},
	}
	app := tui.TestApp(comp, 20, 4)

	// When — click lands on the second row
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 1, Y: 1,
	}})

	// Then — focus moved to the clicked element
	if comp.FocusIndex != 1 || first.Focused || !second.Focused {
		t.Errorf("FocusIndex=%d first.Focused=%v second.Focused=%v; want click-to-focus on second",
			comp.FocusIndex, first.Focused, second.Focused)
	}
	if len(secondEvents) != 1 || secondEvents[0].Type != "focus" {
		t.Errorf("second events = %v, want one focus event", secondEvents)
	}
}

func anchorApp(t *testing.T) (*tui.Component, *[]string) {
	t.Helper()
	var opened []string
	prev := tui.OpenURL
	tui.OpenURL = func(href string) error { opened = append(opened, href); return nil }
	t.Cleanup(func() { tui.OpenURL = prev })

	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{
				{Kind: layout.KindText, Tag: "a", Content: "docs",
					Attrs: map[string]string{"href": "https://example.com/docs"}},
			},
		},
	}
	return comp, &opened
}

func TestEnterOpensFocusedAnchorHref(t *testing.T) {
	// Given — the anchor is the only focusable, so it starts focused
	comp, opened := anchorApp(t)
	app := tui.TestApp(comp, 30, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then
	if len(*opened) != 1 || (*opened)[0] != "https://example.com/docs" {
		t.Errorf("opened = %v, want the anchor href", *opened)
	}
}

func TestClickOpensAnchorHref(t *testing.T) {
	// Given
	comp, opened := anchorApp(t)
	app := tui.TestApp(comp, 30, 3)

	// When — click on the link text (row 0)
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 1, Y: 0,
	}})

	// Then
	if len(*opened) != 1 || (*opened)[0] != "https://example.com/docs" {
		t.Errorf("opened = %v, want the anchor href", *opened)
	}
}

func TestPreventDefaultSuppressesAnchorOpen(t *testing.T) {
	// Given — a click handler on the anchor prevents the default
	comp, opened := anchorApp(t)
	comp.Tree.Children[0].On = map[string]func(*layout.DOMEvent){
		"click": func(evt *layout.DOMEvent) { evt.PreventDefault() },
	}
	app := tui.TestApp(comp, 30, 3)

	// When
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 1, Y: 0,
	}})

	// Then
	if len(*opened) != 0 {
		t.Errorf("opened = %v, want none (preventDefault)", *opened)
	}
}

func TestPreventDefaultSuppressesEnterActivation(t *testing.T) {
	// Given — keydown prevents the default, click would record
	clicks := 0
	comp, _, _ := twoFocusables(map[string]func(*layout.DOMEvent){
		"keydown": func(evt *layout.DOMEvent) { evt.PreventDefault() },
		"click":   func(evt *layout.DOMEvent) { clicks++ },
	}, nil)
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then
	if clicks != 0 {
		t.Errorf("clicks = %d, want 0 (preventDefault suppresses activation)", clicks)
	}
}
