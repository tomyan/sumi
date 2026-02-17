package input

import (
	"strings"
	"testing"
)

func TestEncodeEventPlainRune(t *testing.T) {
	// Given
	evt := Event{Kind: EventKey, Rune: 'a'}

	// When
	encoded := EncodeEvent(evt)
	decoded, err := ReadEvent(strings.NewReader(string(encoded)))

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded.Kind != EventKey {
		t.Errorf("Kind = %d, want EventKey", decoded.Kind)
	}
	if decoded.Rune != 'a' {
		t.Errorf("Rune = %c, want 'a'", decoded.Rune)
	}
}

func TestEncodeEventUTF8(t *testing.T) {
	// Given
	cases := []struct {
		name string
		r    rune
	}{
		{"two-byte", 'é'},
		{"three-byte", '日'},
		{"four-byte", '🎉'},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evt := Event{Kind: EventKey, Rune: tc.r}

			// When
			encoded := EncodeEvent(evt)
			decoded, err := ReadEvent(strings.NewReader(string(encoded)))

			// Then
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decoded.Rune != tc.r {
				t.Errorf("Rune = %U, want %U", decoded.Rune, tc.r)
			}
		})
	}
}

func TestEncodeEventCtrlKey(t *testing.T) {
	// Given
	cases := []struct {
		name string
		r    rune
	}{
		{"ctrl-a", 'a'},
		{"ctrl-c", 'c'},
		{"ctrl-z", 'z'},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evt := Event{Kind: EventKey, Rune: tc.r, Ctrl: true}

			// When
			encoded := EncodeEvent(evt)
			decoded, err := ReadEvent(strings.NewReader(string(encoded)))

			// Then
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !decoded.Ctrl {
				t.Error("expected Ctrl = true")
			}
			if decoded.Rune != tc.r {
				t.Errorf("Rune = %c, want %c", decoded.Rune, tc.r)
			}
		})
	}
}

func TestEncodeEventAltKey(t *testing.T) {
	// Given
	evt := Event{Kind: EventKey, Rune: 'x', Alt: true}

	// When
	encoded := EncodeEvent(evt)
	decoded, err := ReadEvent(strings.NewReader(string(encoded)))

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decoded.Alt {
		t.Error("expected Alt = true")
	}
	if decoded.Rune != 'x' {
		t.Errorf("Rune = %c, want 'x'", decoded.Rune)
	}
}

func TestEncodeEventSpecialKeys(t *testing.T) {
	// Given
	cases := []struct {
		name    string
		special SpecialKey
	}{
		{"up", KeyUp},
		{"down", KeyDown},
		{"left", KeyLeft},
		{"right", KeyRight},
		{"home", KeyHome},
		{"end", KeyEnd},
		{"pgup", KeyPgUp},
		{"pgdn", KeyPgDn},
		{"tab", KeyTab},
		{"shift-tab", KeyShiftTab},
		{"enter", KeyEnter},
		{"escape", KeyEscape},
		{"backspace", KeyBackspace},
		{"delete", KeyDelete},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evt := Event{Kind: EventSpecial, Special: tc.special}

			// When
			encoded := EncodeEvent(evt)
			decoded, err := ReadEvent(strings.NewReader(string(encoded)))

			// Then
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decoded.Kind != EventSpecial {
				t.Errorf("Kind = %d, want EventSpecial", decoded.Kind)
			}
			if decoded.Special != tc.special {
				t.Errorf("Special = %q, want %q", decoded.Special, tc.special)
			}
		})
	}
}

func TestEncodeEventPaste(t *testing.T) {
	// Given
	evt := Event{Kind: EventPaste, PasteText: "hello world"}

	// When
	encoded := EncodeEvent(evt)
	decoded, err := ReadEvent(strings.NewReader(string(encoded)))

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded.Kind != EventPaste {
		t.Errorf("Kind = %d, want EventPaste", decoded.Kind)
	}
	if decoded.PasteText != "hello world" {
		t.Errorf("PasteText = %q, want %q", decoded.PasteText, "hello world")
	}
}

func TestEncodeEventCtrlSlash(t *testing.T) {
	// Given — Ctrl+/ sends 0x1f
	evt := Event{Kind: EventKey, Rune: '/', Ctrl: true}

	// When
	encoded := EncodeEvent(evt)
	decoded, err := ReadEvent(strings.NewReader(string(encoded)))

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decoded.Ctrl {
		t.Error("expected Ctrl = true")
	}
	if decoded.Rune != '/' {
		t.Errorf("Rune = %c, want '/'", decoded.Rune)
	}
}

func TestEncodeEventCtrlBackslash(t *testing.T) {
	// Given — Ctrl+\ sends 0x1c. Note: ReadEvent doesn't decode this
	// back to Ctrl+'\', it returns rune 0x1c without Ctrl flag.
	// So we just verify the encoded byte is correct.
	evt := Event{Kind: EventKey, Rune: '\\', Ctrl: true}

	// When
	encoded := EncodeEvent(evt)

	// Then
	if len(encoded) != 1 || encoded[0] != 0x1c {
		t.Errorf("encoded = %v, want [0x1c]", encoded)
	}
}
