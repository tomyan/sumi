package layout

import "testing"

// B7a: core table layout.

func tableTree() *Input {
	row := func(cells ...string) *Input {
		r := &Input{Kind: KindBox, Display: "table-row"}
		for _, c := range cells {
			r.Children = append(r.Children, &Input{Kind: KindBox, Children: []*Input{
				{Kind: KindText, Content: c},
			}})
		}
		return r
	}
	return &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", Children: []*Input{
			row("name", "value"),
			row("a much longer name", "x"),
		},
	}}}
}

func TestTableColumnsSizeToWidestCell(t *testing.T) {
	// Given / When
	box := Layout(tableTree(), 60, 10)

	// Then: column 0 = 18 ("a much longer name"), column 1 = 5 ("value").
	table := box.Children[0]
	row0 := table.Children[0]
	if got := row0.Children[0].Width; got != 18 {
		t.Errorf("col 0 width = %d, want 18", got)
	}
	if got := row0.Children[1].X; got != 18 {
		t.Errorf("col 1 X = %d, want 18", got)
	}
	if got := row0.Children[1].Width; got != 5 {
		t.Errorf("col 1 width = %d, want 5", got)
	}
}

func TestTableRowsStack(t *testing.T) {
	box := Layout(tableTree(), 60, 10)
	table := box.Children[0]
	if got := table.Children[1].Y; got != 1 {
		t.Errorf("row 1 Y = %d, want 1", got)
	}
	if got := table.Height; got != 2 {
		t.Errorf("table height = %d, want 2", got)
	}
}

func TestTableShrinksToAvailableWidth(t *testing.T) {
	// Given: natural columns 18+5=23 in only 12 available.
	tree := tableTree()
	tree.Children[0].FixedWidth = 12

	// When
	box := Layout(tree, 60, 10)

	// Then: columns scale to fit 12.
	table := box.Children[0]
	row0 := table.Children[0]
	total := row0.Children[0].Width + row0.Children[1].Width
	if total != 12 {
		t.Errorf("total column width = %d, want 12", total)
	}
}

func TestTableCellsStretchToRowHeight(t *testing.T) {
	// Given: one tall cell makes the whole row tall.
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", Children: []*Input{
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, FixedHeight: 3},
				{Kind: KindBox},
			}},
		},
	}}}
	box := Layout(tree, 40, 10)
	row := box.Children[0].Children[0]
	if got := row.Children[1].Height; got != 3 {
		t.Errorf("short cell height = %d, want 3 (stretched)", got)
	}
}

// B7b: colspan/rowspan, caption, row groups.

func TestTableColspanWidensCell(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", Children: []*Input{
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, ColSpan: 2, Children: []*Input{{Kind: KindText, Content: "wide"}}},
			}},
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "aaaaa"}}},
				{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "bbb"}}},
			}},
		},
	}}}
	box := Layout(tree, 60, 10)
	table := box.Children[0]
	spanCell := table.Children[0].Children[0]
	if got := spanCell.Width; got != 8 {
		t.Errorf("colspan cell width = %d, want 8 (5+3)", got)
	}
}

func TestTableRowspanExtendsHeight(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", Children: []*Input{
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, RowSpan: 2, Children: []*Input{{Kind: KindText, Content: "tall"}}},
				{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "r1"}}},
			}},
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "r2"}}},
			}},
		},
	}}}
	box := Layout(tree, 60, 10)
	table := box.Children[0]
	tall := table.Children[0].Children[0]
	if got := tall.Height; got != 2 {
		t.Errorf("rowspan cell height = %d, want 2", got)
	}
	// Second row's cell lands in column 2 (column 1 occupied).
	r2cell := table.Children[1].Children[0]
	if got := r2cell.X; got != 4 {
		t.Errorf("second-row cell X = %d, want 4 (after 4-wide col)", got)
	}
}

func TestTableCaptionAboveRows(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", Children: []*Input{
			{Tag: "caption", Kind: KindText, Content: "My Table"},
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "cell"}}},
			}},
		},
	}}}
	box := Layout(tree, 60, 10)
	table := box.Children[0]
	if got := table.Children[0].Y; got != 0 {
		t.Errorf("caption Y = %d, want 0", got)
	}
	if got := table.Children[1].Y; got != 1 {
		t.Errorf("row Y = %d, want 1 (below caption)", got)
	}
}

func TestTableRowGroupsFlatten(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", Children: []*Input{
			{Tag: "thead", Kind: KindBox, Children: []*Input{
				{Kind: KindBox, Display: "table-row", Children: []*Input{
					{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "head"}}},
				}},
			}},
			{Tag: "tbody", Kind: KindBox, Children: []*Input{
				{Kind: KindBox, Display: "table-row", Children: []*Input{
					{Kind: KindBox, Children: []*Input{{Kind: KindText, Content: "body"}}},
				}},
			}},
		},
	}}}
	box := Layout(tree, 60, 10)
	table := box.Children[0]
	if len(table.Children) != 2 {
		t.Fatalf("rows = %d, want 2 (groups flattened)", len(table.Children))
	}
	if table.Children[1].Y != 1 {
		t.Errorf("tbody row Y = %d, want 1", table.Children[1].Y)
	}
}

// B7c: border-spacing gaps columns and rows.
func TestTableBorderSpacing(t *testing.T) {
	// Given
	tree := tableTree()
	tree.Children[0].BorderSpacingH = 2
	tree.Children[0].BorderSpacingV = 1

	// When
	box := Layout(tree, 60, 10)

	// Then — column 1 sits 2 past column 0; row 1 one below row 0
	table := box.Children[0]
	row0 := table.Children[0]
	if got := row0.Children[1].X; got != 20 {
		t.Errorf("col 1 X = %d, want 20 (18 + 2 spacing)", got)
	}
	if got := table.Children[1].Y; got != 2 {
		t.Errorf("row 1 Y = %d, want 2 (1 + 1 spacing)", got)
	}
	if got := row0.Width; got != 25 {
		t.Errorf("row width = %d, want 25 (18+2+5)", got)
	}
}

// B7c: table-layout fixed sizes columns from the first row only.
func TestTableLayoutFixed(t *testing.T) {
	// Given — first-row cells: explicit 10 and unsized; table 30 wide
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "table", FixedWidth: 30, TableLayout: "fixed",
		Children: []*Input{
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, Display: "table-cell", FixedWidth: 10,
					Children: []*Input{{Kind: KindText, Content: "a"}}},
				{Kind: KindBox, Display: "table-cell",
					Children: []*Input{{Kind: KindText, Content: "b"}}},
			}},
			{Kind: KindBox, Display: "table-row", Children: []*Input{
				{Kind: KindBox, Display: "table-cell",
					Children: []*Input{{Kind: KindText, Content: "a much longer content"}}},
				{Kind: KindBox, Display: "table-cell",
					Children: []*Input{{Kind: KindText, Content: "x"}}},
			}},
		},
	}}}

	// When
	box := Layout(tree, 60, 10)

	// Then — 10 explicit + remainder 20; long content did not widen col 0
	table := box.Children[0]
	row0 := table.Children[0]
	if got := row0.Children[0].Width; got != 10 {
		t.Errorf("col 0 width = %d, want fixed 10", got)
	}
	if got := row0.Children[1].Width; got != 20 {
		t.Errorf("col 1 width = %d, want remainder 20", got)
	}
}
