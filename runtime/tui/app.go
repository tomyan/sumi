package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// App owns the terminal event loop for a Sumi application.
// Generated code builds the tree, defines handlers, and calls app.Run().
type App struct {
	OnRender  func()            // called to perform a render
	OnEvent   func(input.Event) // called for each input event
	OnResize  func()            // called on terminal resize
	HasMouse  bool              // enable SGR mouse mode
	Title     string            // static terminal title (saved/set/restored around Run)
	SaveTitle bool              // save/restore terminal title only (for dynamic titles set in doRender)
	Dirty        bool              // set by handlers to trigger re-render
	OnPostRender func()            // called after each converge() cycle (if non-nil)
	quitCh       chan struct{}     // closed by Quit() to exit the event loop
	wakeCh       chan struct{}     // receives from RequestFrame() to wake the event loop

	// Test mode fields — set by CreateApp for synchronous stepping.
	TestWidth  int            // test viewport width (0 = use real terminal)
	TestHeight int            // test viewport height (0 = use real terminal)
	TestBuffer *render.Buffer // populated by doRender in test mode instead of writing to stdout

	componentDispose func() // component cleanup, called by Cleanup()
}

// Quit signals the event loop to exit.
func (a *App) Quit() {
	select {
	case a.quitCh <- struct{}{}:
	default:
	}
}

// Wake immediately wakes the event loop to trigger a re-render.
// Used by background goroutines (e.g. editor PTY readers) to signal new content.
// Safe to call before Run() — the wake will be picked up once the loop starts.
func (a *App) Wake() {
	if a.wakeCh == nil {
		a.Dirty = true
		return
	}
	a.Dirty = true
	select {
	case a.wakeCh <- struct{}{}:
	default:
	}
}

// RequestFrame schedules an animation frame tick after ~16ms.
// When the tick fires, the event loop dispatches EventFrame to OnEvent.
func (a *App) RequestFrame() {
	go func() {
		time.Sleep(16 * time.Millisecond)
		a.Dirty = true
		select {
		case a.wakeCh <- struct{}{}:
		default:
		}
	}()
}

// initQuit creates the quit and wake channels. Called by Run() and tests.
func (a *App) initQuit() {
	a.quitCh = make(chan struct{}, 1)
	a.wakeCh = make(chan struct{}, 1)
}

// Step dispatches a single event and runs bounded convergence rendering.
// Used for synchronous testing — no terminal setup, no goroutines.
func (a *App) Step(evt input.Event) {
	a.ensureInit()
	a.dispatchEvent(evt)
	a.converge()
}

// Render triggers an initial render with bounded convergence.
// Used for synchronous testing to produce the first rendered frame.
func (a *App) Render() {
	a.ensureInit()
	a.Dirty = true
	a.converge()
}

// converge runs up to 3 render passes while Dirty remains true,
// then calls OnPostRender if any rendering occurred.
func (a *App) converge() {
	rendered := false
	for i := 0; i < 3 && a.Dirty; i++ {
		a.Dirty = false
		a.OnRender()
		rendered = true
	}
	if rendered {
		a.postRender()
	}
}

// postRender calls OnPostRender if set.
func (a *App) postRender() {
	if a.OnPostRender != nil {
		a.OnPostRender()
	}
}

// ensureInit creates channels if not already initialized.
func (a *App) ensureInit() {
	if a.quitCh == nil {
		a.initQuit()
	}
}

// Run sets up the terminal, starts the event reader, and runs the event loop.
func (a *App) Run() {
	a.initQuit()

	if a.Title != "" || a.SaveTitle {
		fmt.Fprint(os.Stdout, "\033[22;2t") // save current title
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer func() {
		if restore != nil {
			restore()
		}
	}()
	render.EnterAlternateScreen(os.Stdout)
	defer render.ExitAlternateScreen(os.Stdout)

	fmt.Fprint(os.Stdout, input.PasteEnableSeq)
	defer fmt.Fprint(os.Stdout, input.PasteDisableSeq)

	if a.HasMouse {
		fmt.Fprint(os.Stdout, input.MouseEnableSeq)
		defer fmt.Fprint(os.Stdout, input.MouseDisableSeq)
	}

	if a.Title != "" || a.SaveTitle {
		defer fmt.Fprint(os.Stdout, "\033[23;2t") // restore title
	}
	if a.Title != "" {
		fmt.Fprintf(os.Stdout, "\033]2;%s\007", a.Title)
	}

	eventCh := make(chan input.Event, 64)
	go func() {
		for {
			evt, err := input.ReadEvent(os.Stdin)
			if err != nil {
				close(eventCh)
				return
			}
			eventCh <- evt
		}
	}()

	resizeCh, stopResize := term.WatchResize()
	defer stopResize()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	a.runLoop(eventCh, resizeCh, sigCh)
}

// runLoop is the internal event loop, injectable for testing.
// eventCh delivers input events, resizeCh delivers resize signals,
// sigCh delivers OS signals (dispatched as EventSignal events).
func (a *App) runLoop(eventCh <-chan input.Event, resizeCh <-chan struct{}, sigCh <-chan os.Signal) {
	// Initial render with bounded convergence (e.g., self-measurement signals may need a second pass)
	a.Dirty = true
	a.converge()

	for {
		// Wait for at least one event
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return
			}
			a.dispatchEvent(evt)

		case <-resizeCh:
			a.dispatchResize()

		case sig := <-sigCh:
			a.dispatchEvent(input.Event{Kind: input.EventSignal, Signal: sig.(syscall.Signal)})

		case <-a.wakeCh:
			a.dispatchEvent(input.Event{Kind: input.EventFrame})

		case <-a.quitCh:
			return
		}

		// Drain pending events before rendering
		done := a.drain(eventCh, resizeCh, sigCh)

		a.converge()

		if done {
			return
		}
	}
}

// dispatchEvent calls OnEvent if set.
// If no OnEvent handler is set, SIGINT/SIGTERM signals quit the app by default.
func (a *App) dispatchEvent(evt input.Event) {
	if a.OnEvent != nil {
		a.OnEvent(evt)
		return
	}
	// Default: quit on signal, Ctrl+C, q, or Enter when no handler is set.
	if evt.Kind == input.EventSignal {
		a.Quit()
	} else if evt.Kind == input.EventKey && (evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c')) {
		a.Quit()
	} else if evt.Kind == input.EventSpecial && evt.Special == input.KeyEnter {
		a.Quit()
	}
}

// dispatchResize calls OnResize if set and marks dirty.
func (a *App) dispatchResize() {
	if a.OnResize != nil {
		a.OnResize()
	}
	a.Dirty = true
}

// drain non-blocking reads all pending events from all channels.
// Returns true if a quit was requested or the event channel closed.
func (a *App) drain(eventCh <-chan input.Event, resizeCh <-chan struct{}, sigCh <-chan os.Signal) bool {
	for {
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return true
			}
			a.dispatchEvent(evt)

		case <-resizeCh:
			a.dispatchResize()

		case sig := <-sigCh:
			a.dispatchEvent(input.Event{Kind: input.EventSignal, Signal: sig.(syscall.Signal)})

		case <-a.wakeCh:
			a.dispatchEvent(input.Event{Kind: input.EventFrame})

		case <-a.quitCh:
			return true

		default:
			return false
		}
	}
}
