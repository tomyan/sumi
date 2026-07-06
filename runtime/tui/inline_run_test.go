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

// F3d: CPR-derived origin maps mouse clicks into the zone; events
// outside it are dropped.
func TestInlineMouseMapsThroughOrigin(t *testing.T) {
	// Given: an inline app with a click handler on its only row, run
	// with a CPR reply placing the zone at screen row 10 (1-based),
	// then a click on that screen row, then quit.
	var clicks []int
	target := &layout.Input{Kind: layout.KindText, Content: "click me",
		CursorCol: -1, CursorRow: -1,
		On: map[string]func(*layout.DOMEvent){
			"click": func(evt *layout.DOMEvent) { clicks = append(clicks, evt.Key.Mouse.Y) },
		}}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{target},
	}}
	var out bytes.Buffer
	// CPR reply row 10; click at screen row 9 (0-based) col 2 → zone row 0;
	// a second click at screen row 3 is outside the zone and dropped.
	in := "\x1b[10;1R" + "\x1b[<0;3;10M" + "\x1b[<0;3;4M" + "q"

	// When
	tui.RunWithOptions(comp, tui.RunOptions{
		Inline: true, In: strings.NewReader(in), Out: &out, ExitOn: []string{"q"},
	})

	// Then: exactly one click, delivered at zone row 0.
	if len(clicks) != 1 || clicks[0] != 0 {
		t.Errorf("clicks = %v, want one click at zone row 0", clicks)
	}
}
