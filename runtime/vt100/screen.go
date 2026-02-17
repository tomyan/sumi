package vt100

import "github.com/tomyan/sumi/runtime/render"

// Screen is a virtual terminal screen backed by a render.Buffer.
// It tracks cursor position and current style state, updated by
// feeding ANSI byte streams via the Write method.
type Screen struct {
	buf      *render.Buffer
	curRow   int
	curCol   int
	style    render.Style
	sentinel bool // true after \x1b]999;done\x07 is seen
}

// NewScreen creates a screen with the given dimensions.
func NewScreen(width, height int) *Screen {
	return &Screen{buf: render.NewBuffer(width, height)}
}

// Cell returns the cell at (row, col).
func (s *Screen) Cell(row, col int) render.Cell {
	return s.buf.Cell(row, col)
}

// Width returns the screen width.
func (s *Screen) Width() int { return s.buf.Width() }

// Height returns the screen height.
func (s *Screen) Height() int { return s.buf.Height() }

// Buffer returns the underlying render buffer.
func (s *Screen) Buffer() *render.Buffer { return s.buf }

// SentinelSeen returns true if the frame sentinel was detected.
func (s *Screen) SentinelSeen() bool { return s.sentinel }

// ResetSentinel clears the sentinel flag for the next frame.
func (s *Screen) ResetSentinel() { s.sentinel = false }

// Resize creates a new buffer with the given dimensions, resetting cursor and style.
func (s *Screen) Resize(width, height int) {
	s.buf = render.NewBuffer(width, height)
	s.curRow = 0
	s.curCol = 0
	s.style = render.Style{}
}
