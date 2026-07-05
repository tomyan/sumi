package layout

// layoutTable lays out a display:table container: children with
// display:table-row hold cells (rows inside thead/tbody/tfoot groups are
// flattened); column widths derive from the widest cell content per column
// (shrunk proportionally on overflow), rows size to their tallest cell,
// and cells stretch to their slot. colspan/rowspan place like a grid; a
// caption child renders above the table at full width.
// Collapsed cell borders are still to come (B7c).
func layoutTable(input *Input, children []*Input, offsetX, offsetY, availW, availH int) []*Box {
	caption, rows := splitCaption(children)
	cells := tableCells(rows)
	colCount := tableColumnCount(cells)
	spacingH, spacingV := input.BorderSpacingH, input.BorderSpacingV
	trackW := availW - spacingH*max(colCount-1, 0)
	var colWidths []int
	if input.TableLayout == "fixed" {
		colWidths = fixedColumnWidths(cells, colCount, trackW)
	} else {
		colWidths = tableColumnWidths(cells, colCount, trackW, availH)
	}
	tableW := sum(colWidths) + spacingH*max(colCount-1, 0)

	var out []*Box
	cursorY := 0
	if caption != nil {
		capBox := layoutNode(caption, availW, availH)
		capBox.X = offsetX
		capBox.Y = offsetY
		capBox.Width = tableW
		out = append(out, capBox)
		cursorY += capBox.Height
	}

	rowBoxes, spans := placeTableRows(rows, cells, colWidths, spacingH, spacingV, offsetX, offsetY, &cursorY, availH)
	extendRowSpans(spans, rowBoxes)
	return append(out, rowBoxes...)
}

// fixedColumnWidths sizes columns from the first row only (table-layout:
// fixed): explicit cell widths are kept, and the remaining space splits
// evenly across the unsized columns. Later rows never widen a column.
func fixedColumnWidths(cells [][]*Input, colCount, availW int) []int {
	widths := make([]int, colCount)
	flexible := colCount
	used := 0
	if len(cells) > 0 {
		col := 0
		for _, cell := range cells[0] {
			if col >= colCount {
				break
			}
			if cell.FixedWidth > 0 {
				widths[col] = cell.FixedWidth
				used += cell.FixedWidth
				flexible--
			}
			col += spanOf(cell.ColSpan)
		}
	}
	if flexible > 0 {
		share := (availW - used) / flexible
		if share < 1 {
			share = 1
		}
		for i := range widths {
			if widths[i] == 0 {
				widths[i] = share
			}
		}
	}
	return widths
}

// spanningCell records a rowspanning cell so its height can be extended
// once the covered rows are sized.
type spanningCell struct {
	box      *Box
	rowStart int
	rowEnd   int
}

// placeTableRows lays out each row at the shared column widths, with
// border-spacing gaps between columns and rows.
func placeTableRows(rows []*Input, cells [][]*Input, colWidths []int, spacingH, spacingV, offsetX, offsetY int, cursorY *int, availH int) ([]*Box, []spanningCell) {
	occupied := map[[2]int]bool{} // [row, col] taken by an earlier rowspan
	var spans []spanningCell
	rowBoxes := make([]*Box, len(rows))
	rowWidth := sum(colWidths) + spacingH*max(len(colWidths)-1, 0)
	for r, row := range rows {
		cellBoxes, rowHeight := layoutSpannedRow(cells[r], colWidths, spacingH, r, occupied, &spans, offsetX, offsetY+*cursorY, availH)
		rowBoxes[r] = &Box{
			Kind:     KindBox,
			X:        offsetX,
			Y:        offsetY + *cursorY,
			Width:    rowWidth,
			Height:   rowHeight,
			Style:    row.Style,
			Children: cellBoxes,
		}
		*cursorY += rowHeight
		if r < len(rows)-1 {
			*cursorY += spacingV
		}
	}
	return rowBoxes, spans
}

