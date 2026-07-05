package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func textareaApp() (*tui.Component, *layout.Input) {
	area := &layout.Input{Kind: layout.KindBox, Tag: "textarea",
		CursorCol: -1, CursorRow: -1}
	other := &layout.Input{Kind: layout.KindText, Tag: "button", Content: "Next"}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{area, other},
	}}
	return comp, area
}

func TestTextareaEnterInsertsNewline(t *testing.T) {
	// Given
	comp, area := textareaApp()
	app := tui.TestApp(comp, 30, 6)

	// When
	for _, r := range "ab" {
		app.Step(input.Event{Kind: input.EventKey, Rune: r})
	}
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'c'})

	// Then — multi-line value with a (row, col) cursor
	child := valueChild(t, area)
	if child.Content != "ab\nc" {
		t.Errorf("content = %q, want \"ab\\nc\"", child.Content)
	}
	if child.WhiteSpace != "pre" {
		t.Errorf("value child WhiteSpace = %q, want pre", child.WhiteSpace)
	}
	if area.CursorRow != 1 || area.CursorCol != 1 {
		t.Errorf("cursor = (%d,%d), want (1,1)", area.CursorCol, area.CursorRow)
	}
}

func TestTextareaArrowsMoveBetweenLines(t *testing.T) {
	// Given
	comp, area := textareaApp()
	app := tui.TestApp(comp, 30, 6)
	for _, r := range "ab" {
		app.Step(input.Event{Kind: input.EventKey, Rune: r})
	}
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyEnter})
	app.Step(input.Event{Kind: input.EventKey, Rune: 'c'})

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyUp})

	// Then
	if area.CursorRow != 0 || area.CursorCol != 1 {
		t.Errorf("cursor after Up = (%d,%d), want (1,0 row 0 col 1)", area.CursorCol, area.CursorRow)
	}
}

func TestTextareaTabStillCyclesFocus(t *testing.T) {
	// Given
	comp, area := textareaApp()
	app := tui.TestApp(comp, 30, 6)

	// When
	app.Step(input.Event{Kind: input.EventSpecial, Special: input.KeyTab})

	// Then — focus moved to the button; textarea cursor hidden
	if comp.FocusIndex != 1 {
		t.Errorf("FocusIndex = %d, want 1", comp.FocusIndex)
	}
	if area.CursorCol != -1 {
		t.Errorf("blurred textarea cursor = %d, want -1", area.CursorCol)
	}
}
