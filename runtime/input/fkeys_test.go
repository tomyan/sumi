package input

import (
	"strings"
	"testing"
)

// D2: F1–F12 function keys.

func TestFunctionKeysSS3(t *testing.T) {
	cases := map[string]SpecialKey{
		"\x1bOP": KeyF1, "\x1bOQ": KeyF2, "\x1bOR": KeyF3, "\x1bOS": KeyF4,
	}
	for seq, want := range cases {
		evt, err := ReadEvent(strings.NewReader(seq))
		if err != nil || evt.Special != want {
			t.Errorf("%q → %+v (err %v), want %s", seq, evt, err, want)
		}
	}
}

func TestFunctionKeysCSI(t *testing.T) {
	cases := map[string]SpecialKey{
		"\x1b[11~": KeyF1, "\x1b[15~": KeyF5, "\x1b[17~": KeyF6,
		"\x1b[21~": KeyF10, "\x1b[23~": KeyF11, "\x1b[24~": KeyF12,
	}
	for seq, want := range cases {
		evt, err := ReadEvent(strings.NewReader(seq))
		if err != nil || evt.Special != want {
			t.Errorf("%q → %+v (err %v), want %s", seq, evt, err, want)
		}
	}
}

func TestFunctionKeyWithModifier(t *testing.T) {
	// Shift+F5: ESC[15;2~
	evt, err := ReadEvent(strings.NewReader("\x1b[15;2~"))
	if err != nil || evt.Special != KeyF5 || !evt.Shift {
		t.Errorf("shift+f5 → %+v (err %v)", evt, err)
	}
}