// layoutSpannedRow places one row's cells, skipping columns covered by
// earlier rowspans and widening colspan cells across their columns.
func layoutSpannedRow(cellInputs []*Input, colWidths []int, spacingH, rowIdx int, occupied map[[2]int]bool, spans *[]spanningCell, offsetX, offsetY, availH int) ([]*Box, int) {
	var cellBoxes []*Box
	rowHeight := 1
	col := 0
	for _, cell := range cellInputs {
		for col < len(colWidths) && occupied[[2]int{rowIdx, col}] {
			col++
		}
		if col >= len(colWidths) {
			break
		}
		cspan := spanOf(cell.ColSpan)
		rspan := spanOf(cell.RowSpan)
		w := 0
		covered := 0
		for i := col; i < col+cspan && i < len(colWidths); i++ {
			w += colWidths[i]
			covered++
		}
		w += spacingH * max(covered-1, 0)
		box := layoutNode(cell, w, availH)
		box.X = offsetX + columnOffset(colWidths, col, spacingH)
		box.Y = offsetY
		box.Width = w
		if rspan == 1 && box.Height > rowHeight {
			rowHeight = box.Height
		}
		for r := rowIdx; r < rowIdx+rspan; r++ {
			for c := col; c < col+cspan; c++ {
				occupied[[2]int{r, c}] = true
			}
		}
		if rspan > 1 {
			*spans = append(*spans, spanningCell{box, rowIdx, rowIdx + rspan})
		}
		cellBoxes = append(cellBoxes, box)
		col += cspan
	}
	for _, box := range cellBoxes {
		if box.Height < rowHeight {
			box.Height = rowHeight
		}
	}
	return cellBoxes, rowHeight
}

// extendRowSpans stretches rowspanning cells over their covered rows.
func extendRowSpans(spans []spanningCell, rowBoxes []*Box) {
	for _, s := range spans {
		total := 0
		for r := s.rowStart; r < s.rowEnd && r < len(rowBoxes); r++ {
			total += rowBoxes[r].Height
		}
		if total > s.box.Height {
			s.box.Height = total
		}
	}
}

func spanOf(v int) int {
	if v < 1 {
		return 1
	}
	return v
}

func columnOffset(colWidths []int, col, spacingH int) int {
	off := 0
	for i := 0; i < col && i < len(colWidths); i++ {
		off += colWidths[i] + spacingH
	}
	return off
}

// splitCaption separates a caption child from row content, flattening
// thead/tbody/tfoot row groups.
func splitCaption(children []*Input) (*Input, []*Input) {
	var caption *Input
	var rows []*Input
	for _, c := range children {
		switch {
		case c.Tag == "caption":
			caption = c
		case c.Tag == "thead" || c.Tag == "tbody" || c.Tag == "tfoot":
			for _, r := range c.Children {
				if r != nil {
					rows = append(rows, r)
				}
			}
		default:
			rows = append(rows, c)
		}
	}
	return caption, rows
}

func tableColumnCount(cells [][]*Input) int {
	count := 0
	for _, row := range cells {
		n := 0
		for _, c := range row {
			n += spanOf(c.ColSpan)
		}
		if n > count {
			count = n
		}
	}
	return count
}

// tableCells collects each row's cell inputs (all visible children).
func tableCells(rows []*Input) [][]*Input {
	cells := make([][]*Input, len(rows))
	for r, row := range rows {
		for _, c := range row.Children {
			if c == nil || c.Display == "none" {
				continue
			}
			cells[r] = append(cells[r], c)
		}
	}
	return cells
}

// tableColumnWidths sizes each column to its widest single-column cell's
// natural width, then shrinks proportionally if the total exceeds the
// available width. Spanning cells don't contribute to column sizing (v1).
func tableColumnWidths(cells [][]*Input, colCount, availW, availH int) []int {
	widths := make([]int, colCount)
	for _, row := range cells {
		col := 0
		for _, cell := range row {
			span := spanOf(cell.ColSpan)
			if col >= colCount {
				break
			}
			if span == 1 {
				w := layoutNode(cell, availW, availH).Width
				if w > widths[col] {
					widths[col] = w
				}
			}
			col += span
		}
	}
	if total := sum(widths); total > availW && total > 0 {
		used := 0
		scaled := 0
		for i := range widths {
			scaled += widths[i]
			target := availW * scaled / total
			widths[i] = target - used
			used = target
		}
	}
	return widths
}

func sum(vals []int) int {
	total := 0
	for _, v := range vals {
		total += v
	}
	return total
}
