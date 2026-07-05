package edit

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestReadOnlyAllowsNavigationOnly(t *testing.T) {
	// Given
	s := &State{Value: "abc", Cursor: 3}
	c := Constraints{ReadOnly: true}

	// When / Then — navigation works
	if !HandleKeyWith(s, input.Event{Kind: input.EventSpecial, Special: input.KeyLeft}, c) || s.Cursor != 2 {
		t.Errorf("Left in readonly: cursor %d, want 2", s.Cursor)
	}

	// When / Then — edits are consumed but do nothing
	for _, evt := range []input.Event{
		{Kind: input.EventKey, Rune: 'x'},
		{Kind: input.EventSpecial, Special: input.KeyBackspace},
		{Kind: input.EventPaste, PasteText: "zz"},
		{Kind: input.EventKey, Rune: 'k', Ctrl: true},
	} {
		if !HandleKeyWith(s, evt, c) {
			t.Errorf("readonly should consume %+v", evt)
		}
	}
	if s.Value != "abc" {
		t.Errorf("readonly value changed: %q", s.Value)
	}
}

func TestMaxLengthCapsTyping(t *testing.T) {
	// Given
	s := &State{Value: "abcd", Cursor: 4}
	c := Constraints{MaxLength: 5}

	// When
	HandleKeyWith(s, input.Event{Kind: input.EventKey, Rune: 'e'}, c)
	HandleKeyWith(s, input.Event{Kind: input.EventKey, Rune: 'f'}, c)

	// Then
	if s.Value != "abcde" {
		t.Errorf("value = %q, want capped \"abcde\"", s.Value)
	}
}

func TestMaxLengthTruncatesPaste(t *testing.T) {
	// Given
	s := &State{Value: "ab", Cursor: 2}
	c := Constraints{MaxLength: 5}

	// When
	HandleKeyWith(s, input.Event{Kind: input.EventPaste, PasteText: "cdefgh"}, c)

	// Then
	if s.Value != "abcde" {
		t.Errorf("value = %q, want truncated \"abcde\"", s.Value)
	}
}
