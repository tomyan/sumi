package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestContentEditableCursorRendersInverse(t *testing.T) {
	// Given a text node with contenteditable and cursor at position 2
	input := &Input{
		Kind:            KindText,
		Content:         "hello",
		ContentEditable: true,
		CursorCol:       2,
		CursorRow:       0,
	}

	// When
	box := Layout(input, 20, 5)
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then the cursor position should have inverse style
	cell := buf.Cell(0, 2)
	if cell.Ch != 'l' {
		t.Errorf("expected 'l' at cursor, got %c", cell.Ch)
	}
	if !cell.Style.Inverse {
		t.Error("expected inverse style at cursor position")
	}
	// Non-cursor positions should not be inverse
	if buf.Cell(0, 0).Style.Inverse {
		t.Error("expected non-inverse style at non-cursor position")
	}
}

func TestContentEditableCursorAtEndShowsBlock(t *testing.T) {
	// Given cursor past the text
	input := &Input{
		Kind:            KindText,
		Content:         "hi",
		ContentEditable: true,
		CursorCol:       2,
		CursorRow:       0,
	}

	// When
	box := Layout(input, 20, 5)
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then cursor shows inverse space
	cell := buf.Cell(0, 2)
	if cell.Ch != ' ' {
		t.Errorf("expected ' ' at cursor end, got %c", cell.Ch)
	}
	if !cell.Style.Inverse {
		t.Error("expected inverse style at cursor-past-end")
	}
}

func TestContentEditableCursorWraps(t *testing.T) {
	// Given text that wraps at width 10, cursor at offset 12
	// "hello world foo" wraps to ["hello", "world foo"] or similar
	input := &Input{
		Kind:            KindText,
		Content:         "hello world foo",
		ContentEditable: true,
		CursorCol:       12, // flat offset into "hello world foo"
		CursorRow:       0,
	}

	// When laid out with width 10 (wraps at 9 for cursor space)
	box := Layout(input, 10, 5)
	buf := render.NewBuffer(10, 5)
	RenderTree(buf, box, nil)

	// Then the cursor should be on a wrapped line, not row 0
	if box.CursorRow == 0 {
		t.Errorf("expected cursor on wrapped line, got row 0 (offset 12 in width 10)")
	}
	// The cursor cell should have inverse style
	cell := buf.Cell(box.CursorRow, box.CursorCol)
	if !cell.Style.Inverse {
		t.Errorf("expected inverse at cursor (%d, %d)", box.CursorRow, box.CursorCol)
	}
}

func TestContentEditableWrapsOneEarly(t *testing.T) {
	// Given text exactly filling the width
	input := &Input{
		Kind:            KindText,
		Content:         "1234567890", // 10 chars in width 10
		ContentEditable: true,
		CursorCol:       10, // cursor at end
		CursorRow:       0,
	}

	// When laid out with width 10 (wraps at 9)
	box := Layout(input, 10, 5)

	// Then text should wrap (9 chars per line + 1 on second line)
	if box.Height < 2 {
		t.Errorf("expected wrapping for 10 chars in width 10 (contenteditable reserves cursor column), got height %d", box.Height)
	}
}

func TestNonEditableTextNoCursor(t *testing.T) {
	// Given a normal text node with cursor fields set
	input := &Input{
		Kind:      KindText,
		Content:   "hello",
		CursorCol: 2,
		CursorRow: 0,
	}

	// When
	box := Layout(input, 20, 5)
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then no inverse cursor (contenteditable is false)
	if buf.Cell(0, 2).Style.Inverse {
		t.Error("non-editable text should not have inverse cursor")
	}
}
