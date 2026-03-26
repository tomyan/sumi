package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestComponentTreeRendersToTestApp(t *testing.T) {
	// Given — a simple component with static text
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind:      layout.KindBox,
			CursorCol: -1,
			CursorRow: -1,
			Children: []*layout.Input{
				{Kind: layout.KindText, Content: "Hello"},
			},
		},
	}

	// When — create a test app and render
	app := tui.TestApp(comp, 20, 3)

	// Then — buffer should contain "Hello"
	buf := app.TestBuffer
	if buf == nil {
		t.Fatal("TestBuffer is nil")
	}
	if buf.Cell(0, 0).Ch != 'H' {
		t.Errorf("Cell(0,0) = %c, want 'H'", buf.Cell(0, 0).Ch)
	}
}

func TestComponentSignalUpdatesTree(t *testing.T) {
	// Given — a component with a signal-driven text node
	count := signal.New(0)
	textNode := &layout.Input{
		Kind:    layout.KindText,
		Content: "Count: 0",
	}
	signal.Effect(func() {
		textNode.Content = "Count: " + itoa(count.Get())
	})

	comp := &tui.Component{
		Tree: &layout.Input{
			Kind:      layout.KindBox,
			CursorCol: -1,
			CursorRow: -1,
			Children:  []*layout.Input{textNode},
		},
	}

	app := tui.TestApp(comp, 20, 3)

	// When — update the signal and re-render
	count.Set(5)
	app.Render()

	// Then — buffer should reflect the new value
	row := extractRow(app.TestBuffer, 0, 8)
	if row != "Count: 5" {
		t.Errorf("row = %q, want %q", row, "Count: 5")
	}
}

func TestComponentOnEventDispatches(t *testing.T) {
	// Given — a component with an event handler
	count := signal.New(0)
	textNode := &layout.Input{
		Kind:    layout.KindText,
		Content: "0",
	}
	signal.Effect(func() {
		textNode.Content = itoa(count.Get())
	})

	comp := &tui.Component{
		Tree: &layout.Input{
			Kind:      layout.KindBox,
			CursorCol: -1,
			CursorRow: -1,
			Children:  []*layout.Input{textNode},
		},
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventKey && evt.Rune == '+' {
				count.Update(func(n int) int { return n + 1 })
			}
		},
	}

	app := tui.TestApp(comp, 20, 3)

	// When — step with a key event
	app.Step(input.Event{Kind: input.EventKey, Rune: '+'})

	// Then
	row := extractRow(app.TestBuffer, 0, 1)
	if row != "1" {
		t.Errorf("row = %q, want %q", row, "1")
	}
}

func TestComponentDispose(t *testing.T) {
	// Given
	disposed := false
	comp := &tui.Component{
		Tree: &layout.Input{
			Kind:      layout.KindBox,
			CursorCol: -1,
			CursorRow: -1,
		},
		Dispose: func() { disposed = true },
	}

	// When
	app := tui.TestApp(comp, 10, 3)
	app.Cleanup()

	// Then
	if !disposed {
		t.Error("Dispose was not called")
	}
}

// helpers

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func extractRow(buf *render.Buffer, row, n int) string {
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		c := buf.Cell(row, i)
		if c.Ch == 0 {
			s[i] = ' '
		} else {
			s[i] = byte(c.Ch)
		}
	}
	return string(s)
}
