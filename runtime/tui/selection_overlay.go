package tui

import "github.com/tomyan/sumi/runtime/render"

// ApplySelectionOverlay paints a selection onto a rendered buffer by
// toggling inverse video over the ribbon: first row from Start.Col,
// middle rows whole, last row through End.Col. Toggling (not forcing)
// keeps already-inverse content readable inside the selection, and
// leaves characters untouched so text extraction stays exact.
func ApplySelectionOverlay(buf *render.Buffer, r *SelectionRange) {
	if r == nil {
		return
	}
	for row := r.Start.Row; row <= r.End.Row && row < buf.Height(); row++ {
		from, to := 0, buf.Width()-1
		if row == r.Start.Row {
			from = r.Start.Col
		}
		if row == r.End.Row && r.End.Col < to {
			to = r.End.Col
		}
		for col := from; col <= to; col++ {
			cell := buf.Cell(row, col)
			st := cell.Style
			st.Inverse = !st.Inverse
			ch := cell.Ch
			if ch == 0 {
				ch = ' '
			}
			buf.SetStyledCell(row, col, ch, st)
		}
	}
}
