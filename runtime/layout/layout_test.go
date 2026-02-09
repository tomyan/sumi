package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestParsePadding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Padding
	}{
		{
			name:  "empty string returns zero padding",
			input: "",
			want:  Padding{0, 0, 0, 0},
		},
		{
			name:  "single value applies to all sides",
			input: "1",
			want:  Padding{1, 1, 1, 1},
		},
		{
			name:  "two values: top/bottom and left/right",
			input: "1 2",
			want:  Padding{1, 2, 1, 2},
		},
		{
			name:  "four values: top right bottom left",
			input: "1 2 3 4",
			want:  Padding{1, 2, 3, 4},
		},
		{
			name:  "larger values",
			input: "10",
			want:  Padding{10, 10, 10, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePadding(tt.input)
			if got != tt.want {
				t.Errorf("ParsePadding(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLayoutSingleTextNode(t *testing.T) {
	input := &Input{
		Kind:    KindText,
		Content: "hello",
	}

	box := Layout(input, 80, 24)

	if box.Width != 5 {
		t.Errorf("Width = %d, want 5", box.Width)
	}
	if box.Height != 1 {
		t.Errorf("Height = %d, want 1", box.Height)
	}
	if box.X != 0 {
		t.Errorf("X = %d, want 0", box.X)
	}
	if box.Y != 0 {
		t.Errorf("Y = %d, want 0", box.Y)
	}
	if box.Content != "hello" {
		t.Errorf("Content = %q, want %q", box.Content, "hello")
	}
}

func TestLayoutColumnTwoTextNodes(t *testing.T) {
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindText, Content: "second"},
		},
	}

	box := Layout(input, 80, 24)

	if len(box.Children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(box.Children))
	}

	first := box.Children[0]
	if first.X != 0 || first.Y != 0 {
		t.Errorf("first child position = (%d, %d), want (0, 0)", first.X, first.Y)
	}
	if first.Width != 5 {
		t.Errorf("first child Width = %d, want 5", first.Width)
	}
	if first.Height != 1 {
		t.Errorf("first child Height = %d, want 1", first.Height)
	}

	second := box.Children[1]
	if second.X != 0 || second.Y != 1 {
		t.Errorf("second child position = (%d, %d), want (0, 1)", second.X, second.Y)
	}
	if second.Width != 6 {
		t.Errorf("second child Width = %d, want 6", second.Width)
	}

	// Auto-sized parent: width = max child width, height = sum child heights
	if box.Width != 6 {
		t.Errorf("parent Width = %d, want 6 (max of children)", box.Width)
	}
	if box.Height != 2 {
		t.Errorf("parent Height = %d, want 2 (sum of children)", box.Height)
	}
}

func TestLayoutBoxFixedSize(t *testing.T) {
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
	}

	box := Layout(input, 80, 24)

	if box.Width != 20 {
		t.Errorf("Width = %d, want 20", box.Width)
	}
	if box.Height != 10 {
		t.Errorf("Height = %d, want 10", box.Height)
	}
}

func TestLayoutBoxWithPadding(t *testing.T) {
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Padding:   Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	box := Layout(input, 80, 24)

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
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Border:    "single",
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	box := Layout(input, 80, 24)

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
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Border:    "single",
		Padding:   Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	box := Layout(input, 80, 24)

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

	box := Layout(input, 80, 24)

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

	box := Layout(input, 80, 24)

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

func TestLayoutTextNodeFixedSize(t *testing.T) {
	input := &Input{
		Kind:        KindText,
		Content:     "hello",
		FixedWidth:  10,
		FixedHeight: 3,
	}

	box := Layout(input, 80, 24)

	if box.Width != 10 {
		t.Errorf("Width = %d, want 10", box.Width)
	}
	if box.Height != 3 {
		t.Errorf("Height = %d, want 3", box.Height)
	}
}

func TestLayoutBorderBoxFixedSize(t *testing.T) {
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

	box := Layout(input, 80, 24)

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

func TestLayoutEmptyBox(t *testing.T) {
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
	}

	box := Layout(input, 80, 24)

	if box.Width != 0 {
		t.Errorf("Width = %d, want 0", box.Width)
	}
	if box.Height != 0 {
		t.Errorf("Height = %d, want 0", box.Height)
	}
}

func TestLayoutDefaultDirectionIsColumn(t *testing.T) {
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindText, Content: "second"},
		},
	}

	box := Layout(input, 80, 24)

	// Default direction should be column, so children stack vertically
	second := box.Children[1]
	if second.Y != 1 {
		t.Errorf("second child Y = %d, want 1 (column stacking)", second.Y)
	}
}

func TestLayoutColumnThreeChildren(t *testing.T) {
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bb"},
			{Kind: KindText, Content: "c"},
		},
	}

	box := Layout(input, 80, 24)

	if box.Width != 3 {
		t.Errorf("Width = %d, want 3 (max child width)", box.Width)
	}
	if box.Height != 3 {
		t.Errorf("Height = %d, want 3 (sum of children)", box.Height)
	}

	// Check Y positions
	for i, wantY := range []int{0, 1, 2} {
		if box.Children[i].Y != wantY {
			t.Errorf("child[%d].Y = %d, want %d", i, box.Children[i].Y, wantY)
		}
	}
}

func TestLayoutStylePassthrough(t *testing.T) {
	s := render.Style{FG: render.Color{Name: "red"}, Bold: true}
	input := &Input{
		Kind:    KindText,
		Content: "styled",
		Style:   s,
	}
	box := Layout(input, 80, 24)
	if box.Style != s {
		t.Errorf("Style = %+v, want %+v", box.Style, s)
	}
}

func TestLayoutStylePassthroughBox(t *testing.T) {
	boxStyle := render.Style{BG: render.Color{Name: "blue"}}
	childStyle := render.Style{FG: render.Color{Name: "green"}}
	input := &Input{
		Kind:  KindBox,
		Style: boxStyle,
		Children: []*Input{
			{Kind: KindText, Content: "hi", Style: childStyle},
		},
	}
	box := Layout(input, 80, 24)
	if box.Style != boxStyle {
		t.Errorf("box Style = %+v, want %+v", box.Style, boxStyle)
	}
	if box.Children[0].Style != childStyle {
		t.Errorf("child Style = %+v, want %+v", box.Children[0].Style, childStyle)
	}
}
