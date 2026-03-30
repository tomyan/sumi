package edit

import "testing"

// --- Insert ---

func TestInsertAtEnd(t *testing.T) {
	s := &State{Value: "hello", Cursor: 5}
	s.Insert('!')
	if s.Value != "hello!" || s.Cursor != 6 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestInsertAtMiddle(t *testing.T) {
	s := &State{Value: "hllo", Cursor: 1}
	s.Insert('e')
	if s.Value != "hello" || s.Cursor != 2 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestInsertString(t *testing.T) {
	s := &State{Value: "hd", Cursor: 1}
	s.InsertString("ello worl")
	if s.Value != "hello world" || s.Cursor != 10 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

// --- Delete ---

func TestBackspace(t *testing.T) {
	s := &State{Value: "hello", Cursor: 5}
	s.Backspace()
	if s.Value != "hell" || s.Cursor != 4 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestBackspaceAtStartNoOp(t *testing.T) {
	s := &State{Value: "hello", Cursor: 0}
	s.Backspace()
	if s.Value != "hello" {
		t.Errorf("got %q", s.Value)
	}
}

func TestDelete(t *testing.T) {
	s := &State{Value: "hello", Cursor: 0}
	s.Delete()
	if s.Value != "ello" || s.Cursor != 0 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestDeleteAtEndNoOp(t *testing.T) {
	s := &State{Value: "hello", Cursor: 5}
	s.Delete()
	if s.Value != "hello" {
		t.Errorf("got %q", s.Value)
	}
}

// --- Navigation ---

func TestLeftRight(t *testing.T) {
	s := &State{Value: "hello", Cursor: 3}
	s.Left()
	if s.Cursor != 2 {
		t.Errorf("Left: cursor = %d", s.Cursor)
	}
	s.Right()
	s.Right()
	if s.Cursor != 4 {
		t.Errorf("Right: cursor = %d", s.Cursor)
	}
}

func TestHomeEnd(t *testing.T) {
	s := &State{Value: "hello", Cursor: 3}
	s.Home()
	if s.Cursor != 0 {
		t.Errorf("Home: cursor = %d", s.Cursor)
	}
	s.End()
	if s.Cursor != 5 {
		t.Errorf("End: cursor = %d", s.Cursor)
	}
}

func TestWordLeftRight(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 8}
	s.WordLeft()
	if s.Cursor != 6 {
		t.Errorf("WordLeft: cursor = %d, want 6", s.Cursor)
	}
	s.WordRight()
	if s.Cursor != 11 {
		t.Errorf("WordRight: cursor = %d, want 11", s.Cursor)
	}
}

// --- Kill + Kill Ring ---

func TestKillToEnd(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 5}
	s.KillToEnd()
	if s.Value != "hello" {
		t.Errorf("got %q", s.Value)
	}
	if len(s.killRing) != 1 || s.killRing[0] != " world" {
		t.Errorf("kill ring = %v", s.killRing)
	}
}

func TestKillToStart(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 5}
	s.KillToStart()
	if s.Value != " world" || s.Cursor != 0 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestKillWord(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 11}
	s.KillWord()
	if s.Value != "hello " || s.Cursor != 6 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestKillWordForward(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 0}
	s.KillWordForward()
	if s.Value != " world" || s.Cursor != 0 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestYank(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 5}
	s.KillToEnd() // kills " world", cursor at 5
	s.Home()      // cursor at 0
	s.Yank()      // yanks " world" at position 0
	if s.Value != " worldhello" || s.Cursor != 6 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestYankPop(t *testing.T) {
	s := &State{Value: "aaa bbb ccc", Cursor: 3}
	s.KillToEnd() // kills " bbb ccc"
	s.Value = "aaa"
	s.Cursor = 3
	s.KillToStart() // kills "aaa"
	// Kill ring: [" bbb ccc", "aaa"], idx=1

	s.Value = ""
	s.Cursor = 0
	s.Yank() // yanks "aaa" (most recent)
	if s.Value != "aaa" {
		t.Errorf("after Yank: %q", s.Value)
	}
	s.YankPop() // cycles to " bbb ccc"
	if s.Value != " bbb ccc" {
		t.Errorf("after YankPop: %q", s.Value)
	}
}

func TestYankPopRequiresYankFirst(t *testing.T) {
	s := &State{Value: "hello", Cursor: 5}
	s.KillToEnd() // kills "" — nothing
	s.killRing = []string{"test"}
	s.killRingIdx = 0
	s.YankPop() // should no-op (lastYank is false)
	if s.Value != "hello" {
		t.Errorf("got %q", s.Value)
	}
}

// --- Transpose ---

func TestTransposeChars(t *testing.T) {
	s := &State{Value: "abcd", Cursor: 2}
	s.TransposeChars()
	if s.Value != "acbd" || s.Cursor != 3 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestTransposeCharsAtEnd(t *testing.T) {
	s := &State{Value: "abcd", Cursor: 4}
	s.TransposeChars()
	if s.Value != "abdc" || s.Cursor != 4 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

// --- Word transforms ---

func TestUppercaseWord(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 0}
	s.UppercaseWord()
	if s.Value != "HELLO world" || s.Cursor != 5 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestLowercaseWord(t *testing.T) {
	s := &State{Value: "HELLO world", Cursor: 0}
	s.LowercaseWord()
	if s.Value != "hello world" || s.Cursor != 5 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

func TestCapitalizeWord(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 0}
	s.CapitalizeWord()
	if s.Value != "Hello world" || s.Cursor != 5 {
		t.Errorf("got %q cursor %d", s.Value, s.Cursor)
	}
}

// --- Undo / Redo ---

func TestUndoRedo(t *testing.T) {
	s := &State{}
	s.InsertString("hello")
	s.InsertString(" world")
	if s.Value != "hello world" {
		t.Fatalf("setup: %q", s.Value)
	}
	s.Undo()
	if s.Value != "hello" {
		t.Errorf("after first undo: %q, want %q", s.Value, "hello")
	}
	s.Undo()
	if s.Value != "" {
		t.Errorf("after second undo: %q, want %q", s.Value, "")
	}
	s.Redo()
	if s.Value != "hello" {
		t.Errorf("after redo: %q, want %q", s.Value, "hello")
	}
}

func TestUndoAfterKill(t *testing.T) {
	s := &State{Value: "hello world", Cursor: 5}
	s.KillToEnd()
	if s.Value != "hello" {
		t.Fatalf("after kill: %q", s.Value)
	}
	s.Undo()
	if s.Value != "hello world" || s.Cursor != 5 {
		t.Errorf("after undo: %q cursor %d", s.Value, s.Cursor)
	}
}

func TestNewEditClearsRedo(t *testing.T) {
	s := &State{}
	s.InsertString("hello")
	s.Undo()
	s.Insert('x')
	s.Redo() // should no-op — redo cleared by the insert
	if s.Value != "x" {
		t.Errorf("got %q, want %q", s.Value, "x")
	}
}

// --- History ---

func TestSubmitAndHistory(t *testing.T) {
	s := &State{}
	s.InsertString("first")
	s.Submit()
	s.InsertString("second")
	s.Submit()
	s.InsertString("current")

	s.HistoryUp()
	if s.Value != "second" {
		t.Errorf("first Up: %q", s.Value)
	}
	s.HistoryUp()
	if s.Value != "first" {
		t.Errorf("second Up: %q", s.Value)
	}
	s.HistoryDown()
	if s.Value != "second" {
		t.Errorf("first Down: %q", s.Value)
	}
	s.HistoryDown()
	if s.Value != "current" {
		t.Errorf("second Down: %q", s.Value)
	}
}

func TestHistoryPreservesEdits(t *testing.T) {
	s := &State{}
	s.InsertString("original")
	s.Submit()
	s.InsertString("current")

	// Go up, edit the history entry.
	s.HistoryUp()
	s.End()
	s.InsertString("-edited")
	if s.Value != "original-edited" {
		t.Fatalf("after edit: %q", s.Value)
	}

	// Go back to current.
	s.HistoryDown()
	if s.Value != "current" {
		t.Errorf("back to current: %q", s.Value)
	}

	// Go back up — edit should be preserved.
	s.HistoryUp()
	if s.Value != "original-edited" {
		t.Errorf("edit not preserved: %q, want %q", s.Value, "original-edited")
	}
}

func TestHistoryPreservesCursorPosition(t *testing.T) {
	s := &State{}
	s.InsertString("hello")
	s.Submit()
	s.InsertString("world")

	s.HistoryUp()
	s.Home() // cursor at 0
	s.HistoryDown()
	s.HistoryUp()
	if s.Cursor != 0 {
		t.Errorf("cursor not preserved: %d, want 0", s.Cursor)
	}
}

func TestSubmitEmptyNoHistory(t *testing.T) {
	s := &State{}
	s.Submit()
	s.InsertString("test")
	s.HistoryUp()
	if s.Value != "test" {
		t.Errorf("got %q", s.Value)
	}
}
