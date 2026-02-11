package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestRenderTreeCollapsedColumnBorders(t *testing.T) {
	// Given — two stacked boxes with shared horizontal border
	input := &Input{
		Kind:           KindBox,
		BorderCollapse: true,
		FixedWidth:     10,
		FixedHeight:    8,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FlexGrow: 1},
			{Kind: KindBox, Border: "single", FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 10, 8)
	buf := render.NewBuffer(10, 8)
	RenderTree(buf, box, nil)

	// Then — shared edge should use ├ on the left and ┤ on the right
	// child[0] bottom = child[1] top at the shared row
	sharedRow := box.Children[1].Y
	if c := buf.Cell(sharedRow, 0); c.Ch != '├' {
		t.Errorf("shared left = %c, want ├", c.Ch)
	}
	if c := buf.Cell(sharedRow, 9); c.Ch != '┤' {
		t.Errorf("shared right = %c, want ┤", c.Ch)
	}
	// Top-left corner of first child should be normal ┌
	if c := buf.Cell(0, 0); c.Ch != '┌' {
		t.Errorf("top-left = %c, want ┌", c.Ch)
	}
	// Bottom-right corner of last child should be normal ┘
	lastChild := box.Children[1]
	bottomRow := lastChild.Y + lastChild.Height - 1
	rightCol := lastChild.X + lastChild.Width - 1
	if c := buf.Cell(bottomRow, rightCol); c.Ch != '┘' {
		t.Errorf("bottom-right at (%d,%d) = %c, want ┘", bottomRow, rightCol, c.Ch)
	}
}

func TestRenderTreeCollapsedRowBorders(t *testing.T) {
	// Given — two side-by-side boxes with shared vertical border
	input := &Input{
		Kind:           KindBox,
		Direction:      "row",
		BorderCollapse: true,
		FixedWidth:     20,
		FixedHeight:    6,
		Children: []*Input{
			{Kind: KindBox, Border: "single", FlexGrow: 1},
			{Kind: KindBox, Border: "single", FlexGrow: 1},
		},
	}

	// When
	box := Layout(input, 20, 6)
	buf := render.NewBuffer(20, 6)
	RenderTree(buf, box, nil)

	// Then — shared edge should use ┬ at top and ┴ at bottom
	sharedCol := box.Children[1].X
	if c := buf.Cell(0, sharedCol); c.Ch != '┬' {
		t.Errorf("shared top = %c, want ┬", c.Ch)
	}
	if c := buf.Cell(5, sharedCol); c.Ch != '┴' {
		t.Errorf("shared bottom = %c, want ┴", c.Ch)
	}
}
