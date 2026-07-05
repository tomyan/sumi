package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// dialogApp: a page button, and a dialog with two buttons inside.
func dialogApp(open bool, on map[string]func(*layout.DOMEvent)) (comp *tui.Component, pageBtn, dialog, yesBtn, noBtn *layout.Input) {
	attrs := map[string]string{}
	if open {
		attrs["open"] = "true"
	}
	pageBtn = &layout.Input{Kind: layout.KindText, Tag: "button", Content: "Page"}
	yesBtn = &layout.Input{Kind: layout.KindText, Tag: "button", Content: "Yes"}
	noBtn = &layout.Input{Kind: layout.KindText, Tag: "button", Content: "No"}
	dialog = &layout.Input{Kind: layout.KindBox, Tag: "dialog", Attrs: attrs, On: on,
		CursorCol: -1, CursorRow: -1, Children: []*layout.Input{yesBtn, noBtn}}
	comp = &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{pageBtn, dialog},
	}}
	return comp, pageBtn, dialog, yesBtn, noBtn
}

func TestDialogHiddenWithoutOpen(t *testing.T) {
	// Given / When
	comp, _, dialog, _, _ := dialogApp(false, nil)
	tui.TestApp(comp, 30, 6)

	// Then
	if !dialog.Hidden {
		t.Error("closed dialog should be hidden")
	}
}

func TestOpenDialogPullsFocusInAndTrapsTab(t *testing.T) {
	// Given — dialog open from the start
	comp, pageBtn, _, yesBtn, noBtn := dialogApp(true, nil)
	app := tui.TestApp(comp, 30, 6)

	// Then — focus starts inside the dialog, not on the page button
	if !yesBtn.Focused || pageBtn.Focused {
		t.Fatalf("focus = page %v yes %v, want dialog Yes focused", pageBtn.Focused, yesBtn.Focused)
	}

	// When — Tab cycles inside the dialog only
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})
	if !noBtn.Focused {
		t.Fatalf("after Tab: No should be focused")
	}
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then — wraps back to Yes, never reaching the page button
	if !yesBtn.Focused || pageBtn.Focused {
		t.Errorf("after wrap: yes %v page %v, want trap inside dialog", yesBtn.Focused, pageBtn.Focused)
	}
}

func TestEscapeClosesDialogAndFiresCloseEvent(t *testing.T) {
	// Given
	var closes []*layout.DOMEvent
	comp, _, dialog, _, _ := dialogApp(true, map[string]func(*layout.DOMEvent){
		"close": func(evt *layout.DOMEvent) { closes = append(closes, evt) },
	})
	app := tui.TestApp(comp, 30, 6)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEscape})

	// Then
	if _, open := dialog.Attrs["open"]; open {
		t.Error("Escape should remove the open attribute")
	}
	if !dialog.Hidden {
		t.Error("closed dialog should be hidden")
	}
	if len(closes) != 1 {
		t.Errorf("close events = %d, want 1", len(closes))
	}
}

func TestClicksOutsideOpenDialogAreCaptured(t *testing.T) {
	// Given — a click handler on the page button
	clicks := 0
	comp, pageBtn, _, _, _ := dialogApp(true, nil)
	pageBtn.On = map[string]func(*layout.DOMEvent){
		"click": func(evt *layout.DOMEvent) { clicks++ },
	}
	app := tui.TestApp(comp, 30, 6)

	// When — click the page button (row 0)
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: 1, Y: 0,
	}})

	// Then — the modal captured the click
	if clicks != 0 {
		t.Errorf("page button clicked %d times through a modal, want 0", clicks)
	}
}
