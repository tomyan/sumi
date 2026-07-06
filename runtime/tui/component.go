package tui

import (
	"io"
	"os"

	"github.com/tomyan/sumi/parser/style"

	"github.com/tomyan/sumi/runtime/anim"
	"github.com/tomyan/sumi/runtime/css"
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
	FocusIndex   int                                // index into the focus scope's focusables of the focused element
	Keyframes    map[string]*anim.KeyframeAnimation // named keyframe animations from CSS
	Stylesheet   *style.Stylesheet                  // component CSS for runtime resolution

	lastFocusScope *layout.Input // focus-trap scope from the previous render (open dialog or tree root)
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
		syncFocus(comp)
		layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
		tree := layout.Layout(comp.Tree, termW, termH)
		if comp.Stylesheet != nil && comp.Stylesheet.HasContainerRules() {
			// Container queries need laid-out ancestor sizes: re-resolve
			// against the sizes just stamped and lay out again.
			layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
			tree = layout.Layout(comp.Tree, termW, termH)
		}
		comp.LayoutResult = tree
		if resyncInputElements(comp) {
			app.Dirty = true
		}
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
		if comp.Dirty {
			comp.Dirty = false
			app.Dirty = true
		}
		buf := render.NewBuffer(termW, termH)
		layout.RenderTree(buf, tree, nil)
		if app.Selection != nil {
			ApplySelectionOverlay(buf, app.Selection.Range())
		}
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		}
	}

	app.OnEvent = componentEventHandler(app, comp)
	app.Selection = NewSelectionController(func() *render.Buffer { return app.TestBuffer }, nil)

	app.componentDispose = comp.Dispose

	// Initial render.
	initFocus(comp)
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
	OnPostRender  func()            // called after each render
	OnResize      func()            // called on terminal resize
	SetApp        func(a *App)      // called with the app reference before Run
	ColorDepth    render.ColorDepth // emission depth; DepthAuto detects from env
	ExitOn        []string          // quit chords ("ctrl+c", "q", "escape"); nil = ctrl+c
	ReducedMotion bool              // prefers-reduced-motion: reduce (also via SUMI_REDUCED_MOTION env)
	ColorScheme   string            // "light" or "dark" forces the scheme (skips the OSC 11 probe); "" = detect
	Mouse         *bool             // override mouse-mode auto-detection (hover styles / click handlers)
	In            io.Reader         // terminal input stream (nil = os.Stdin)
	Out           io.Writer         // terminal output stream (nil = os.Stdout)
	OnLog         func(string)      // capture stdlib log lines while the app owns the terminal (nil = log untouched)
}

// RunWithOptions runs a component with additional configuration.
func RunWithOptions(comp *Component, opts RunOptions) {
	if opts.ColorDepth == render.DepthAuto {
		render.SetColorDepth(term.DetectColorDepth())
	} else {
		render.SetColorDepth(opts.ColorDepth)
	}
	css.SetReducedMotion(opts.ReducedMotion || os.Getenv("SUMI_REDUCED_MOTION") != "")
	if opts.ColorScheme != "" {
		scheme := render.SchemeDark
		if opts.ColorScheme == "light" {
			scheme = render.SchemeLight
		}
		render.SetColorScheme(scheme)
	}
	app := &App{}
	app.ExitOn = opts.ExitOn
	app.SchemeLocked = opts.ColorScheme != ""
	app.In = opts.In
	app.Out = opts.Out
	if opts.OnLog != nil {
		restore := captureLogs(opts.OnLog)
		defer restore()
	}
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
		termW, termH := app.terminalSize()
		updateEnvSignals(termW, termH)
		// Update hover state on Input tree using previous layout positions.
		if comp.LayoutResult != nil {
			layout.UpdateHover(comp.Tree, comp.LayoutResult, app.mouseX, app.mouseY)
		}
		syncFocus(comp)
		layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
		if stepLengthTransitions(comp.Tree, engine) {
			app.RequestFrame()
		}
		tree := layout.Layout(comp.Tree, termW, termH)
		if comp.Stylesheet != nil && comp.Stylesheet.HasContainerRules() {
			// Container queries need laid-out ancestor sizes: re-resolve
			// against the sizes just stamped and lay out again.
			layout.ResolveStyles(comp.Tree, comp.Stylesheet, termW, termH)
			tree = layout.Layout(comp.Tree, termW, termH)
		}
		comp.LayoutResult = tree
		if resyncInputElements(comp) {
			app.Dirty = true
		}
		if comp.AfterLayout != nil {
			comp.AfterLayout()
		}
		if comp.Dirty {
			comp.Dirty = false
			app.Dirty = true
		}
		frameBuf.Resize(termW, termH)
		layout.RenderTreeWithEngine(frameBuf, tree, nil, engine)
		if app.Selection != nil {
			ApplySelectionOverlay(frameBuf, app.Selection.Range())
		}
		if engine.HasActive() {
			app.RequestFrame()
		}
		if termW != prevW || termH != prevH || app.NeedsFullRedraw {
			// Resize or post-suspend: clear + full redraw in one write.
			app.NeedsFullRedraw = false
			frameBuf.RenderWithClear(app.out())
			prevW = termW
			prevH = termH
		} else {
			// Normal: diff against previous screen.
			frameBuf.RenderDiff(app.out(), screenBuf, termW, termH)
		}
	}

	app.OnEvent = componentEventHandler(app, comp)
	app.Selection = NewSelectionController(func() *render.Buffer { return frameBuf }, nil)
	app.Clipboard = app.systemClipboard

	if opts.OnPostRender != nil {
		app.OnPostRender = opts.OnPostRender
	}
	if opts.OnResize != nil {
		app.OnResize = opts.OnResize
	}

	// Mouse mode defaults on (global selection replaces the terminal's
	// native selection); an explicit option wins.
	app.HasMouse = true
	if opts.Mouse != nil {
		app.HasMouse = *opts.Mouse
	}

	app.componentDispose = comp.Dispose
	activeApp = app
	initFocus(comp)
	app.Run()
	activeApp = nil

	if comp.Dispose != nil {
		comp.Dispose()
	}
}

// Run runs a component as a full-screen terminal application.
func Run(comp *Component) {
	RunWithOptions(comp, RunOptions{})
}

// Cleanup disposes the component associated with this app (for testing).
func (a *App) Cleanup() {
	if a.componentDispose != nil {
		a.componentDispose()
	}
}
