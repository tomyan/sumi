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
