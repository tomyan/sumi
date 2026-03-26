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
	Tree        *layout.Input
	OnEvent     func(input.Event)
	AfterLayout func() // called after layout to sync self-measurement signals
	Dispose     func()
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
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
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

// activeApp holds the currently running app for Quit().
var activeApp *App

// Quit signals the running application to exit.
// Can be called from any component's event handler.
func Quit() {
	if activeApp != nil {
		activeApp.Quit()
	}
}

// RunOptions configures optional behaviors for Run.
type RunOptions struct {
	OnPostRender func()       // called after each render
	OnResize     func()       // called on terminal resize
	SetApp       func(a *App) // called with the app reference before Run
}

// RunWithOptions runs a component with additional configuration.
func RunWithOptions(comp *Component, opts RunOptions) {
	app := &App{}
	if opts.SetApp != nil {
		opts.SetApp(app)
	}

	var prevTree *layout.Box
	var prevW, prevH int

	app.OnRender = func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		updateEnvSignals(termW, termH)
		tree := layout.Layout(comp.Tree, termW, termH)
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
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

	if opts.OnPostRender != nil {
		app.OnPostRender = opts.OnPostRender
	}
	if opts.OnResize != nil {
		app.OnResize = opts.OnResize
	}

	app.componentDispose = comp.Dispose
	activeApp = app
	app.Run()
	activeApp = nil

	if comp.Dispose != nil {
		comp.Dispose()
	}
}

// Run runs a component as a full-screen terminal application.
func Run(comp *Component) {
	app := &App{}

	var prevTree *layout.Box
	var prevW, prevH int

	app.OnRender = func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		updateEnvSignals(termW, termH)
		tree := layout.Layout(comp.Tree, termW, termH)
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
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
	activeApp = app
	app.Run()
	activeApp = nil

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
