package edit

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestLineColLocatesCursor(t *testing.T) {
	// Given
	s := &State{Value: "ab\ncdef\ng", Cursor: 5} // after "cd" on line 1

	// When
	row, col := s.LineCol()

	// Then
	if row != 1 || col != 2 {
		t.Errorf("LineCol = (%d,%d), want (1,2)", row, col)
	}
}

func TestCursorUpAndDownPreserveColumnAndClamp(t *testing.T) {
	// Given — cursor at col 3 on the long middle line
	s := &State{Value: "ab\ncdef\ng", Cursor: 6} // after "cde"

	// When — up to a shorter line clamps the column
	s.CursorUp()

	// Then — line 0 is "ab": col clamps to 2 (end of line)
	if row, col := s.LineCol(); row != 0 || col != 2 {
		t.Errorf("after up: (%d,%d), want (0,2)", row, col)
	}

	// When — down twice lands on the last short line
	s.CursorDown()
	s.CursorDown()

	// Then
	if row, col := s.LineCol(); row != 2 || col != 1 {
		t.Errorf("after downs: (%d,%d), want (2,1)", row, col)
	}

	// When — down at the last line stays put
	s.CursorDown()
	if row, _ := s.LineCol(); row != 2 {
		t.Errorf("down past end moved to row %d", row)
	}
}

func TestMultilineConstraintHandlesEnterAndArrows(t *testing.T) {
	// Given
	s := &State{Value: "hi", Cursor: 2}
	c := Constraints{Multiline: true}

	// When
	if !HandleKeyWith(s, input.Event{Kind: input.EventSpecial, Special: input.KeyEnter}, c) {
		t.Fatal("Enter not handled in multiline mode")
	}
	HandleKeyWith(s, input.Event{Kind: input.EventKey, Rune: 'x'}, c)

	// Then
	if s.Value != "hi\nx" {
		t.Errorf("value = %q, want \"hi\\nx\"", s.Value)
	}

	// When — Up moves back to line 0
	if !HandleKeyWith(s, input.Event{Kind: input.EventSpecial, Special: input.KeyUp}, c) {
		t.Fatal("Up not handled in multiline mode")
	}
	if row, _ := s.LineCol(); row != 0 {
		t.Errorf("after Up: row %d, want 0", row)
	}
}

func TestMultilineReadonlyBlocksEnter(t *testing.T) {
	// Given
	s := &State{Value: "hi", Cursor: 2}
	c := Constraints{Multiline: true, ReadOnly: true}

	// When — Enter consumed but no newline; Up still navigates
	if !HandleKeyWith(s, input.Event{Kind: input.EventSpecial, Special: input.KeyEnter}, c) {
		t.Fatal("Enter should be consumed in readonly multiline")
	}
	if s.Value != "hi" {
		t.Errorf("readonly value changed: %q", s.Value)
	}
}
