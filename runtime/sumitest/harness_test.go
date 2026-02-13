package sumitest

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestNewHarness(t *testing.T) {
	// Given
	app := &tui.App{
		TestWidth:  10,
		TestHeight: 5,
		TestBuffer: render.NewBuffer(10, 5),
		OnRender:   func() {},
		OnEvent:    func(evt input.Event) {},
	}

	// When
	h := New(app)

	// Then
	if h == nil {
		t.Fatal("New returned nil")
	}
}

func TestHarnessText(t *testing.T) {
	// Given
	buf := render.NewBuffer(10, 1)
	buf.WriteText(0, 0, "Hello")
	app := &tui.App{
		TestBuffer: buf,
		OnRender:   func() {},
	}
	h := New(app)

	// When
	text := h.Text()

	// Then
	if text != "Hello" {
		t.Errorf("Text() = %q, want %q", text, "Hello")
	}
}

func TestHarnessStyledText(t *testing.T) {
	// Given
	buf := render.NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Hi", render.Style{FG: render.Color{Name: "green"}})
	app := &tui.App{
		TestBuffer: buf,
		OnRender:   func() {},
	}
	h := New(app)

	// When
	styled := h.StyledText()

	// Then
	want := "<<green>>Hi<</>>"
	if styled != want {
		t.Errorf("StyledText() = %q, want %q", styled, want)
	}
}

func TestHarnessStep(t *testing.T) {
	// Given
	var gotEvent bool
	app := &tui.App{
		TestWidth:  10,
		TestHeight: 5,
		TestBuffer: render.NewBuffer(10, 5),
		OnRender:   func() {},
		OnEvent:    func(evt input.Event) { gotEvent = true },
	}
	h := New(app)

	// When
	h.Step(input.Event{Kind: input.EventKey, Rune: 'a'})

	// Then
	if !gotEvent {
		t.Error("Step did not dispatch event")
	}
}

func TestHarnessResizeUpdatesDimensions(t *testing.T) {
	// Given
	app := &tui.App{
		TestWidth:  10,
		TestHeight: 5,
		TestBuffer: render.NewBuffer(10, 5),
		OnRender:   func() {},
	}
	h := New(app)

	// When
	h.Resize(20, 10)

	// Then
	if app.TestWidth != 20 {
		t.Errorf("TestWidth = %d, want 20", app.TestWidth)
	}
	if app.TestHeight != 10 {
		t.Errorf("TestHeight = %d, want 10", app.TestHeight)
	}
}

func TestHarnessResizeCallsOnResize(t *testing.T) {
	// Given
	var resized bool
	app := &tui.App{
		TestWidth:  10,
		TestHeight: 5,
		TestBuffer: render.NewBuffer(10, 5),
		OnRender:   func() {},
		OnResize:   func() { resized = true },
	}
	h := New(app)

	// When
	h.Resize(20, 10)

	// Then
	if !resized {
		t.Error("OnResize was not called")
	}
}

func TestHarnessResizeTriggersRender(t *testing.T) {
	// Given
	renderCount := 0
	app := &tui.App{
		TestWidth:  10,
		TestHeight: 5,
		TestBuffer: render.NewBuffer(10, 5),
	}
	app.OnRender = func() { renderCount++ }
	h := New(app)

	// When
	h.Resize(20, 10)

	// Then — at least one render after resize
	if renderCount < 1 {
		t.Errorf("renderCount = %d, want at least 1", renderCount)
	}
}
