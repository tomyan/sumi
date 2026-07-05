package tui_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func barApp(tag string, attrs map[string]string) (*tui.Component, *layout.Input) {
	bar := &layout.Input{Kind: layout.KindBox, Tag: tag, Attrs: attrs,
		CursorCol: -1, CursorRow: -1}
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{bar},
	}}
	return comp, bar
}

func barContent(t *testing.T, bar *layout.Input) string {
	t.Helper()
	for _, c := range bar.Children {
		if c != nil && c.Kind == layout.KindText && c.Tag == "" {
			return c.Content
		}
	}
	t.Fatal("bar has no projection child")
	return ""
}

func TestProgressRendersFillAndTrack(t *testing.T) {
	// Given / When — HTML default max is 1; UA width is 20
	comp, bar := barApp("progress", map[string]string{"value": "0.5"})
	tui.TestApp(comp, 30, 3)

	// Then — half of 20 cells filled
	got := barContent(t, bar)
	if utf8.RuneCountInString(got) != 20 {
		t.Fatalf("bar width = %d runes (%q), want 20", utf8.RuneCountInString(got), got)
	}
	if !strings.HasPrefix(got, strings.Repeat("█", 10)) || strings.Count(got, "█") != 10 {
		t.Errorf("bar = %q, want 10 full blocks then track", got)
	}
	if !strings.HasSuffix(got, strings.Repeat("░", 10)) {
		t.Errorf("bar = %q, want ░ track suffix", got)
	}
}

func TestProgressEighthPartials(t *testing.T) {
	// Given / When — 0.53125 of 20 cells = 10.625 cells = 10 full + 5/8
	comp, bar := barApp("progress", map[string]string{"value": "0.53125"})
	tui.TestApp(comp, 30, 3)

	// Then
	got := barContent(t, bar)
	want := strings.Repeat("█", 10) + "▋" + strings.Repeat("░", 9)
	if got != want {
		t.Errorf("bar = %q, want %q", got, want)
	}
}

func TestProgressWithoutValueIsIndeterminate(t *testing.T) {
	// Given / When
	comp, bar := barApp("progress", nil)
	tui.TestApp(comp, 30, 3)

	// Then — all track
	if got := barContent(t, bar); got != strings.Repeat("░", 20) {
		t.Errorf("bar = %q, want indeterminate track", got)
	}
}

func TestMeterUsesMinAndMax(t *testing.T) {
	// Given / When — 30 in [20, 60] = 25% of 20 cells = 5
	comp, bar := barApp("meter", map[string]string{"value": "30", "min": "20", "max": "60"})
	tui.TestApp(comp, 30, 3)

	// Then
	got := barContent(t, bar)
	want := strings.Repeat("█", 5) + strings.Repeat("░", 15)
	if got != want {
		t.Errorf("bar = %q, want %q", got, want)
	}
}

func TestBarIsNotFocusable(t *testing.T) {
	// Given
	comp, _ := barApp("progress", map[string]string{"value": "0.5"})

	// When / Then
	if got := layout.CollectFocusables(comp.Tree); len(got) != 0 {
		t.Errorf("progress should not be focusable, got %v", got)
	}
}
