package tui

import (
	"github.com/tomyan/sumi/parser/style"
	"os"

	"github.com/tomyan/sumi/runtime/anim"
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// dispatchMouseScroll routes mouse scroll events to the scrollable container
// under the cursor. Returns true if a scroll was dispatched.
func dispatchMouseScroll(evt input.Event, comp *Component) bool {
	if evt.Kind != input.EventMouse || evt.Mouse.Action != input.MouseScroll {
		return false
	}
	if comp.LayoutResult == nil {
		return false
	}
	idx := layout.HitTestScroll(comp.LayoutResult, evt.Mouse.X, evt.Mouse.Y)
	if idx < 0 {
		return false
	}
	states := layout.CollectScrollStates(comp.Tree)
	if idx >= len(states) {
		return false
	}
	s := states[idx]
	if evt.Mouse.Shift {
		// Shift+scroll → horizontal.
		switch evt.Mouse.Button {
		case input.ScrollUp:
			s.ScrollLeft()
		case input.ScrollDown:
			s.ScrollRight(s.ContentHeight, s.ViewportHeight)
		}
	} else {
		// Normal scroll → vertical.
		switch evt.Mouse.Button {
		case input.ScrollUp:
			s.Follow = false
			s.ScrollUp()
		case input.ScrollDown:
			s.ScrollDown(s.ContentHeight, s.ViewportHeight)
			if s.AtBottom() {
				s.Follow = true
			}
		}
	}
	return true
}

// copy2D copies all cells from src to dst (which must have the same dimensions).
func copy2D(dst, src *render.Buffer) {
	for row := 0; row < src.Height(); row++ {
		for col := 0; col < src.Width(); col++ {
			c := src.Cell(row, col)
			dst.SetStyledCell(row, col, c.Ch, c.Style)
		}
	}
}

// Component represents a sumi component — a layout tree with event handling and lifecycle.
type Component struct {
	Tree         *layout.Input
	OnEvent      func(input.Event)
	AfterLayout  func() // called after layout to sync self-measurement signals
	Dispose      func()
	Dirty        bool                               // set by AfterLayout to request a re-render pass
	LayoutResult *layout.Box                        // set before AfterLayout with the latest layout result
	Keyframes    map[string]*anim.KeyframeAnimation // named keyframe animations from CSS
	Stylesheet   *style.Stylesheet                  // component CSS for runtime resolution
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
		layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
		tree := layout.Layout(comp.Tree, termW, termH)
		comp.LayoutResult = tree
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
		if comp.Dirty {
			comp.Dirty = false
			app.Dirty = true
		}
		buf := render.NewBuffer(termW, termH)
		layout.RenderTree(buf, tree, nil)
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		}
	}

	app.OnEvent = func(evt input.Event) {
		dispatchMouseScroll(evt, comp)
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			if h := layout.FindClickHandler(comp.Tree, comp.LayoutResult, evt.Mouse.X, evt.Mouse.Y); h != nil {
				h()
			}
		}
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
	OnPostRender func()            // called after each render
	OnResize     func()            // called on terminal resize
	SetApp       func(a *App)      // called with the app reference before Run
	ColorDepth   render.ColorDepth // emission depth; DepthAuto detects from env
}

// RunWithOptions runs a component with additional configuration.
func RunWithOptions(comp *Component, opts RunOptions) {
	if opts.ColorDepth == render.DepthAuto {
		render.SetColorDepth(term.DetectColorDepth())
	} else {
		render.SetColorDepth(opts.ColorDepth)
	}
	app := &App{}
	if opts.SetApp != nil {
		opts.SetApp(app)
	}

	engine := anim.NewEngine(anim.WallClock{}, func() { app.RequestFrame() })
	for name, kf := range comp.Keyframes {
		engine.RegisterKeyframes(name, kf)
	}

	screenBuf := render.NewBuffer(0, 0)
	frameBuf := render.NewBuffer(0, 0)
	var prevW, prevH int

	app.OnRender = func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		updateEnvSignals(termW, termH)
		// Update hover state on Input tree using previous layout positions.
		if comp.LayoutResult != nil {
			layout.UpdateHover(comp.Tree, comp.LayoutResult, app.mouseX, app.mouseY)
		}
		layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
		tree := layout.Layout(comp.Tree, termW, termH)
		comp.LayoutResult = tree
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
		if comp.Dirty {
			comp.Dirty = false
			app.Dirty = true
		}
		frameBuf.Resize(termW, termH)
		layout.RenderTreeWithEngine(frameBuf, tree, nil, engine)
		if engine.HasActive() {
			app.RequestFrame()
		}
		if termW != prevW || termH != prevH {
			// Resize: clear + full redraw in one buffered write.
			frameBuf.RenderWithClear(os.Stdout)
			prevW = termW
			prevH = termH
		} else {
			// Normal: diff against previous screen.
			frameBuf.RenderDiff(os.Stdout, screenBuf, termW, termH)
		}
	}

	app.OnEvent = func(evt input.Event) {
		dispatchMouseScroll(evt, comp)
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			if h := layout.FindClickHandler(comp.Tree, comp.LayoutResult, evt.Mouse.X, evt.Mouse.Y); h != nil {
				h()
			}
		}
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

	// Auto-enable mouse when any node has hover styles or click handlers.
	if layout.HasHoverStyles(comp.Tree) || layout.HasClickHandlers(comp.Tree) {
		app.HasMouse = true
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

	engine := anim.NewEngine(anim.WallClock{}, func() { app.RequestFrame() })
	for name, kf := range comp.Keyframes {
		engine.RegisterKeyframes(name, kf)
	}

	screenBuf := render.NewBuffer(0, 0)
	frameBuf := render.NewBuffer(0, 0)
	var prevW, prevH int

	app.OnRender = func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		updateEnvSignals(termW, termH)
		// Update hover state on Input tree using previous layout positions.
		if comp.LayoutResult != nil {
			layout.UpdateHover(comp.Tree, comp.LayoutResult, app.mouseX, app.mouseY)
		}
		layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
		tree := layout.Layout(comp.Tree, termW, termH)
		comp.LayoutResult = tree
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
		if comp.Dirty {
			comp.Dirty = false
			app.Dirty = true
		}
		frameBuf.Resize(termW, termH)
		layout.RenderTreeWithEngine(frameBuf, tree, nil, engine)
		if engine.HasActive() {
			app.RequestFrame()
		}
		if termW != prevW || termH != prevH {
			// Resize: clear + full redraw in one buffered write.
			frameBuf.RenderWithClear(os.Stdout)
			prevW = termW
			prevH = termH
		} else {
			// Normal: diff against previous screen.
			frameBuf.RenderDiff(os.Stdout, screenBuf, termW, termH)
		}
	}

	app.OnEvent = func(evt input.Event) {
		dispatchMouseScroll(evt, comp)
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			if h := layout.FindClickHandler(comp.Tree, comp.LayoutResult, evt.Mouse.X, evt.Mouse.Y); h != nil {
				h()
			}
		}
		if comp.OnEvent != nil {
			comp.OnEvent(evt)
		}
		app.Dirty = true
	}

	if layout.HasHoverStyles(comp.Tree) || layout.HasClickHandlers(comp.Tree) {
		app.HasMouse = true
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
