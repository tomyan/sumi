package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestTextInheritsColorFromParentBox(t *testing.T) {
	// Given a box with green FG containing a text node with no style
	input := &Input{
		Kind:  KindBox,
		Style: render.Style{FG: render.Color{Name: "green"}},
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then the text should render with green FG
	cell := buf.Cell(0, 0)
	if cell.Style.FG.Name != "green" {
		t.Errorf("expected green FG, got %q", cell.Style.FG.Name)
	}
}

func TestTextInheritsBoldFromParentBox(t *testing.T) {
	// Given a bold box containing unstyled text
	input := &Input{
		Kind:  KindBox,
		Style: render.Style{Bold: true},
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should be bold
	if !buf.Cell(0, 0).Style.Bold {
		t.Error("expected bold text")
	}
}

func TestTextOverridesParentColor(t *testing.T) {
	// Given a green box containing red text
	input := &Input{
		Kind:  KindBox,
		Style: render.Style{FG: render.Color{Name: "green"}},
		Children: []*Input{
			{Kind: KindText, Content: "hello", Style: render.Style{FG: render.Color{Name: "red"}}},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should be red (child overrides parent)
	cell := buf.Cell(0, 0)
	if cell.Style.FG.Name != "red" {
		t.Errorf("expected red FG, got %q", cell.Style.FG.Name)
	}
}

func TestInheritanceCascadesThroughNestedBoxes(t *testing.T) {
	// Given: outer (green) > middle (bold) > inner (text)
	input := &Input{
		Kind:  KindBox,
		Style: render.Style{FG: render.Color{Name: "green"}},
		Children: []*Input{
			{
				Kind:  KindBox,
				Style: render.Style{Bold: true},
				Children: []*Input{
					{Kind: KindText, Content: "hello"},
				},
			},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should be both green and bold
	cell := buf.Cell(0, 0)
	if cell.Style.FG.Name != "green" {
		t.Errorf("expected green FG, got %q", cell.Style.FG.Name)
	}
	if !cell.Style.Bold {
		t.Error("expected bold")
	}
}

func TestDimInherits(t *testing.T) {
	// Given a dim box containing unstyled text
	input := &Input{
		Kind:  KindBox,
		Style: render.Style{Dim: true},
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should be dim
	if !buf.Cell(0, 0).Style.Dim {
		t.Error("expected dim text")
	}
}

func TestHoverStyleApplied(t *testing.T) {
	// Given a box with dim style and a hover style that undims with white colour
	input := &Input{
		Kind:       KindBox,
		Style:      render.Style{Dim: true},
		HoverStyle: render.Style{FG: render.Color{Name: "white"}},
		Hovered:    true,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should have hover style (white FG, not dim)
	cell := buf.Cell(0, 0)
	if cell.Style.FG.Name != "white" {
		t.Errorf("expected white FG on hover, got %q", cell.Style.FG.Name)
	}
	if cell.Style.Dim {
		t.Error("expected not dim on hover")
	}
}

func TestHoverStyleNotAppliedWhenNotHovered(t *testing.T) {
	// Given same setup but Hovered=false
	input := &Input{
		Kind:       KindBox,
		Style:      render.Style{Dim: true},
		HoverStyle: render.Style{FG: render.Color{Name: "white"}},
		Hovered:    false,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should have base style (dim, no white)
	cell := buf.Cell(0, 0)
	if cell.Style.FG.Name == "white" {
		t.Error("expected no white FG when not hovered")
	}
	if !cell.Style.Dim {
		t.Error("expected dim when not hovered")
	}
}

func TestBGDoesNotInherit(t *testing.T) {
	// Given a box with red BG containing unstyled text
	// BG is NOT an inheritable property in CSS
	input := &Input{
		Kind:  KindBox,
		Style: render.Style{BG: render.Color{Name: "red"}},
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)
	buf := render.NewBuffer(40, 10)
	RenderTree(buf, box, nil)

	// Then text should NOT have red BG
	cell := buf.Cell(0, 0)
	if cell.Style.BG.Name == "red" {
		t.Error("BG should not inherit")
	}
}
