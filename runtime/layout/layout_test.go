package layout

import (
	"testing"
)

func TestParsePadding(t *testing.T) {
	// Given
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
			// When
			got := ParsePadding(tt.input)

			// Then
			if got != tt.want {
				t.Errorf("ParsePadding(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLayoutSingleTextNode(t *testing.T) {
	// Given
	input := &Input{
		Kind:    KindText,
		Content: "hello",
	}

	// When
	box := Layout(input, 80, 24)

	// Then
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
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindText, Content: "second"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
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
	// Given
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Width != 20 {
		t.Errorf("Width = %d, want 20", box.Width)
	}
	if box.Height != 10 {
		t.Errorf("Height = %d, want 10", box.Height)
	}
}

func TestLayoutEmptyBox(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Width != 0 {
		t.Errorf("Width = %d, want 0", box.Width)
	}
	if box.Height != 0 {
		t.Errorf("Height = %d, want 0", box.Height)
	}
}

func TestLayoutDefaultDirectionIsColumn(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindText, Content: "second"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	// Default direction should be column, so children stack vertically
	second := box.Children[1]
	if second.Y != 1 {
		t.Errorf("second child Y = %d, want 1 (column stacking)", second.Y)
	}
}

func TestLayoutColumnThreeChildren(t *testing.T) {
	// Given
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bb"},
			{Kind: KindText, Content: "c"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
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

func TestLayoutNestedPositionsAreAbsolute(t *testing.T) {
	// Given a column with a bordered box, containing a row with two children.
	// All positions in the resulting tree should be absolute (buffer coordinates).
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "header"},
			{
				Kind:      KindBox,
				Direction: "row",
				Children: []*Input{
					{Kind: KindText, Content: "left"},
					{Kind: KindText, Content: "right"},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	// header is at Y=0
	if box.Children[0].Y != 0 {
		t.Errorf("header Y = %d, want 0", box.Children[0].Y)
	}
	// row is at Y=1 (after header)
	row := box.Children[1]
	if row.Y != 1 {
		t.Errorf("row Y = %d, want 1", row.Y)
	}
	// row's children should have ABSOLUTE Y positions (Y=1, not Y=0)
	if row.Children[0].Y != 1 {
		t.Errorf("left Y = %d, want 1 (absolute, same as row)", row.Children[0].Y)
	}
	if row.Children[1].Y != 1 {
		t.Errorf("right Y = %d, want 1 (absolute, same as row)", row.Children[1].Y)
	}
	// right's X should be absolute
	if row.Children[1].X != 4 {
		t.Errorf("right X = %d, want 4 (after 'left')", row.Children[1].X)
	}
}

func TestLayoutDeeplyNestedPositionsAreAbsolute(t *testing.T) {
	// Given a 3-level deep nesting: column > bordered box > text
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "above"},
			{
				Kind:   KindBox,
				Border: "single",
				Children: []*Input{
					{Kind: KindText, Content: "nested"},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	innerBox := box.Children[1]
	// Inner box is at Y=1 (below "above" text)
	if innerBox.Y != 1 {
		t.Errorf("innerBox Y = %d, want 1", innerBox.Y)
	}
	// Text inside inner box: border=1, so offsetY=1 relative to inner box
	// Absolute Y = innerBox.Y + 1 = 2
	text := innerBox.Children[0]
	if text.Y != 2 {
		t.Errorf("nested text Y = %d, want 2 (absolute: innerBox.Y=1 + border=1)", text.Y)
	}
	// Absolute X = innerBox.X + border = 0 + 1 = 1
	if text.X != 1 {
		t.Errorf("nested text X = %d, want 1 (absolute: innerBox.X=0 + border=1)", text.X)
	}
}

func TestLayoutTextNodeFixedSize(t *testing.T) {
	// Given
	input := &Input{
		Kind:        KindText,
		Content:     "hello",
		FixedWidth:  10,
		FixedHeight: 3,
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Width != 10 {
		t.Errorf("Width = %d, want 10", box.Width)
	}
	if box.Height != 3 {
		t.Errorf("Height = %d, want 3", box.Height)
	}
}
