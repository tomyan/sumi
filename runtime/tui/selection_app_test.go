package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// D5c: selection wired into the app — drag paints inverse cells on the
// rendered frame, release copies via the injectable clipboard.

func mouseEvt(action input.MouseAction, btn input.MouseButton, x, y int) input.Event {
	return input.Event{Kind: input.EventMouse, Mouse: input.MouseEvent{Action: action, Button: btn, X: x, Y: y}}
}

func selectionApp(t *testing.T) (*tui.App, *[]string) {
	t.Helper()
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindText, Content: "hello world", CursorCol: -1, CursorRow: -1},
		},
	}}
	app := tui.TestApp(comp, 20, 3)
	var copies []string
	app.Clipboard = func(text string) { copies = append(copies, text) }
	return app, &copies
}

func TestAppDragPaintsSelectionOverlay(t *testing.T) {
	// Given
	app, _ := selectionApp(t)

	// When: press at (0,0), drag to (4,0).
	app.Step(mouseEvt(input.MousePress, input.ButtonLeft, 0, 0))
	app.Step(mouseEvt(input.MouseMotion, input.ButtonLeft, 4, 0))

	// Then: cells 0..4 inverse, cell 5 not.
	for col := 0; col <= 4; col++ {
		if !app.TestBuffer.Cell(0, col).Style.Inverse {
			t.Errorf("cell (0,%d) should be inverse", col)
		}
	}
	if app.TestBuffer.Cell(0, 5).Style.Inverse {
		t.Error("cell (0,5) should not be inverse")
	}
}

func TestAppReleaseCopiesSelection(t *testing.T) {
	// Given
	app, copies := selectionApp(t)

	// When
	app.Step(mouseEvt(input.MousePress, input.ButtonLeft, 0, 0))
	app.Step(mouseEvt(input.MouseMotion, input.ButtonLeft, 4, 0))
	app.Step(mouseEvt(input.MouseRelease, input.ButtonLeft, 4, 0))

	// Then
	if len(*copies) != 1 || (*copies)[0] != "hello" {
		t.Errorf("copies = %q, want [hello]", *copies)
	}
}

func TestAppFreshClickClearsSelection(t *testing.T) {
	// Given: an existing selection.
	app, _ := selectionApp(t)
	app.Step(mouseEvt(input.MousePress, input.ButtonLeft, 0, 0))
	app.Step(mouseEvt(input.MouseMotion, input.ButtonLeft, 4, 0))
	app.Step(mouseEvt(input.MouseRelease, input.ButtonLeft, 4, 0))

	// When: a later single click elsewhere. (Multi-click window is
	// time-based; a different cell resets the count regardless.)
	app.Step(mouseEvt(input.MousePress, input.ButtonLeft, 9, 0))

	// Then: overlay gone.
	if app.TestBuffer.Cell(0, 0).Style.Inverse {
		t.Error("selection should clear on a fresh click")
	}
}

func TestAppPressOnEditableControlSkipsSelection(t *testing.T) {
	// Given: an input element at the top of the tree.
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindBox, Tag: "input", CursorCol: -1, CursorRow: -1},
		},
	}}
	app := tui.TestApp(comp, 24, 3)

	// When: press + drag inside the input.
	app.Step(mouseEvt(input.MousePress, input.ButtonLeft, 2, 0))
	app.Step(mouseEvt(input.MouseMotion, input.ButtonLeft, 5, 0))

	// Then: no global selection overlay (the control owns its drag).
	if r := app.Selection.Range(); r != nil {
		t.Errorf("global selection = %+v, want nil over editable control", r)
	}
}
