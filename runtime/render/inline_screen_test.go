package render

import (
	"regexp"
	"strings"
	"testing"
)

// F3a: InlineScreen — the inline-mode screen driver. The invariant:
// never emit an absolute row coordinate (no CUP); rows move relatively,
// columns via CHA, growth via LF, shrink via ED 0J, archive via
// ReleaseTop with zero output.

var cupRe = regexp.MustCompile(`\x1b\[\d+;\d+H`)

func inlineBuf(lines ...string) *Buffer {
	w := 0
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	buf := NewBuffer(w, len(lines))
	for row, l := range lines {
		for col, r := range l {
			buf.SetStyledCell(row, col, r, Style{})
		}
	}
	return buf
}

func TestInlineFirstRenderUsesNewlinesNotCUP(t *testing.T) {
	// Given
	s := NewInlineScreen()

	// When
	out := string(s.Render(inlineBuf("ab", "cd")))

	// Then
	if cupRe.MatchString(out) {
		t.Errorf("output %q contains absolute CUP", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("output %q: want exactly 1 LF to realise row 2", out)
	}
	for _, want := range []string{"ab", "cd"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestInlineUpdateRewritesOnlyChangedCells(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("ab", "cd"))

	// When: change one cell on row 0.
	out := string(s.Render(inlineBuf("aX", "cd")))

	// Then: relative up-move + column move + the glyph; no repaint of cd.
	if cupRe.MatchString(out) {
		t.Errorf("output %q contains absolute CUP", out)
	}
	if !strings.Contains(out, "\x1b[1A") {
		t.Errorf("output %q: want CUU to reach row 0", out)
	}
	if !strings.Contains(out, "X") || strings.Contains(out, "cd") {
		t.Errorf("output %q: want only the changed cell", out)
	}
}

func TestInlineGrowEmitsLF(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("ab"))

	// When
	out := string(s.Render(inlineBuf("ab", "cd", "ef")))

	// Then: two new physical lines via LF; no CUP.
	if strings.Count(out, "\n") != 2 {
		t.Errorf("output %q: want 2 LFs", out)
	}
	if cupRe.MatchString(out) {
		t.Errorf("output %q contains absolute CUP", out)
	}
}

func TestInlineShrinkErasesBelow(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("ab", "cd", "ef"))

	// When
	out := string(s.Render(inlineBuf("ab")))

	// Then
	if !strings.Contains(out, "\x1b[0J") {
		t.Errorf("output %q: want ED 0J", out)
	}
}

func TestInlineRegrowAfterShrinkReusesLines(t *testing.T) {
	// Given: shrink keeps physical lines realised as blanks.
	s := NewInlineScreen()
	s.Render(inlineBuf("ab", "cd", "ef"))
	s.Render(inlineBuf("ab"))

	// When: grow back within the realised zone.
	out := string(s.Render(inlineBuf("ab", "zz")))

	// Then: no LF needed — the physical line exists.
	if strings.Contains(out, "\n") {
		t.Errorf("output %q: regrowth within realised rows must not LF", out)
	}
	if !strings.Contains(out, "zz") {
		t.Errorf("output %q missing regrown content", out)
	}
}

func TestInlineReleaseTopEmitsNothingAndShiftsDiff(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("old", "abc", "cde"))

	// When: archive the top row.
	s.ReleaseTop(1)

	// Then: zero output; the next render of identical remaining content
	// emits nothing (diff aligns against shifted rows).
	out := string(s.Render(inlineBuf("abc", "cde")))
	if strings.Contains(out, "abc") || strings.Contains(out, "cde") {
		t.Errorf("output %q: shifted content should not repaint", out)
	}
}

func TestInlineWidthChangeRepaintsInPlace(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("abcd"))

	// When: terminal narrowed — same content, new width.
	next := NewBuffer(3, 1)
	for col, r := range "abc" {
		next.SetStyledCell(0, col, r, Style{})
	}
	out := string(s.Render(next))

	// Then: erase live zone + repaint, still no CUP.
	if !strings.Contains(out, "\x1b[0J") {
		t.Errorf("output %q: want in-place erase on width change", out)
	}
	if !strings.Contains(out, "abc") {
		t.Errorf("output %q: want repaint", out)
	}
	if cupRe.MatchString(out) {
		t.Errorf("output %q contains absolute CUP", out)
	}
}

func TestInlineFinishParksCursorBelowContent(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("ab", "cd"))

	// When
	out := string(s.Finish())

	// Then: fresh line + cursor shown.
	if !strings.Contains(out, "\x1b[?25h") {
		t.Errorf("output %q: want show-cursor", out)
	}
	if !strings.Contains(out, "\r\n") {
		t.Errorf("output %q: want cursor parked on a fresh line", out)
	}
}

// F3d: CPR origin bookkeeping + screen→zone mapping.

func TestInlineOriginMapping(t *testing.T) {
	// Given: a 3-row zone whose row 0 sits at screen row 10 (1-based).
	s := NewInlineScreen()
	s.Render(inlineBuf("aa", "bb", "cc"))
	s.SetOriginRow(10)

	// When / Then: screen rows map into zone rows (0-based screen in).
	if got, ok := s.ScreenRowToZone(9, 24); !ok || got != 0 {
		t.Errorf("screen 9 → %d,%v, want 0,true", got, ok)
	}
	if got, ok := s.ScreenRowToZone(11, 24); !ok || got != 2 {
		t.Errorf("screen 11 → %d,%v, want 2,true", got, ok)
	}
	if _, ok := s.ScreenRowToZone(5, 24); ok {
		t.Error("row above the zone must not map")
	}
	if _, ok := s.ScreenRowToZone(12, 24); ok {
		t.Error("row below the zone must not map")
	}
}

func TestInlineOriginUnknownMapsNothing(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("aa"))

	// Then
	if _, ok := s.ScreenRowToZone(0, 24); ok {
		t.Error("unknown origin must not map")
	}
}

func TestInlineOriginBottomClamp(t *testing.T) {
	// Given: zone grew past the bottom — physical rows exceed what the
	// stale origin allows; the zone bottom is pinned to the screen bottom.
	s := NewInlineScreen()
	s.Render(inlineBuf("a", "b", "c", "d"))
	s.SetOriginRow(23) // stale: 4 rows can't start at 23 in a 24-row term

	// When: effective origin clamps to 24-4+1 = 21.
	got, ok := s.ScreenRowToZone(23, 24) // screen row 24 (1-based)

	// Then: bottom row of the zone.
	if !ok || got != 3 {
		t.Errorf("screen 23 → %d,%v, want 3,true", got, ok)
	}
}

func TestInlineReleaseTopShiftsOrigin(t *testing.T) {
	// Given
	s := NewInlineScreen()
	s.Render(inlineBuf("aa", "bb", "cc"))
	s.SetOriginRow(10)

	// When: top row archived — zone row 0 moves down one screen row.
	s.ReleaseTop(1)

	// Then
	if got, ok := s.ScreenRowToZone(10, 24); !ok || got != 0 {
		t.Errorf("screen 10 → %d,%v, want 0,true after release", got, ok)
	}
}
