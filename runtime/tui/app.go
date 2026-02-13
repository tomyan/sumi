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
	Dirty     bool              // set by handlers to trigger re-render
	quitCh    chan struct{}      // closed by Quit() to exit the event loop
	wakeCh    chan struct{}      // receives from RequestFrame() to wake the event loop
}

// Quit signals the event loop to exit.
func (a *App) Quit() {
	select {
	case a.quitCh <- struct{}{}:
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

// Run sets up the terminal, starts the event reader, and runs the event loop.
func (a *App) Run() {
	a.initQuit()

	if a.Title != "" || a.SaveTitle {
		fmt.Fprint(os.Stdout, "\033[22;2t") // save current title
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
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
	// Initial render with bounded convergence (e.g., $self measurement needs a second pass)
	a.Dirty = true
	for i := 0; i < 3 && a.Dirty; i++ {
		a.Dirty = false
		a.OnRender()
	}

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

		for i := 0; i < 3 && a.Dirty; i++ {
			a.Dirty = false
			a.OnRender()
		}

		if done {
			return
		}
	}
}

// dispatchEvent calls OnEvent if set.
func (a *App) dispatchEvent(evt input.Event) {
	if a.OnEvent != nil {
		a.OnEvent(evt)
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
