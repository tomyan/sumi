package tui_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

// F3b: inline run mode — no alternate screen, no absolute row moves,
// the final frame stays in the output with the cursor parked below.

var absCUP = regexp.MustCompile(`\x1b\[\d+;\d+H`)

func TestInlineRunAvoidsAltScreenAndCUP(t *testing.T) {
	// Given
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindText, Content: "inline hello", CursorCol: -1, CursorRow: -1},
			{Kind: layout.KindText, Content: "second row", CursorCol: -1, CursorRow: -1},
		},
	}}
	var out bytes.Buffer

	// When
	tui.RunWithOptions(comp, tui.RunOptions{
		Inline: true, In: strings.NewReader("q"), Out: &out, ExitOn: []string{"q"},
	})
	s := out.String()

	// Then
	if strings.Contains(s, "\x1b[?1049h") {
		t.Error("inline mode must not enter the alternate screen")
	}
	if absCUP.MatchString(s) {
		t.Errorf("inline output contains absolute CUP: %q", s)
	}
	for _, want := range []string{"inline hello", "second row"} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q", want)
		}
	}
	// Cursor shown at the end (parked below the final frame).
	if !strings.HasSuffix(strings.TrimRight(s, "\r\n"), "\x1b[?25h") &&
		!strings.Contains(s[strings.LastIndex(s, "\x1b[?25h"):], "\x1b[?25h") {
		t.Errorf("output should end by showing the cursor: %q", s)
	}
	if strings.Contains(s, "\x1b[?1049l") {
		t.Error("inline exit must not pop the alternate screen")
	}
}

func TestInlineRunUpdatesDiffInPlace(t *testing.T) {
	// Given: a key mutates row content between frames.
	text := &layout.Input{Kind: layout.KindText, Content: "n=0", CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{text},
	}}
	comp.OnEvent = func(evt input.Event) {
		if evt.Rune == 'x' {
			text.Content = "n=1"
		}
	}
	var out bytes.Buffer

	// When
	tui.RunWithOptions(comp, tui.RunOptions{
		Inline: true, In: strings.NewReader("xq"), Out: &out, ExitOn: []string{"q"},
	})
	s := out.String()

	// Then: both states appear, still no CUP anywhere.
	if !strings.Contains(s, "n=0") || !strings.Contains(s, "1") {
		t.Errorf("output missing frames: %q", s)
	}
	if absCUP.MatchString(s) {
		t.Errorf("inline diff used absolute CUP: %q", s)
	}
}
