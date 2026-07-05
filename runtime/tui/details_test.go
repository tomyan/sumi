package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func detailsApp(open bool, on map[string]func(*layout.DOMEvent)) (*tui.Component, *layout.Input, *layout.Input, *layout.Input) {
	attrs := map[string]string{}
	if open {
		attrs["open"] = "true"
	}
	summary := &layout.Input{Kind: layout.KindText, Tag: "summary", Content: "More info"}
	content := &layout.Input{Kind: layout.KindText, Tag: "span", Content: "Hidden facts"}
	details := &layout.Input{Kind: layout.KindBox, Tag: "details", Attrs: attrs, On: on,
		CursorCol: -1, CursorRow: -1, Children: []*layout.Input{summary, content}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{details},
	}}
	return comp, details, summary, content
}

func TestClosedDetailsHidesContentAndShowsMarker(t *testing.T) {
	// Given / When
	comp, _, summary, content := detailsApp(false, nil)
	tui.TestApp(comp, 30, 4)

	// Then
	if summary.Content != "▶ More info" {
		t.Errorf("summary = %q, want \"▶ More info\"", summary.Content)
	}
	if content.Display != "none" {
		t.Errorf("content Display = %q, want none", content.Display)
	}
}

func TestOpenDetailsShowsContent(t *testing.T) {
	// Given / When
	comp, _, summary, content := detailsApp(true, nil)
	tui.TestApp(comp, 30, 4)

	// Then
	if summary.Content != "▼ More info" {
		t.Errorf("summary = %q, want \"▼ More info\"", summary.Content)
	}
	if content.Display == "none" {
		t.Error("open details must show its content")
	}
}

func TestEnterOnSummaryTogglesAndFiresToggleEvent(t *testing.T) {
	// Given — summary is focusable and starts focused
	var toggles []*layout.DOMEvent
	comp, _, summary, content := detailsApp(false, map[string]func(*layout.DOMEvent){
		"toggle": func(evt *layout.DOMEvent) { toggles = append(toggles, evt) },
	})
	app := tui.TestApp(comp, 30, 4)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then
	if summary.Content != "▼ More info" || content.Display == "none" {
		t.Errorf("details did not open: summary %q, content display %q", summary.Content, content.Display)
	}
	if len(toggles) != 1 {
		t.Fatalf("toggle events = %d, want 1", len(toggles))
	}
	if open, _ := toggles[0].Data["open"].(bool); !open {
		t.Errorf("toggle event open = %v, want true", toggles[0].Data["open"])
	}

	// When — toggle back closed
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})

	// Then
	if content.Display != "none" {
		t.Error("details did not close again")
	}
}

func TestClickOnSummaryToggles(t *testing.T) {
	// Given
	comp, _, _, content := detailsApp(false, nil)
	app := tui.TestApp(comp, 30, 4)

	// When — click the summary row
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 2, Y: 0,
	}})

	// Then
	if content.Display == "none" {
		t.Error("click on summary should open the details")
	}
}

func TestHiddenContentExcludedFromFocusTraversal(t *testing.T) {
	// Given — a focusable inside closed details
	button := &layout.Input{Kind: layout.KindText, Tag: "button", Content: "Hidden"}
	summary := &layout.Input{Kind: layout.KindText, Tag: "summary", Content: "More"}
	details := &layout.Input{Kind: layout.KindBox, Tag: "details",
		CursorCol: -1, CursorRow: -1, Children: []*layout.Input{summary, button}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{details},
	}}
	tui.TestApp(comp, 30, 4)

	// Then — only the summary is reachable
	focusables := layout.CollectFocusables(comp.Tree)
	if len(focusables) != 1 || focusables[0] != summary {
		t.Errorf("focusables = %v, want only the summary", focusables)
	}
}
