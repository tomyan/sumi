package tui

import (
	"strings"
	"time"

	"github.com/tomyan/sumi/runtime/render"
)

// Screen-space text selection over the painted cell grid (svelterm's
// selection model): coordinates are buffer cells, the column is the
// text index, and extraction reads whatever is painted when the mouse
// is released. Selection state survives re-renders at the same screen
// location; it is not anchored to logical content.

// CellPos is a 0-indexed buffer cell.
type CellPos struct {
	Col, Row int
}

// SelectionRange is a normalized (row-major start..end) selection.
type SelectionRange struct {
	Start, End CellPos
}

const multiClickMs = 400

// SelectionController tracks a mouse-driven selection: anchor is the
// fixed end (set on press), point the moving end (set while dragging).
// A plain click leaves point nil, so it clears any prior selection.
type SelectionController struct {
	getBuffer func() *render.Buffer
	now       func() int64 // milliseconds; injectable for tests

	anchor, point *CellPos
	dragging      bool
	pressed       bool
	lastClick     struct {
		pos   CellPos
		time  int64
		count int
	}
}

// NewSelectionController builds a controller reading text through
// getBuffer. now returns a monotonic timestamp in milliseconds.
func NewSelectionController(getBuffer func() *render.Buffer, now func() int64) *SelectionController {
	if now == nil {
		now = func() int64 { return time.Now().UnixMilli() }
	}
	return &SelectionController{getBuffer: getBuffer, now: now}
}

// Range returns the normalized selection, or nil when empty.
func (s *SelectionController) Range() *SelectionRange {
	if s.anchor == nil || s.point == nil {
		return nil
	}
	if before(*s.anchor, *s.point) {
		return &SelectionRange{Start: *s.anchor, End: *s.point}
	}
	return &SelectionRange{Start: *s.point, End: *s.anchor}
}

func before(a, b CellPos) bool {
	if a.Row != b.Row {
		return a.Row < b.Row
	}
	return a.Col <= b.Col
}

// OnPress handles a left-button press: multi-clicks on the same cell
// select the word (2) or line (3+); a single press arms a fresh drag
// and clears any prior selection.
func (s *SelectionController) OnPress(col, row int) {
	pos := CellPos{Col: col, Row: row}
	s.pressed = true
	t := s.now()
	if s.lastClick.count > 0 && s.lastClick.pos == pos && t-s.lastClick.time < multiClickMs {
		s.lastClick.count++
	} else {
		s.lastClick.count = 1
	}
	s.lastClick.pos = pos
	s.lastClick.time = t

	switch {
	case s.lastClick.count == 2:
		s.selectWord(pos)
	case s.lastClick.count >= 3:
		s.selectLine(pos)
	default:
		s.anchor = &pos
		s.point = nil
		s.dragging = false
	}
}

// OnMotion extends the selection while the button is held. Returns
// whether the visible selection changed (caller repaints).
func (s *SelectionController) OnMotion(col, row int) bool {
	if !s.pressed || s.anchor == nil {
		return false
	}
	pos := CellPos{Col: col, Row: row}
	if !s.dragging && pos == *s.anchor {
		return false
	}
	s.dragging = true
	moved := s.point == nil || *s.point != pos
	s.point = &pos
	return moved
}

// OnRelease ends the gesture and returns the selected text (empty when
// nothing is selected). The highlight persists after release.
func (s *SelectionController) OnRelease() string {
	s.pressed = false
	s.dragging = false
	r := s.Range()
	if r == nil {
		return ""
	}
	return s.extractText(*r)
}

// Clear drops the selection; reports whether anything was cleared.
func (s *SelectionController) Clear() bool {
	had := s.anchor != nil || s.point != nil
	s.anchor = nil
	s.point = nil
	s.dragging = false
	return had
}

// selectWord selects the whitespace-delimited word under pos.
// Clicking whitespace leaves the selection untouched.
func (s *SelectionController) selectWord(pos CellPos) {
	text := []rune(s.rowText(pos.Row))
	if pos.Col >= len(text) || text[pos.Col] == ' ' {
		return
	}
	start, end := pos.Col, pos.Col
	for start > 0 && text[start-1] != ' ' {
		start--
	}
	for end < len(text)-1 && text[end+1] != ' ' {
		end++
	}
	s.anchor = &CellPos{Col: start, Row: pos.Row}
	s.point = &CellPos{Col: end, Row: pos.Row}
}

// selectLine selects the full buffer row.
func (s *SelectionController) selectLine(pos CellPos) {
	buf := s.getBuffer()
	if buf == nil {
		return
	}
	s.anchor = &CellPos{Col: 0, Row: pos.Row}
	s.point = &CellPos{Col: buf.Width() - 1, Row: pos.Row}
}

// rowText reads the painted characters of one buffer row (empty cells
// read as spaces).
func (s *SelectionController) rowText(row int) string {
	buf := s.getBuffer()
	if buf == nil || row < 0 || row >= buf.Height() {
		return ""
	}
	var b strings.Builder
	for col := 0; col < buf.Width(); col++ {
		ch := buf.Cell(row, col).Ch
		if ch == 0 {
			ch = ' '
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// extractText converts the selection ribbon to a string: first row from
// Start.Col, last row through End.Col, middle rows whole; each line
// right-trimmed; lines joined with newlines.
func (s *SelectionController) extractText(r SelectionRange) string {
	var lines []string
	for row := r.Start.Row; row <= r.End.Row; row++ {
		text := []rune(s.rowText(row))
		from, to := 0, len(text)-1
		if row == r.Start.Row {
			from = r.Start.Col
		}
		if row == r.End.Row {
			to = r.End.Col
		}
		if from > len(text) {
			from = len(text)
		}
		if to >= len(text) {
			to = len(text) - 1
		}
		line := ""
		if to >= from {
			line = string(text[from : to+1])
		}
		lines = append(lines, strings.TrimRight(line, " \t"))
	}
	return strings.Join(lines, "\n")
}
