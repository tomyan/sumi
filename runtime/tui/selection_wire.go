package tui

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// handleSelectionMouse extends or completes the global selection for
// left-drag motion and release. Press arming happens in the click
// branch (armSelectionPress) where the hit path is already computed.
// Motion and release still flow on to the app afterwards — controls
// with their own drag behaviour keep receiving events.
func handleSelectionMouse(app *App, evt input.Event) {
	if app.Selection == nil || evt.Kind != input.EventMouse || evt.Mouse.Button != input.ButtonLeft {
		return
	}
	switch evt.Mouse.Action {
	case input.MouseMotion:
		app.Selection.OnMotion(evt.Mouse.X, evt.Mouse.Y)
	case input.MouseRelease:
		if text := app.Selection.OnRelease(); text != "" && app.Clipboard != nil {
			app.Clipboard(text)
		}
	}
}

// armSelectionPress arms the selection on a left press, unless the
// press lands on an editable control, which owns its own drag
// selection — then only the existing global selection is cleared.
func armSelectionPress(app *App, path []*layout.Input, evt input.Event) {
	if app.Selection == nil {
		return
	}
	if n := deepestNode(path); n != nil && isEditableControl(n) {
		app.Selection.Clear()
		return
	}
	app.Selection.OnPress(evt.Mouse.X, evt.Mouse.Y)
}

func deepestNode(path []*layout.Input) *layout.Input {
	if len(path) == 0 {
		return nil
	}
	return path[len(path)-1]
}

func isEditableControl(n *layout.Input) bool {
	return n.Edit != nil || n.ContentEditable || n.Tag == "input" || n.Tag == "textarea"
}
