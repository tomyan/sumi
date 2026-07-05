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
