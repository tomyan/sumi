package render

import "fmt"

// InlineScreen drives a live render zone at the shell cursor without
// the alternate screen. The zone's absolute position is unknown by
// design: rows move relatively (CUU/CUD), columns absolutely (CHA),
// growth appends LF newlines (LF scrolls where cursor-down cannot),
// shrink erases with ED 0J, and archiving (ReleaseTop) emits nothing —
// released rows already sit on the terminal and enter scrollback
// naturally.
type InlineScreen struct {
	prev          *Buffer // last painted content, padded to physicalRows
	physicalRows  int     // lines realised on the terminal (shrink keeps blanks)
	contentHeight int     // rows the last frame actually used
	cursorRow     int     // relative to zone origin (may go negative after ReleaseTop)
	cursorCol     int     // -1 = wrap-pending (last write hit the final column)
	originRow     int     // 1-based screen row of zone row 0 (CPR); 0 = unknown
}

// NewInlineScreen returns a driver for a fresh zone at the cursor.
func NewInlineScreen() *InlineScreen {
	return &InlineScreen{}
}

// PhysicalRows reports the lines the zone has realised on the terminal.
func (s *InlineScreen) PhysicalRows() int { return s.physicalRows }

// Render diffs next against the previous frame and returns the ANSI to
// update the zone. Never emits an absolute row coordinate.
func (s *InlineScreen) Render(next *Buffer) []byte {
	var out []byte
	if s.prev != nil && s.prev.Width() != next.Width() {
		// The terminal may have rewrapped our rows: erase the live zone
		// in place and repaint fully.
		out = append(out, s.moveRow(0)...)
		out = append(out, '\r')
		out = append(out, "\x1b[0J"...)
		s.cursorCol = 0
		s.prev = nil
	}
	if next.Height() > s.physicalRows {
		out = append(out, s.grow(next.Height())...)
	}
	if next.Height() < s.contentHeight {
		out = append(out, s.eraseBelow(next.Height())...)
	}
	out = append(out, s.diff(next)...)
	s.prev = padToHeight(next, s.physicalRows)
	s.contentHeight = next.Height()
	return out
}

// diff walks cells, moving relatively and rewriting only changes.
func (s *InlineScreen) diff(next *Buffer) []byte {
	var out []byte
	var prevStyle Style
	styled := false
	for row := 0; row < next.Height(); row++ {
		for col := 0; col < next.Width(); col++ {
			desired := next.Cell(row, col)
			if s.cellUnchanged(desired, row, col) {
				continue
			}
			out = append(out, s.moveTo(row, col)...)
			if desired.Style != prevStyle || (styled != !desired.Style.IsZero()) {
				if desired.Style.IsZero() {
					out = append(out, "\x1b[0m"...)
					styled = false
				} else {
					out = appendSGR(out, desired.Style)
					styled = true
				}
				prevStyle = desired.Style
			}
			ch := desired.Ch
			if ch == 0 {
				ch = ' '
			}
			out = appendRune(out, ch)
			s.cursorRow = row
			if col+1 >= next.Width() {
				s.cursorCol = -1 // wrap-pending: force explicit moves next time
			} else {
				s.cursorCol = col + 1
			}
		}
	}
	if styled {
		out = append(out, "\x1b[0m"...)
	}
	return out
}

// cellUnchanged reports whether the terminal already shows the desired
// cell. With no comparison frame the zone is known blank (fresh lines
// or a just-erased zone), so zero cells are already correct.
func (s *InlineScreen) cellUnchanged(desired Cell, row, col int) bool {
	if s.prev == nil {
		return desired.Ch == 0 && desired.Style.IsZero()
	}
	if row >= s.prev.Height() || col >= s.prev.Width() {
		return desired.Ch == 0 && desired.Style.IsZero()
	}
	return desired == s.prev.Cell(row, col)
}

// moveTo positions the cursor with a relative row move and an absolute
// column (CHA) — the column is knowable, the row is not.
func (s *InlineScreen) moveTo(row, col int) []byte {
	var out []byte
	out = append(out, s.moveRow(row)...)
	if col != s.cursorCol {
		out = append(out, fmt.Sprintf("\x1b[%dG", col+1)...)
		s.cursorCol = col
	}
	return out
}

// moveRow emits a relative CUU/CUD to reach the zone row.
func (s *InlineScreen) moveRow(row int) []byte {
	delta := row - s.cursorRow
	s.cursorRow = row
	switch {
	case delta > 0:
		return []byte(fmt.Sprintf("\x1b[%dB", delta))
	case delta < 0:
		return []byte(fmt.Sprintf("\x1b[%dA", -delta))
	}
	return nil
}

