package render

import "testing"

func TestDrawCollapsedBorderTopCollapsed(t *testing.T) {
	// Given — top edge is collapsed, so top-left should use ├ and top-right should use ┤
	b := NewBuffer(10, 5)
	collapsed := CollapsedEdges{Top: true}

	// When
	b.DrawCollapsedBorder(0, 0, 10, 5, "single", Style{}, collapsed)

	// Then
	if c := b.Cell(0, 0); c.Ch != '├' {
		t.Errorf("top-left = %c, want ├", c.Ch)
	}
	if c := b.Cell(0, 9); c.Ch != '┤' {
		t.Errorf("top-right = %c, want ┤", c.Ch)
	}
	// Bottom corners should be normal
	if c := b.Cell(4, 0); c.Ch != '└' {
		t.Errorf("bottom-left = %c, want └", c.Ch)
	}
	if c := b.Cell(4, 9); c.Ch != '┘' {
		t.Errorf("bottom-right = %c, want ┘", c.Ch)
	}
}

func TestDrawCollapsedBorderLeftCollapsed(t *testing.T) {
	// Given — left edge is collapsed, so top-left should use ┬ and bottom-left should use ┴
	b := NewBuffer(10, 5)
	collapsed := CollapsedEdges{Left: true}

	// When
	b.DrawCollapsedBorder(0, 0, 10, 5, "single", Style{}, collapsed)

	// Then
	if c := b.Cell(0, 0); c.Ch != '┬' {
		t.Errorf("top-left = %c, want ┬", c.Ch)
	}
	if c := b.Cell(4, 0); c.Ch != '┴' {
		t.Errorf("bottom-left = %c, want ┴", c.Ch)
	}
	// Right corners should be normal
	if c := b.Cell(0, 9); c.Ch != '┐' {
		t.Errorf("top-right = %c, want ┐", c.Ch)
	}
	if c := b.Cell(4, 9); c.Ch != '┘' {
		t.Errorf("bottom-right = %c, want ┘", c.Ch)
	}
}

func TestDrawCollapsedBorderBothCollapsed(t *testing.T) {
	// Given — top + left collapsed: top-left should use ┼
	b := NewBuffer(10, 5)
	collapsed := CollapsedEdges{Top: true, Left: true}

	// When
	b.DrawCollapsedBorder(0, 0, 10, 5, "single", Style{}, collapsed)

	// Then
	if c := b.Cell(0, 0); c.Ch != '┼' {
		t.Errorf("top-left = %c, want ┼", c.Ch)
	}
	// top-right: top collapsed → ┤
	if c := b.Cell(0, 9); c.Ch != '┤' {
		t.Errorf("top-right = %c, want ┤", c.Ch)
	}
	// bottom-left: left collapsed → ┴
	if c := b.Cell(4, 0); c.Ch != '┴' {
		t.Errorf("bottom-left = %c, want ┴", c.Ch)
	}
	// bottom-right: nothing collapsed → normal ┘
	if c := b.Cell(4, 9); c.Ch != '┘' {
		t.Errorf("bottom-right = %c, want ┘", c.Ch)
	}
}

func TestDrawCollapsedBorderNoneCollapsed(t *testing.T) {
	// Given — no edges collapsed: same as normal DrawStyledBorder
	b := NewBuffer(10, 5)
	collapsed := CollapsedEdges{}

	// When
	b.DrawCollapsedBorder(0, 0, 10, 5, "single", Style{}, collapsed)

	// Then — normal corners
	if c := b.Cell(0, 0); c.Ch != '┌' {
		t.Errorf("top-left = %c, want ┌", c.Ch)
	}
	if c := b.Cell(0, 9); c.Ch != '┐' {
		t.Errorf("top-right = %c, want ┐", c.Ch)
	}
	if c := b.Cell(4, 0); c.Ch != '└' {
		t.Errorf("bottom-left = %c, want └", c.Ch)
	}
	if c := b.Cell(4, 9); c.Ch != '┘' {
		t.Errorf("bottom-right = %c, want ┘", c.Ch)
	}
}
