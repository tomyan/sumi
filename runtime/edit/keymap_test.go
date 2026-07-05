package edit

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestHandleKeyTypesAndEdits(t *testing.T) {
	// Given
	s := &State{}

	// When — type, move, edit through the key mapping
	steps := []input.Event{
		{Kind: input.EventKey, Rune: 'h'},
		{Kind: input.EventKey, Rune: 'i'},
		{Kind: input.EventSpecial, Special: input.KeyLeft},
		{Kind: input.EventKey, Rune: 'e'},
		{Kind: input.EventSpecial, Special: input.KeyEnd},
		{Kind: input.EventKey, Rune: '!'},
		{Kind: input.EventSpecial, Special: input.KeyBackspace},
		{Kind: input.EventPaste, PasteText: "??"},
	}
	for _, evt := range steps {
		if !HandleKey(s, evt) {
			t.Fatalf("HandleKey did not handle %+v", evt)
		}
	}

	// Then
	if s.Value != "hei??" {
		t.Errorf("value = %q, want \"hei??\"", s.Value)
	}
}

func TestHandleKeyReadlineChords(t *testing.T) {
	// Given
	s := &State{Value: "abc def", Cursor: 3}

	// When / Then
	if !HandleKey(s, input.Event{Kind: input.EventKey, Rune: 'a', Ctrl: true}) || s.Cursor != 0 {
		t.Errorf("Ctrl+A: cursor %d, want 0", s.Cursor)
	}
	if !HandleKey(s, input.Event{Kind: input.EventKey, Rune: 'e', Ctrl: true}) || s.Cursor != 7 {
		t.Errorf("Ctrl+E: cursor %d, want 7", s.Cursor)
	}
	if !HandleKey(s, input.Event{Kind: input.EventKey, Rune: 'w', Ctrl: true}) || s.Value != "abc " {
		t.Errorf("Ctrl+W: value %q, want \"abc \"", s.Value)
	}
	if !HandleKey(s, input.Event{Kind: input.EventKey, Rune: 'u', Ctrl: true}) || s.Value != "" {
		t.Errorf("Ctrl+U: value %q, want empty", s.Value)
	}
	if !HandleKey(s, input.Event{Kind: input.EventKey, Rune: 'y', Ctrl: true}) || s.Value != "abc " {
		t.Errorf("Ctrl+Y yank: value %q, want \"abc \"", s.Value)
	}
}

func TestHandleKeyLeavesUnrelatedEventsAlone(t *testing.T) {
	// Given
	s := &State{Value: "x", Cursor: 1}

	// When / Then — events it doesn't own are not claimed
	for _, evt := range []input.Event{
		{Kind: input.EventKey, Rune: 'c', Ctrl: true},
		{Kind: input.EventSpecial, Special: input.KeyTab},
		{Kind: input.EventSpecial, Special: input.KeyEnter},
		{Kind: input.EventSpecial, Special: input.KeyEscape},
		{Kind: input.EventFrame},
	} {
		if HandleKey(s, evt) {
			t.Errorf("HandleKey claimed %+v", evt)
		}
	}
	if s.Value != "x" {
		t.Errorf("value changed: %q", s.Value)
	}
}
