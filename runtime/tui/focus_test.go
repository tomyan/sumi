package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// focusComponent builds a component with two focusable inputs whose OnKey
// handlers record the events they receive.
func focusComponent() (comp *tui.Component, first, second *layout.Input, firstEvents, secondEvents *[]input.Event) {
	var fe, se []input.Event
	first = &layout.Input{
		Kind: layout.KindBox, Focusable: true,
		OnKey: func(evt input.Event) { fe = append(fe, evt) },
	}
	second = &layout.Input{
		Kind: layout.KindBox, Focusable: true,
		OnKey: func(evt input.Event) { se = append(se, evt) },
	}
	comp = &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{first, second},
		},
	}
	return comp, first, second, &fe, &se
}

func TestInitialFocusGoesToFirstFocusable(t *testing.T) {
	// Given
	comp, first, second, firstEvents, _ := focusComponent()

	// When
	tui.TestApp(comp, 20, 3)

	// Then
	if !first.Focused || second.Focused {
		t.Errorf("Focused flags = (%v, %v), want (true, false)", first.Focused, second.Focused)
	}
	if comp.FocusIndex != 0 {
		t.Errorf("FocusIndex = %d, want 0", comp.FocusIndex)
	}
	if len(*firstEvents) != 1 || (*firstEvents)[0].Kind != input.EventFocus {
		t.Errorf("first focusable events = %v, want one EventFocus", *firstEvents)
	}
}

func TestTabMovesFocusToNextFocusable(t *testing.T) {
	// Given
	comp, first, second, firstEvents, secondEvents := focusComponent()
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then
	if comp.FocusIndex != 1 {
		t.Errorf("FocusIndex = %d, want 1", comp.FocusIndex)
	}
	if first.Focused || !second.Focused {
		t.Errorf("Focused flags = (%v, %v), want (false, true)", first.Focused, second.Focused)
	}
	last := (*firstEvents)[len(*firstEvents)-1]
	if last.Kind != input.EventBlur {
		t.Errorf("first focusable last event = %v, want EventBlur", last)
	}
	if len(*secondEvents) != 1 || (*secondEvents)[0].Kind != input.EventFocus {
		t.Errorf("second focusable events = %v, want one EventFocus", *secondEvents)
	}
}

func TestShiftTabWrapsFocusBackward(t *testing.T) {
	// Given — focus starts on the first of two focusables
	comp, first, second, _, secondEvents := focusComponent()
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyShiftTab})

	// Then — wraps to the last focusable
	if comp.FocusIndex != 1 {
		t.Errorf("FocusIndex = %d, want 1", comp.FocusIndex)
	}
	if first.Focused || !second.Focused {
		t.Errorf("Focused flags = (%v, %v), want (false, true)", first.Focused, second.Focused)
	}
	if len(*secondEvents) != 1 || (*secondEvents)[0].Kind != input.EventFocus {
		t.Errorf("second focusable events = %v, want one EventFocus", *secondEvents)
	}
}

func TestTabConsumedWhenFocusablesExist(t *testing.T) {
	// Given
	comp, _, _, _, _ := focusComponent()
	var rootEvents []input.Event
	comp.OnEvent = func(evt input.Event) { rootEvents = append(rootEvents, evt) }
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then — Tab is a focus command, not delivered to the root handler
	for _, evt := range rootEvents {
		if evt.Kind == input.EventSpecial && evt.Special == input.KeyTab {
			t.Errorf("root OnEvent received Tab; want it consumed by focus cycling")
		}
	}
}

func TestTabPassesThroughWithoutFocusables(t *testing.T) {
	// Given — no focusable elements in the tree
	var rootEvents []input.Event
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{{Kind: layout.KindText, Content: "hi"}},
		},
		OnEvent: func(evt input.Event) { rootEvents = append(rootEvents, evt) },
	}
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then
	if len(rootEvents) != 1 || rootEvents[0].Special != input.KeyTab {
		t.Errorf("root OnEvent events = %v, want the Tab event delivered", rootEvents)
	}
}

func TestTabWithSingleFocusableKeepsFocus(t *testing.T) {
	// Given — one focusable: Tab has nowhere to go
	var events []input.Event
	only := &layout.Input{
		Kind: layout.KindBox, Focusable: true,
		OnKey: func(evt input.Event) { events = append(events, evt) },
	}
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{only},
		},
	}
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then — still focused, no blur dispatched
	if !only.Focused {
		t.Error("single focusable lost focus after Tab")
	}
	for _, evt := range events {
		if evt.Kind == input.EventBlur {
			t.Errorf("single focusable received EventBlur; events = %v", events)
		}
	}
}
