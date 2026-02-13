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
