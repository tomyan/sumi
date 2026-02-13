package sumitest

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

// Harness wraps a tui.App for ergonomic component testing.
type Harness struct {
	app *tui.App
}

// New creates a Harness wrapping the given App.
func New(app *tui.App) *Harness {
	return &Harness{app: app}
}

// Step dispatches a single event and runs convergence rendering.
func (h *Harness) Step(evt input.Event) {
	h.app.Step(evt)
}

// Text returns the current buffer content as plain text.
func (h *Harness) Text() string {
	if h.app.TestBuffer == nil {
		return ""
	}
	return h.app.TestBuffer.ToPlainText()
}

// Resize updates test dimensions, calls OnResize if set, and triggers a render.
func (h *Harness) Resize(w, h2 int) {
	h.app.TestWidth = w
	h.app.TestHeight = h2
	if h.app.OnResize != nil {
		h.app.OnResize()
	}
	h.app.Render()
}

// Buffer returns the underlying render buffer for direct access.
func (h *Harness) Buffer() *render.Buffer {
	return h.app.TestBuffer
}

// StyledText returns the current buffer content as styled markup.
func (h *Harness) StyledText() string {
	if h.app.TestBuffer == nil {
		return ""
	}
	return h.app.TestBuffer.ToStyledText()
}
