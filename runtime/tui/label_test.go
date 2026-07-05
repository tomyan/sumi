package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func clickAt(app *tui.App, x, y int) {
	app.Step(input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{
		Action: input.MousePress, Button: input.ButtonLeft, X: x, Y: y,
	}})
}

func TestClickingWrappingLabelTogglesControl(t *testing.T) {
	// Given — a label wrapping a checkbox and its caption
	box := &layout.Input{Kind: layout.KindBox, Tag: "input", CursorCol: -1, CursorRow: -1,
		Attrs: map[string]string{"type": "checkbox"}}
	caption := &layout.Input{Kind: layout.KindText, Content: "Enable turbo"}
	label := &layout.Input{Kind: layout.KindBox, Tag: "label", Direction: "row",
		CursorCol: -1, CursorRow: -1, Children: []*layout.Input{box, caption}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{label},
	}}
	app := tui.TestApp(comp, 30, 3)

	// When — click on the caption text, not the checkbox itself
	clickAt(app, 10, 0)

	// Then — the wrapped checkbox toggles and takes focus
	if got := valueChild(t, box).Content; got != "[x]" {
		t.Errorf("checkbox = %q, want \"[x]\"", got)
	}
	if !box.Focused {
		t.Error("checkbox should be focused after label click")
	}
}

func TestClickingForLabelTogglesReferencedControl(t *testing.T) {
	// Given — a label pointing at a checkbox elsewhere via for=id
	box := &layout.Input{Kind: layout.KindBox, Tag: "input", ID: "turbo",
		CursorCol: -1, CursorRow: -1, Attrs: map[string]string{"type": "checkbox"}}
	label := &layout.Input{Kind: layout.KindText, Tag: "label", Content: "Turbo mode",
		Attrs: map[string]string{"for": "turbo"}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{box, label},
	}}
	app := tui.TestApp(comp, 30, 4)

	// When — click the label text on row 1
	clickAt(app, 3, 1)

	// Then
	if got := valueChild(t, box).Content; got != "[x]" {
		t.Errorf("checkbox = %q, want \"[x]\"", got)
	}
}

func TestLabelWithoutControlDoesNothing(t *testing.T) {
	// Given
	label := &layout.Input{Kind: layout.KindText, Tag: "label", Content: "orphan"}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{label},
	}}
	app := tui.TestApp(comp, 30, 3)

	// When / Then — no panic, nothing to activate
	clickAt(app, 2, 0)
}
