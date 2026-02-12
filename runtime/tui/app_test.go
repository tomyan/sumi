package tui

import (
	"os"
	"syscall"
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestAppCallsOnRender(t *testing.T) {
	// Given
	rendered := false
	eventCh := make(chan input.Event)

	app := &App{
		OnRender: func() { rendered = true },
	}
	app.initQuit()

	// When — close the event channel immediately so the loop exits after initial render
	close(eventCh)
	app.runLoop(eventCh, nil, nil)

	// Then
	if !rendered {
		t.Error("OnRender was not called")
	}
}

func TestAppDispatchesCtrlCAsEvent(t *testing.T) {
	// Given — Ctrl+C should be dispatched as a regular key event, not auto-quit
	eventCh := make(chan input.Event, 2)
	eventCh <- input.Event{Kind: input.EventKey, Rune: 3}
	close(eventCh)

	var gotCtrlC bool
	app := &App{
		OnRender: func() {},
	}
	app.initQuit()
	app.OnEvent = func(evt input.Event) {
		if evt.Kind == input.EventKey && evt.Rune == 3 {
			gotCtrlC = true
		}
	}

	// When
	app.runLoop(eventCh, nil, nil)

	// Then — Ctrl+C dispatched to OnEvent, loop exits when channel closes
	if !gotCtrlC {
		t.Error("Ctrl+C was not dispatched to OnEvent")
	}
}

func TestAppQuitExitsLoop(t *testing.T) {
	// Given — a long-lived event channel
	eventCh := make(chan input.Event, 1)
	eventCh <- input.Event{Kind: input.EventKey, Rune: 'a'}

	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
	}
	app.initQuit()
	app.OnEvent = func(evt input.Event) {
		app.Dirty = true
		app.Quit()
	}

	// When
	app.runLoop(eventCh, nil, nil)

	// Then — initial render + dirty render after event, then quit
	if renderCount != 2 {
		t.Errorf("renderCount = %d, want 2", renderCount)
	}
}

func TestAppDispatchesSignalAsEvent(t *testing.T) {
	// Given — a signal arrives on sigCh
	sigCh := make(chan os.Signal, 1)
	sigCh <- syscall.SIGINT
	eventCh := make(chan input.Event)

	var gotSignal bool
	var gotSignalKind input.EventKind
	var gotSignalValue syscall.Signal
	app := &App{
		OnRender: func() {},
	}
	app.initQuit()
	app.OnEvent = func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			gotSignal = true
			gotSignalKind = evt.Kind
			gotSignalValue = evt.Signal
			app.Quit()
		}
	}

	// When
	app.runLoop(eventCh, nil, sigCh)

	// Then — signal dispatched as EventSignal
	if !gotSignal {
		t.Fatal("signal was not dispatched to OnEvent")
	}
	if gotSignalKind != input.EventSignal {
		t.Errorf("Kind = %d, want EventSignal", gotSignalKind)
	}
	if gotSignalValue != syscall.SIGINT {
		t.Errorf("Signal = %v, want SIGINT", gotSignalValue)
	}
}

func TestAppDrainsEvents(t *testing.T) {
	// Given — queue 3 events before loop starts, then close
	eventCh := make(chan input.Event, 4)
	eventCh <- input.Event{Kind: input.EventKey, Rune: 'a'}
	eventCh <- input.Event{Kind: input.EventKey, Rune: 'b'}
	eventCh <- input.Event{Kind: input.EventKey, Rune: 'c'}
	close(eventCh)

	var receivedRunes []rune
	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
	}
	app.initQuit()
	app.OnEvent = func(evt input.Event) {
		receivedRunes = append(receivedRunes, evt.Rune)
		app.Dirty = true
	}

	// When
	app.runLoop(eventCh, nil, nil)

	// Then — all 3 events delivered
	if len(receivedRunes) != 3 {
		t.Fatalf("received %d events, want 3", len(receivedRunes))
	}
	// Batching: initial render + one render after draining all 3 events = 2 renders
	if renderCount != 2 {
		t.Errorf("renderCount = %d, want 2 (initial + one batch)", renderCount)
	}
}

func TestAppHandlesResize(t *testing.T) {
	// Given
	resizeCh := make(chan struct{}, 1)
	resizeCh <- struct{}{}
	eventCh := make(chan input.Event)

	resized := false
	renderCount := 0
	app := &App{
		OnRender: func() {
			renderCount++
			// Close eventCh after resize render to stop the loop
			if renderCount == 2 {
				close(eventCh)
			}
		},
		OnResize: func() { resized = true },
	}
	app.initQuit()

	// When
	app.runLoop(eventCh, resizeCh, nil)

	// Then
	if !resized {
		t.Error("OnResize was not called")
	}
	if renderCount != 2 {
		t.Errorf("renderCount = %d, want 2 (initial + resize)", renderCount)
	}
}

func TestAppNoQuitOnQ(t *testing.T) {
	// Given — 'q' should be delivered as a normal key, not cause quit
	eventCh := make(chan input.Event, 2)
	eventCh <- input.Event{Kind: input.EventKey, Rune: 'q'}
	close(eventCh)

	gotQ := false
	app := &App{
		OnRender: func() {},
		OnEvent: func(evt input.Event) {
			if evt.Rune == 'q' {
				gotQ = true
			}
		},
	}
	app.initQuit()

	// When
	app.runLoop(eventCh, nil, nil)

	// Then
	if !gotQ {
		t.Error("'q' event was not delivered to OnEvent")
	}
}

func TestAppSignalWithNoHandlerDoesNotCrash(t *testing.T) {
	// Given — signal arrives but no OnEvent handler is set
	sigCh := make(chan os.Signal, 1)
	sigCh <- syscall.SIGTERM
	eventCh := make(chan input.Event, 1)
	close(eventCh)

	app := &App{
		OnRender: func() {},
	}
	app.initQuit()

	// When — signal dispatched (no-op), then eventCh closes, loop exits
	app.runLoop(eventCh, nil, sigCh)

	// Then — no panic, no hang
}