// grow realises new physical lines with LF (cursor-down cannot scroll
// past the viewport bottom; LF can).
func (s *InlineScreen) grow(to int) []byte {
	var out []byte
	count := to - s.physicalRows
	if s.physicalRows == 0 {
		count = to - 1 // the cursor already sits on row 0
	} else {
		out = append(out, s.moveRow(s.physicalRows-1)...)
	}
	out = append(out, '\r')
	for i := 0; i < count; i++ {
		out = append(out, '\n')
	}
	s.cursorRow = to - 1
	s.cursorCol = 0
	s.physicalRows = to
	return out
}

// eraseBelow clears rows from height down; the physical lines stay
// realised as blanks so regrowth reuses them without scrolling.
func (s *InlineScreen) eraseBelow(height int) []byte {
	var out []byte
	out = append(out, s.moveRow(height)...)
	out = append(out, '\r')
	out = append(out, "\x1b[0J"...)
	s.cursorCol = 0
	if s.prev != nil {
		blankBelow(s.prev, height)
	}
	return out
}

// ReleaseTop archives the top n zone rows into scrollback: no output —
// the rows already look right — the driver just narrows its comparison
// window. The cursor's zone coordinate may go negative (it now sits in
// the archived area); relative moves remain correct.
func (s *InlineScreen) ReleaseTop(n int) {
	if n > s.contentHeight {
		n = s.contentHeight
	}
	if n <= 0 {
		return
	}
	s.prev = dropTopRows(s.prev, n)
	s.physicalRows -= n
	s.contentHeight -= n
	s.cursorRow -= n
	if s.originRow > 0 {
		s.originRow += n // zone row 0 moves down as top rows release
	}
}

// CursorRow reports the cursor's zone row (for CPR query snapshots).
func (s *InlineScreen) CursorRow() int { return s.cursorRow }

// SetOriginRow records the 1-based screen row of zone row 0, derived
// from a CPR reply minus the cursor's zone row at query time.
func (s *InlineScreen) SetOriginRow(row int) {
	if row < 1 {
		row = 1
	}
	s.originRow = row
}

// ScreenRowToZone maps a 0-based screen row to a zone row. The
// effective origin clamps so the zone bottom stays pinned to the
// screen bottom after LF growth scrolled the zone up. Reports false
// when the origin is unknown or the row lies outside the zone.
func (s *InlineScreen) ScreenRowToZone(screenRow, termH int) (int, bool) {
	if s.originRow == 0 || s.physicalRows == 0 {
		return 0, false
	}
	origin := s.originRow
	if maxOrigin := termH - s.physicalRows + 1; origin > maxOrigin {
		origin = maxOrigin
	}
	zone := screenRow - (origin - 1)
	if zone < 0 || zone >= s.physicalRows {
		return 0, false
	}
	return zone, true
}

// Finish parks the cursor on a fresh line after the content and shows
// it — the final frame stays in the scrollback.
func (s *InlineScreen) Finish() []byte {
	var out []byte
	if s.contentHeight < s.physicalRows {
		out = append(out, s.moveRow(s.contentHeight)...)
		out = append(out, '\r')
	} else {
		if s.physicalRows > 0 {
			out = append(out, s.moveRow(s.physicalRows-1)...)
		}
		out = append(out, "\r\n"...)
		s.cursorRow++
	}
	out = append(out, "\x1b[?25h"...)
	return out
}

// Reset forgets the zone entirely (suspend: the shell owned the screen
// while we were stopped; a fresh zone starts wherever the cursor is).
func (s *InlineScreen) Reset() {
	*s = InlineScreen{}
}

// padToHeight returns a copy of buf extended with blank rows to height.
func padToHeight(buf *Buffer, height int) *Buffer {
	if buf.Height() >= height {
		return buf
	}
	padded := NewBuffer(buf.Width(), height)
	copyCells(padded, buf)
	return padded
}

// dropTopRows returns buf without its first n rows.
func dropTopRows(buf *Buffer, n int) *Buffer {
	if buf == nil {
		return nil
	}
	if n >= buf.Height() {
		return NewBuffer(buf.Width(), 0)
	}
	out := NewBuffer(buf.Width(), buf.Height()-n)
	for row := n; row < buf.Height(); row++ {
		for col := 0; col < buf.Width(); col++ {
			c := buf.Cell(row, col)
			out.SetStyledCell(row-n, col, c.Ch, c.Style)
		}
	}
	return out
}

// blankBelow zeroes rows at and below height.
func blankBelow(buf *Buffer, height int) {
	for row := height; row < buf.Height(); row++ {
		for col := 0; col < buf.Width(); col++ {
			buf.SetStyledCell(row, col, 0, Style{})
		}
	}
}

// copyCells copies overlapping cells from src to dst.
func copyCells(dst, src *Buffer) {
	for row := 0; row < src.Height() && row < dst.Height(); row++ {
		for col := 0; col < src.Width() && col < dst.Width(); col++ {
			c := src.Cell(row, col)
			dst.SetStyledCell(row, col, c.Ch, c.Style)
		}
	}
}
