package layout

// layoutTable lays out a display:table container: children with
// display:table-row hold cells; column widths derive from the widest cell
// content per column (shrunk proportionally on overflow), rows size to
// their tallest cell, and cells stretch to fill their grid slot.
// B7a scope: no colspan/rowspan, caption, or collapsed cell borders yet.
func layoutTable(input *Input, rows []*Input, offsetX, offsetY, availW, availH int) []*Box {
	cells := tableCells(rows)
	colWidths := tableColumnWidths(cells, availW, availH)

	rowBoxes := make([]*Box, len(rows))
	cursorY := 0
	for r, row := range rows {
		cellBoxes, rowHeight := layoutTableRow(cells[r], colWidths, offsetX, offsetY+cursorY, availH)
		rowBoxes[r] = &Box{
			Kind:     KindBox,
			X:        offsetX,
			Y:        offsetY + cursorY,
			Width:    sum(colWidths),
			Height:   rowHeight,
			Style:    row.Style,
			Children: cellBoxes,
		}
		cursorY += rowHeight
	}
	return rowBoxes
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

// tableColumnWidths sizes each column to its widest cell's natural width,
// then shrinks proportionally if the total exceeds the available width.
func tableColumnWidths(cells [][]*Input, availW, availH int) []int {
	colCount := 0
	for _, row := range cells {
		if len(row) > colCount {
			colCount = len(row)
		}
	}
	widths := make([]int, colCount)
	for _, row := range cells {
		for c, cell := range row {
			w := layoutNode(cell, availW, availH).Width
			if w > widths[c] {
				widths[c] = w
			}
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

// layoutTableRow lays out one row's cells at the given column widths.
func layoutTableRow(cells []*Input, colWidths []int, offsetX, offsetY, availH int) ([]*Box, int) {
	cellBoxes := make([]*Box, len(cells))
	rowHeight := 1
	cursorX := 0
	for c, cell := range cells {
		w := colWidths[c]
		box := layoutNode(cell, w, availH)
		box.X = offsetX + cursorX
		box.Y = offsetY
		box.Width = w
		if box.Height > rowHeight {
			rowHeight = box.Height
		}
		cellBoxes[c] = box
		cursorX += w
	}
	for _, box := range cellBoxes {
		box.Height = rowHeight // cells stretch to the row
	}
	return cellBoxes, rowHeight
}

func sum(vals []int) int {
	total := 0
	for _, v := range vals {
		total += v
	}
	return total
}
