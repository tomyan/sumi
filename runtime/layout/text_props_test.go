package layout

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// B5: text-align, white-space, text-overflow, text-transform, visibility.

func renderToString(tree *Input, w, h int) []string {
	box := Layout(tree, w, h)
	buf := render.NewBuffer(w, h)
	RenderTree(buf, box, nil)
	var lines []string
	for row := 0; row < h; row++ {
		var b strings.Builder
		for col := 0; col < w; col++ {
			ch := buf.Cell(row, col).Ch
			if ch == 0 {
				ch = ' '
			}
			b.WriteRune(ch)
		}
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}
	return lines
}

func TestTextAlignCenterAndRight(t *testing.T) {
	tree := &Input{Kind: KindBox, FixedWidth: 11, Children: []*Input{
		{Kind: KindText, Content: "hi", TextAlign: "center"},
		{Kind: KindText, Content: "hi", TextAlign: "right"},
	}}
	lines := renderToString(tree, 11, 3)
	if lines[0] != "    hi" {
		t.Errorf("center = %q", lines[0])
	}
	if lines[1] != "         hi" {
		t.Errorf("right = %q", lines[1])
	}
}

func TestWhiteSpaceNowrapSingleLine(t *testing.T) {
	tree := &Input{Kind: KindBox, FixedWidth: 5, Children: []*Input{
		{Kind: KindText, Content: "one two three", WhiteSpace: "nowrap"},
	}}
	box := Layout(tree, 5, 5)
	if got := box.Children[0].Height; got != 1 {
		t.Errorf("nowrap height = %d, want 1", got)
	}
}

func TestWhiteSpacePrePreservesLines(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindText, Content: "a  b\nc", WhiteSpace: "pre"},
	}}
	box := Layout(tree, 40, 5)
	if got := box.Children[0].Height; got != 2 {
		t.Errorf("pre height = %d, want 2", got)
	}
	if got := box.Children[0].Width; got != 4 {
		t.Errorf("pre width = %d, want 4", got)
	}
}

func TestTextOverflowEllipsis(t *testing.T) {
	tree := &Input{Kind: KindBox, FixedWidth: 6, Children: []*Input{
		{Kind: KindText, Content: "overflowing", WhiteSpace: "nowrap",
			TextOverflow: "ellipsis", FixedWidth: 6},
	}}
	lines := renderToString(tree, 10, 2)
	if lines[0] != "overf…" {
		t.Errorf("ellipsis = %q, want overf…", lines[0])
	}
}

func TestTextOverflowEllipsisMiddle(t *testing.T) {
	tree := &Input{Kind: KindBox, FixedWidth: 7, Children: []*Input{
		{Kind: KindText, Content: "abcdefghij", WhiteSpace: "nowrap",
			TextOverflow: "ellipsis-middle", FixedWidth: 7},
	}}
	lines := renderToString(tree, 10, 2)
	if lines[0] != "abc…hij" {
		t.Errorf("ellipsis-middle = %q, want abc…hij", lines[0])
	}
}

func TestTextTransform(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindText, Content: "hello world", TextTransform: "uppercase"},
		{Kind: KindText, Content: "hello world", TextTransform: "capitalize"},
	}}
	lines := renderToString(tree, 20, 3)
	if lines[0] != "HELLO WORLD" {
		t.Errorf("uppercase = %q", lines[0])
	}
	if lines[1] != "Hello World" {
		t.Errorf("capitalize = %q", lines[1])
	}
}

func TestVisibilityHiddenOccupiesSpaceNotPainted(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindText, Content: "ghost", Visibility: "hidden"},
		{Kind: KindText, Content: "real"},
	}}
	lines := renderToString(tree, 10, 3)
	if lines[0] != "" {
		t.Errorf("hidden line painted: %q", lines[0])
	}
	if lines[1] != "real" {
		t.Errorf("second line = %q, want real (space preserved)", lines[1])
	}
}
