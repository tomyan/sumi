package tui

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestOnPostRenderCalledAfterInitialRender(t *testing.T) {
	// Given
	eventCh := make(chan input.Event)
	close(eventCh)

	postRenderCount := 0
	app := &App{
		OnRender:     func() {},
		OnPostRender: func() { postRenderCount++ },
	}
	app.initQuit()

	// When
	app.runLoop(eventCh, nil, nil)

	// Then — called once after initial converge
	if postRenderCount != 1 {
		t.Errorf("postRenderCount = %d, want 1", postRenderCount)
	}
}

func TestOnPostRenderCalledAfterEventConverge(t *testing.T) {
	// Given
	eventCh := make(chan input.Event, 2)
	eventCh <- input.Event{Kind: input.EventKey, Rune: 'a'}
	close(eventCh)

	postRenderCount := 0
	app := &App{
		OnRender:     func() {},
		OnPostRender: func() { postRenderCount++ },
	}
	app.initQuit()
	app.OnEvent = func(evt input.Event) { app.Dirty = true }

	// When
	app.runLoop(eventCh, nil, nil)

	// Then — once after initial converge + once after event converge
	if postRenderCount != 2 {
		t.Errorf("postRenderCount = %d, want 2", postRenderCount)
	}
}

func TestOnPostRenderNilDoesNotPanic(t *testing.T) {
	// Given — no OnPostRender set
	eventCh := make(chan input.Event)
	close(eventCh)

	app := &App{
		OnRender: func() {},
	}
	app.initQuit()

	// When — should not panic
	app.runLoop(eventCh, nil, nil)
}

func TestOnPostRenderCalledAfterStep(t *testing.T) {
	// Given
	postRenderCount := 0
	app := &App{
		OnRender:     func() {},
		OnPostRender: func() { postRenderCount++ },
	}
	app.OnEvent = func(evt input.Event) { app.Dirty = true }

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then
	if postRenderCount != 1 {
		t.Errorf("postRenderCount = %d, want 1", postRenderCount)
	}
}

func TestOnPostRenderCalledAfterRender(t *testing.T) {
	// Given
	postRenderCount := 0
	app := &App{
		OnRender:     func() {},
		OnPostRender: func() { postRenderCount++ },
	}

	// When
	app.Render()

	// Then
	if postRenderCount != 1 {
		t.Errorf("postRenderCount = %d, want 1", postRenderCount)
	}
}

func TestOnPostRenderNotCalledWhenClean(t *testing.T) {
	// Given — event doesn't set dirty, so no converge happens
	postRenderCount := 0
	app := &App{
		OnRender:     func() {},
		OnPostRender: func() { postRenderCount++ },
		OnEvent:      func(evt input.Event) {}, // does not set dirty
	}

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then — no converge means no post-render
	if postRenderCount != 0 {
		t.Errorf("postRenderCount = %d, want 0", postRenderCount)
	}
}
