package tui

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// D5a: SelectionController — screen-space cell selection driven by
// mouse gestures (svelterm's src/input/selection.ts is the reference).

func selBuffer(lines ...string) *render.Buffer {
	w := 0
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	buf := render.NewBuffer(w, len(lines))
	for row, l := range lines {
		for col, r := range l {
			buf.SetStyledCell(row, col, r, render.Style{})
		}
	}
	return buf
}

func newSel(buf *render.Buffer) *SelectionController {
	return NewSelectionController(func() *render.Buffer { return buf }, func() int64 { return 0 })
}

func TestSelectionClickWithoutDragSelectsNothing(t *testing.T) {
	// Given
	s := newSel(selBuffer("hello world"))

	// When: press and release on the same cell.
	s.OnPress(2, 0)
	text := s.OnRelease()

	// Then
	if s.Range() != nil {
		t.Errorf("range = %+v, want nil", s.Range())
	}
	if text != "" {
		t.Errorf("release text = %q, want empty", text)
	}
}

func TestSelectionDragSelectsAndExtracts(t *testing.T) {
	// Given
	s := newSel(selBuffer("hello world"))

	// When: drag from col 0 to col 4.
	s.OnPress(0, 0)
	moved := s.OnMotion(4, 0)
	text := s.OnRelease()

	// Then
	if !moved {
		t.Error("motion should report a change")
	}
	if text != "hello" {
		t.Errorf("text = %q, want %q", text, "hello")
	}
	if s.Range() == nil {
		t.Error("highlight should persist after release")
	}
}

func TestSelectionBackwardDragNormalizes(t *testing.T) {
	// Given
	s := newSel(selBuffer("hello world"))

	// When: drag right-to-left.
	s.OnPress(4, 0)
	s.OnMotion(0, 0)

	// Then
	r := s.Range()
	if r == nil || r.Start.Col != 0 || r.End.Col != 4 {
		t.Errorf("range = %+v, want cols 0..4", r)
	}
}

func TestSelectionMultiRowExtraction(t *testing.T) {
	// Given: trailing spaces on row 0 must be trimmed.
	s := newSel(selBuffer("ab    ", "cdef"))

	// When: drag from (1,0) to (2,1).
	s.OnPress(1, 0)
	s.OnMotion(2, 1)
	text := s.OnRelease()

	// Then
	if text != "b\ncde" {
		t.Errorf("text = %q, want %q", text, "b\ncde")
	}
}

func TestSelectionDoubleClickSelectsWord(t *testing.T) {
	// Given
	s := newSel(selBuffer("foo bar baz"))

	// When: two quick presses on "bar".
	s.OnPress(5, 0)
	s.OnRelease()
	s.OnPress(5, 0)

	// Then
	r := s.Range()
	if r == nil || r.Start.Col != 4 || r.End.Col != 6 {
		t.Errorf("range = %+v, want cols 4..6 (bar)", r)
	}
}

func TestSelectionTripleClickSelectsLine(t *testing.T) {
	// Given
	s := newSel(selBuffer("foo bar baz"))

	// When
	s.OnPress(5, 0)
	s.OnRelease()
	s.OnPress(5, 0)
	s.OnRelease()
	s.OnPress(5, 0)

	// Then
	r := s.Range()
	if r == nil || r.Start.Col != 0 || r.End.Col != 10 {
		t.Errorf("range = %+v, want cols 0..10", r)
	}
}

func TestSelectionSlowClicksDoNotAccumulate(t *testing.T) {
	// Given: a clock we control; clicks 500ms apart.
	now := int64(0)
	buf := selBuffer("foo bar baz")
	s := NewSelectionController(func() *render.Buffer { return buf }, func() int64 { return now })

	// When
	s.OnPress(5, 0)
	s.OnRelease()
	now = 500
	s.OnPress(5, 0)

	// Then: second press is a fresh single click — no word selection.
	if s.Range() != nil {
		t.Errorf("range = %+v, want nil", s.Range())
	}
}

func TestSelectionNewPressClearsOldSelection(t *testing.T) {
	// Given: an existing drag selection.
	now := int64(0)
	buf := selBuffer("hello world")
	s := NewSelectionController(func() *render.Buffer { return buf }, func() int64 { return now })
	s.OnPress(0, 0)
	s.OnMotion(4, 0)
	s.OnRelease()

	// When: a later single press elsewhere.
	now = 1000
	s.OnPress(8, 0)

	// Then
	if s.Range() != nil {
		t.Errorf("range = %+v, want nil after fresh press", s.Range())
	}
}

func TestSelectionMotionWithoutPressIgnored(t *testing.T) {
	// Given
	s := newSel(selBuffer("hello"))

	// When
	moved := s.OnMotion(3, 0)

	// Then
	if moved || s.Range() != nil {
		t.Error("motion without a press must not select")
	}
}

func TestSelectionDoubleClickOnWhitespaceDoesNothing(t *testing.T) {
	// Given
	s := newSel(selBuffer("foo bar"))

	// When: double-click the gap.
	s.OnPress(3, 0)
	s.OnRelease()
	s.OnPress(3, 0)

	// Then
	if s.Range() != nil {
		t.Errorf("range = %+v, want nil", s.Range())
	}
}
