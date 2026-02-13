package tui

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
)

func TestAppStepDispatchesEvent(t *testing.T) {
	// Given
	var gotEvent input.Event
	app := &App{
		OnRender: func() {},
		OnEvent:  func(evt input.Event) { gotEvent = evt },
	}

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'a'})

	// Then
	if gotEvent.Rune != 'a' {
		t.Errorf("got rune %c, want 'a'", gotEvent.Rune)
	}
}

func TestAppStepRendersWhenDirty(t *testing.T) {
	// Given
	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
	}
	app.OnEvent = func(evt input.Event) { app.Dirty = true }

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then — one render from dirty
	if renderCount != 1 {
		t.Errorf("renderCount = %d, want 1", renderCount)
	}
}

func TestAppStepConverges(t *testing.T) {
	// Given — render always re-dirties (like $self measurement)
	renderCount := 0
	app := &App{}
	app.OnRender = func() {
		renderCount++
		app.Dirty = true
	}
	app.OnEvent = func(evt input.Event) { app.Dirty = true }

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then — bounded at 3
	if renderCount != 3 {
		t.Errorf("renderCount = %d, want 3 (bounded convergence)", renderCount)
	}
}

func TestAppStepNoRenderWhenClean(t *testing.T) {
	// Given
	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
		OnEvent:  func(evt input.Event) {}, // does not set dirty
	}

	// When
	app.Step(input.Event{Kind: input.EventKey, Rune: 'x'})

	// Then
	if renderCount != 0 {
		t.Errorf("renderCount = %d, want 0", renderCount)
	}
}

func TestAppRenderTriggersRender(t *testing.T) {
	// Given
	renderCount := 0
	app := &App{
		OnRender: func() { renderCount++ },
	}

	// When
	app.Render()

	// Then
	if renderCount != 1 {
		t.Errorf("renderCount = %d, want 1", renderCount)
	}
}

func TestAppRenderConverges(t *testing.T) {
	// Given
	renderCount := 0
	app := &App{}
	app.OnRender = func() {
		renderCount++
		app.Dirty = true
	}

	// When
	app.Render()

	// Then — bounded at 3
	if renderCount != 3 {
		t.Errorf("renderCount = %d, want 3 (bounded convergence)", renderCount)
	}
}

func TestAppTestBufferField(t *testing.T) {
	// Given
	buf := render.NewBuffer(10, 5)
	app := &App{
		TestBuffer: buf,
	}

	// Then
	if app.TestBuffer != buf {
		t.Error("TestBuffer not set correctly")
	}
}

func TestAppTestDimensionFields(t *testing.T) {
	// Given
	app := &App{
		TestWidth:  80,
		TestHeight: 24,
	}

	// Then
	if app.TestWidth != 80 {
		t.Errorf("TestWidth = %d, want 80", app.TestWidth)
	}
	if app.TestHeight != 24 {
		t.Errorf("TestHeight = %d, want 24", app.TestHeight)
	}
}

func TestAppStepInitializesChannels(t *testing.T) {
	// Given — app with no channels initialized
	app := &App{
		OnRender: func() {},
		OnEvent:  func(evt input.Event) {},
	}

	// When — Step should not panic even without initQuit
	app.Step(input.Event{Kind: input.EventKey, Rune: 'a'})

	// Then — Quit should not panic either
	app.Quit()
}
