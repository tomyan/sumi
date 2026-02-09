package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestLayoutBoxWithPadding(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Padding:   Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if len(box.Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(box.Children))
	}

	child := box.Children[0]
	// Child should be offset by padding
	if child.X != 2 {
		t.Errorf("child X = %d, want 2 (left padding)", child.X)
	}
	if child.Y != 1 {
		t.Errorf("child Y = %d, want 1 (top padding)", child.Y)
	}

	// Parent auto-size: child width + left padding + right padding
	if box.Width != 2+2+2 {
		t.Errorf("Width = %d, want %d (content + padding)", box.Width, 2+2+2)
	}
	// Parent auto-size: child height + top padding + bottom padding
	if box.Height != 1+1+1 {
		t.Errorf("Height = %d, want %d (content + padding)", box.Height, 1+1+1)
	}
}

func TestLayoutBoxWithBorder(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Border:    "single",
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if len(box.Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(box.Children))
	}

	child := box.Children[0]
	// Border adds 1 cell each side; child offset by border
	if child.X != 1 {
		t.Errorf("child X = %d, want 1 (border left)", child.X)
	}
	if child.Y != 1 {
		t.Errorf("child Y = %d, want 1 (border top)", child.Y)
	}

	// Auto-size: content + border on each side
	if box.Width != 2+2 {
		t.Errorf("Width = %d, want %d (content + borders)", box.Width, 2+2)
	}
	if box.Height != 1+2 {
		t.Errorf("Height = %d, want %d (content + borders)", box.Height, 1+2)
	}
	if box.Border != "single" {
		t.Errorf("Border = %q, want %q", box.Border, "single")
	}
}

func TestLayoutBoxWithBorderAndPadding(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Border:    "single",
		Padding:   Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if len(box.Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(box.Children))
	}

	child := box.Children[0]
	// Child offset by border (1) + padding
	if child.X != 1+2 {
		t.Errorf("child X = %d, want %d (border + left padding)", child.X, 1+2)
	}
	if child.Y != 1+1 {
		t.Errorf("child Y = %d, want %d (border + top padding)", child.Y, 1+1)
	}

	// Auto-size: content + padding + border
	wantWidth := 2 + 2 + 2 + 2  // content(2) + padding-left(2) + padding-right(2) + border(2)
	wantHeight := 1 + 1 + 1 + 2 // content(1) + padding-top(1) + padding-bottom(1) + border(2)
	if box.Width != wantWidth {
		t.Errorf("Width = %d, want %d", box.Width, wantWidth)
	}
	if box.Height != wantHeight {
		t.Errorf("Height = %d, want %d", box.Height, wantHeight)
	}
}

func TestLayoutAutoSizedBoxMultipleChildren(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Padding:   Padding{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Border:    "single",
		Children: []*Input{
			{Kind: KindText, Content: "short"},
			{Kind: KindText, Content: "a longer line"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	// Max child width = 13 ("a longer line")
	// + padding left(1) + padding right(1) + border(2) = 17
	wantWidth := 13 + 1 + 1 + 2
	if box.Width != wantWidth {
		t.Errorf("Width = %d, want %d", box.Width, wantWidth)
	}

	// Sum child heights = 2
	// + padding top(1) + padding bottom(1) + border(2) = 6
	wantHeight := 2 + 1 + 1 + 2
	if box.Height != wantHeight {
		t.Errorf("Height = %d, want %d", box.Height, wantHeight)
	}
}

func TestLayoutNestedBoxes(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Padding:   Padding{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Children: []*Input{
			{
				Kind:      KindBox,
				Direction: "column",
				Border:    "single",
				Children: []*Input{
					{Kind: KindText, Content: "inner"},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if len(box.Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(box.Children))
	}

	inner := box.Children[0]
	// Inner box is offset by outer padding
	if inner.X != 1 {
		t.Errorf("inner X = %d, want 1 (outer padding left)", inner.X)
	}
	if inner.Y != 1 {
		t.Errorf("inner Y = %d, want 1 (outer padding top)", inner.Y)
	}

	// Inner box: "inner" (5) + border (2) = 7 wide, 1 + border (2) = 3 tall
	if inner.Width != 7 {
		t.Errorf("inner Width = %d, want 7", inner.Width)
	}
	if inner.Height != 3 {
		t.Errorf("inner Height = %d, want 3", inner.Height)
	}

	// Inner box's child (text) is offset by border
	if len(inner.Children) != 1 {
		t.Fatalf("inner.Children = %d, want 1", len(inner.Children))
	}
	text := inner.Children[0]
	if text.X != 1 {
		t.Errorf("text X = %d, want 1 (inner border)", text.X)
	}
	if text.Y != 1 {
		t.Errorf("text Y = %d, want 1 (inner border)", text.Y)
	}

	// Outer box: inner(7) + padding(1+1) = 9 wide, inner(3) + padding(1+1) = 5 tall
	if box.Width != 9 {
		t.Errorf("outer Width = %d, want 9", box.Width)
	}
	if box.Height != 5 {
		t.Errorf("outer Height = %d, want 5", box.Height)
	}
}

func TestLayoutBorderBoxFixedSize(t *testing.T) {
	// Given
	// Border-box model: border is included in fixed size, not added
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedWidth:  20,
		FixedHeight: 10,
		Border:      "single",
		Padding:     Padding{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	// Fixed size should be the total size (border-box)
	if box.Width != 20 {
		t.Errorf("Width = %d, want 20", box.Width)
	}
	if box.Height != 10 {
		t.Errorf("Height = %d, want 10", box.Height)
	}

	// Child offset: border(1) + padding(1) = 2 from each side
	child := box.Children[0]
	if child.X != 2 {
		t.Errorf("child X = %d, want 2 (border + padding)", child.X)
	}
	if child.Y != 2 {
		t.Errorf("child Y = %d, want 2 (border + padding)", child.Y)
	}
}

func TestLayoutStylePassthrough(t *testing.T) {
	// Given
	s := render.Style{FG: render.Color{Name: "red"}, Bold: true}
	input := &Input{
		Kind:    KindText,
		Content: "styled",
		Style:   s,
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Style != s {
		t.Errorf("Style = %+v, want %+v", box.Style, s)
	}
}

func TestLayoutStylePassthroughBox(t *testing.T) {
	// Given
	boxStyle := render.Style{BG: render.Color{Name: "blue"}}
	childStyle := render.Style{FG: render.Color{Name: "green"}}
	input := &Input{
		Kind:  KindBox,
		Style: boxStyle,
		Children: []*Input{
			{Kind: KindText, Content: "hi", Style: childStyle},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Style != boxStyle {
		t.Errorf("box Style = %+v, want %+v", box.Style, boxStyle)
	}
	if box.Children[0].Style != childStyle {
		t.Errorf("child Style = %+v, want %+v", box.Children[0].Style, childStyle)
	}
}
