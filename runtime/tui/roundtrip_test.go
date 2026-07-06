package tui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
	"github.com/tomyan/sumi/runtime/vt100"
)

// D7: ANSI round-trip — a real app run emits escape sequences through
// the injected writer; replaying them through sumi's vt100 terminal
// model must reproduce the frame, including diffed updates. Injected
// non-file output pins the viewport at 80x24.

func screenRow(s *vt100.Screen, row int, width int) string {
	var b strings.Builder
	for col := 0; col < width; col++ {
		ch := s.Cell(row, col).Ch
		if ch == 0 {
			ch = ' '
		}
		b.WriteRune(ch)
	}
	return strings.TrimRight(b.String(), " ")
}

func TestANSIRoundTripStaticFrame(t *testing.T) {
	// Given: a styled tree; input immediately quits.
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindText, Content: "round trip", CursorCol: -1, CursorRow: -1,
				Style: render.Style{FG: render.Color{Name: "red"}, Bold: true}},
		},
	}}
	var out bytes.Buffer

	// When
	tui.RunWithOptions(comp, tui.RunOptions{
		In: strings.NewReader("q"), Out: &out, ExitOn: []string{"q"},
	})
	screen := vt100.NewScreen(80, 24)
	if _, err := screen.Write(out.Bytes()); err != nil {
		t.Fatalf("vt100 replay: %v", err)
	}

	// Then
	if got := screenRow(screen, 0, 80); got != "round trip" {
		t.Errorf("row 0 = %q, want %q", got, "round trip")
	}
	cell := screen.Cell(0, 0)
	if cell.Style.FG.Name != "red" && !cell.Style.FG.IsRGB {
		t.Errorf("cell style = %+v, want red FG", cell.Style)
	}
	if !cell.Style.Bold {
		t.Errorf("cell should be bold: %+v", cell.Style)
	}
}

func TestANSIRoundTripDiffedUpdates(t *testing.T) {
	// Given: each keypress mutates the text; diffs must replay to the
	// final state on the model.
	text := &layout.Input{Kind: layout.KindText, Content: "count 0", CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{text},
	}}
	count := 0
	comp.OnEvent = func(evt input.Event) {
		if evt.Kind == input.EventKey && evt.Rune == 'x' {
			count++
			text.Content = "count " + string(rune('0'+count))
		}
	}
	var out bytes.Buffer

	// When: two increments, then quit.
	tui.RunWithOptions(comp, tui.RunOptions{
		In: strings.NewReader("xxq"), Out: &out, ExitOn: []string{"q"},
	})
	screen := vt100.NewScreen(80, 24)
	if _, err := screen.Write(out.Bytes()); err != nil {
		t.Fatalf("vt100 replay: %v", err)
	}

	// Then
	if got := screenRow(screen, 0, 80); got != "count 2" {
		t.Errorf("row 0 = %q, want %q", got, "count 2")
	}
}
