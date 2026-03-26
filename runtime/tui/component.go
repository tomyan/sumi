package tui

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// Component represents a sumi component — a layout tree with event handling and lifecycle.
type Component struct {
	Tree    *layout.Input
	OnEvent func(input.Event)
	Dispose func()
}

// TestApp creates a test-mode App from a Component with the given viewport dimensions.
// Renders immediately. Use Step() to dispatch events and re-render.
func TestApp(comp *Component, w, h int) *App {
	app := &App{}
	app.TestWidth = w
	app.TestHeight = h
	app.TestBuffer = render.NewBuffer(w, h)

	app.OnRender = func() {
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		tree := layout.Layout(comp.Tree, termW, termH)
		buf := render.NewBuffer(termW, termH)
		layout.RenderTree(buf, tree, nil)
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		}
	}

	app.OnEvent = func(evt input.Event) {
		if comp.OnEvent != nil {
			comp.OnEvent(evt)
		}
		app.Dirty = true
	}

	app.componentDispose = comp.Dispose

	// Initial render.
	app.Render()
	return app
}

// Run runs a component as a full-screen terminal application.
func Run(comp *Component) {
	app := &App{}

	var prevTree *layout.Box
	var prevW, prevH int

	app.OnRender = func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		tree := layout.Layout(comp.Tree, termW, termH)
		changes, scrollChanged := layout.DiffTrees(prevTree, tree)
		if prevTree == nil || termW != prevW || termH != prevH || scrollChanged {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
			layout.ApplyChanges(os.Stdout, changes)
		}
		prevTree = tree
		prevW = termW
		prevH = termH
	}

	app.OnEvent = func(evt input.Event) {
		if comp.OnEvent != nil {
			comp.OnEvent(evt)
		}
		app.Dirty = true
	}

	app.componentDispose = comp.Dispose
	app.Run()

	// Cleanup on exit.
	if comp.Dispose != nil {
		comp.Dispose()
	}
}

// Cleanup disposes the component associated with this app (for testing).
func (a *App) Cleanup() {
	if a.componentDispose != nil {
		a.componentDispose()
	}
}
