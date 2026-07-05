package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// focusComponent builds a component with two focusable inputs whose DOM
// handlers record the events they receive.
func focusComponent() (comp *tui.Component, first, second *layout.Input, firstEvents, secondEvents *[]*layout.DOMEvent) {
	var fe, se []*layout.DOMEvent
	record := func(events *[]*layout.DOMEvent) map[string]func(*layout.DOMEvent) {
		handler := func(evt *layout.DOMEvent) { *events = append(*events, evt) }
		return map[string]func(*layout.DOMEvent){
			"focus": handler, "blur": handler, "keydown": handler, "paste": handler,
		}
	}
	first = &layout.Input{Kind: layout.KindBox, Focusable: true, On: record(&fe)}
	second = &layout.Input{Kind: layout.KindBox, Focusable: true, On: record(&se)}
	comp = &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{first, second},
		},
	}
	return comp, first, second, &fe, &se
}

func lastEvent(events []*layout.DOMEvent) *layout.DOMEvent {
	if len(events) == 0 {
		return nil
	}
	return events[len(events)-1]
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
	if evt := lastEvent(*firstEvents); evt == nil || evt.Type != "focus" {
		t.Errorf("first focusable events = %v, want a focus event", *firstEvents)
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
	if evt := lastEvent(*firstEvents); evt == nil || evt.Type != "blur" {
		t.Errorf("first focusable last event = %v, want blur", evt)
	}
	if evt := lastEvent(*secondEvents); evt == nil || evt.Type != "focus" {
		t.Errorf("second focusable last event = %v, want focus", evt)
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
	if evt := lastEvent(*secondEvents); evt == nil || evt.Type != "focus" {
		t.Errorf("second focusable last event = %v, want focus", evt)
	}
}

func TestKeyEventTargetsFocusedElement(t *testing.T) {
	// Given
	comp, _, _, firstEvents, secondEvents := focusComponent()
	app := tui.TestApp(comp, 20, 3)

	// When — type a character while the first field is focused
	app.Step(input.Event{Kind: input.EventKey, Rune: 'a'})

	// Then — the focused field gets a keydown carrying the key
	evt := lastEvent(*firstEvents)
	if evt == nil || evt.Type != "keydown" || evt.Key.Rune != 'a' {
		t.Errorf("first focusable last event = %v, want keydown 'a'", evt)
	}
	for _, e := range *secondEvents {
		if e.Type == "keydown" {
			t.Error("unfocused element received keydown")
		}
	}
}

func TestKeyEventFollowsFocusAfterTab(t *testing.T) {
	// Given
	comp, _, _, _, secondEvents := focusComponent()
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'b'})

	// Then
	evt := lastEvent(*secondEvents)
	if evt == nil || evt.Type != "keydown" || evt.Key.Rune != 'b' {
		t.Errorf("second focusable last event = %v, want keydown 'b'", evt)
	}
}

func TestPasteEventTargetsFocusedElement(t *testing.T) {
	// Given
	comp, _, _, firstEvents, _ := focusComponent()
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventPaste, PasteText: "hello"})

	// Then
	evt := lastEvent(*firstEvents)
	if evt == nil || evt.Type != "paste" || evt.Key.PasteText != "hello" {
		t.Errorf("first focusable last event = %v, want paste 'hello'", evt)
	}
}

func TestStopPropagationSuppressesRootHandler(t *testing.T) {
	// Given — the focused field consumes keydown events
	var rootKeys []input.Event
	field := &layout.Input{
		Kind: layout.KindBox, Focusable: true,
		On: map[string]func(*layout.DOMEvent){
			"keydown": func(evt *layout.DOMEvent) { evt.StopPropagation() },
		},
	}
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{field},
		},
		OnEvent: func(evt input.Event) { rootKeys = append(rootKeys, evt) },
	}
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'q'})

	// Then — the root component handler never sees the consumed key
	for _, e := range rootKeys {
		if e.Kind == input.EventKey && e.Rune == 'q' {
			t.Error("root OnEvent received a key consumed by StopPropagation")
		}
	}
}

func TestRootHandlerStillSeesUnconsumedKeys(t *testing.T) {
	// Given — field handlers that do NOT stop propagation
	comp, _, _, _, _ := focusComponent()
	var rootKeys []input.Event
	comp.OnEvent = func(evt input.Event) { rootKeys = append(rootKeys, evt) }
	app := tui.TestApp(comp, 20, 3)

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then
	if len(rootKeys) != 1 || rootKeys[0].Rune != 'x' {
		t.Errorf("root events = %v, want the unconsumed 'x' key", rootKeys)
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
	var events []*layout.DOMEvent
	only := &layout.Input{
		Kind: layout.KindBox, Focusable: true,
		On: map[string]func(*layout.DOMEvent){
			"blur": func(evt *layout.DOMEvent) { events = append(events, evt) },
		},
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
	if len(events) != 0 {
		t.Errorf("single focusable received blur events: %v", events)
	}
}
