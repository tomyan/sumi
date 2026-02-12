package tui

import (
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

	// When — close the event channel immediately so the loop exits after initial render
	close(eventCh)
	app.runLoop(eventCh, nil, nil)

	// Then
	if !rendered {
		t.Error("OnRender was not called")
	}
}

func TestAppQuitsOnCtrlC(t *testing.T) {
	// Given
	eventCh := make(chan input.Event, 1)
	eventCh <- input.Event{Kind: input.EventKey, Rune: 3}

	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
	}

	// When
	app.runLoop(eventCh, nil, nil)

	// Then — should have rendered once (initial) then quit
	if renderCount != 1 {
		t.Errorf("renderCount = %d, want 1 (initial render only)", renderCount)
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

	// When
	app.runLoop(eventCh, nil, nil)

	// Then
	if !gotQ {
		t.Error("'q' event was not delivered to OnEvent")
	}
}

func TestAppQuitsOnSignal(t *testing.T) {
	// Given
	sigCh := make(chan struct{}, 1)
	sigCh <- struct{}{}
	eventCh := make(chan input.Event)

	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
	}

	// When
	app.runLoop(eventCh, nil, sigCh)

	// Then — only initial render, then signal causes quit
	if renderCount != 1 {
		t.Errorf("renderCount = %d, want 1", renderCount)
	}
}
