package layout

import (
	"bytes"
	"strings"
	"testing"
)

func TestMapInputToBoxSimpleTree(t *testing.T) {
	// Given — a simple input tree and its laid-out box tree
	textInput := &Input{Kind: KindText, Content: "hello"}
	rootInput := &Input{
		Kind:     KindBox,
		Children: []*Input{textInput},
	}
	tree := Layout(rootInput, 80, 24)

	// When
	m := MapInputToBox(rootInput, tree)

	// Then — root and text node should be mapped
	if m[rootInput] != tree {
		t.Error("expected root input to map to root box")
	}
	if m[textInput] == nil {
		t.Error("expected text input to map to a box")
	}
	if m[textInput].Content != "hello" {
		t.Errorf("mapped box Content = %q, want %q", m[textInput].Content, "hello")
	}
}

func TestMapInputToBoxWithNilChildren(t *testing.T) {
	// Given — display:none produces nil placeholders
	visible := &Input{Kind: KindText, Content: "visible"}
	hidden := &Input{Kind: KindText, Content: "hidden", Display: "none"}
	rootInput := &Input{
		Kind:     KindBox,
		Children: []*Input{visible, hidden},
	}
	tree := Layout(rootInput, 80, 24)

	// When
	m := MapInputToBox(rootInput, tree)

	// Then — visible node mapped, hidden node not (nil box)
	if m[visible] == nil {
		t.Error("expected visible input to be mapped")
	}
	if _, ok := m[hidden]; ok {
		t.Error("hidden input should not be in map (nil box)")
	}
}

func TestMapInputToBoxNestedTree(t *testing.T) {
	// Given — nested boxes
	leaf := &Input{Kind: KindText, Content: "leaf"}
	inner := &Input{Kind: KindBox, Children: []*Input{leaf}}
	rootInput := &Input{
		Kind:     KindBox,
		Children: []*Input{inner},
	}
	tree := Layout(rootInput, 80, 24)

	// When
	m := MapInputToBox(rootInput, tree)

	// Then
	if m[leaf] == nil {
		t.Error("expected leaf to be mapped")
	}
	if m[leaf].Content != "leaf" {
		t.Errorf("leaf box Content = %q, want %q", m[leaf].Content, "leaf")
	}
}

func TestDirectWriteTextSameLength(t *testing.T) {
	// Given — a box at position (5, 3) with content "abc"
	box := &Box{X: 5, Y: 3, Width: 3, Height: 1, Content: "abc"}
	var buf bytes.Buffer

	// When — write new content of same length
	ok := DirectWriteText(&buf, box, "xyz", "abc")

	// Then
	if !ok {
		t.Fatal("expected DirectWriteText to succeed for same-length content")
	}
	output := buf.String()
	// Should contain cursor positioning (ESC[row;colH format, 1-indexed)
	if !strings.Contains(output, "\x1b[4;6H") {
		t.Errorf("expected cursor move to row=4 col=6, got %q", output)
	}
	if !strings.Contains(output, "xyz") {
		t.Errorf("expected new content 'xyz' in output, got %q", output)
	}
}

func TestDirectWriteTextDifferentLength(t *testing.T) {
	// Given — different length content
	box := &Box{X: 5, Y: 3, Width: 3, Height: 1, Content: "abc"}
	var buf bytes.Buffer

	// When
	ok := DirectWriteText(&buf, box, "abcd", "abc")

	// Then — should fail (fall back to full path)
	if ok {
		t.Error("expected DirectWriteText to fail for different-length content")
	}
}

func TestDirectWriteTextWrappedLines(t *testing.T) {
	// Given — box with wrapped lines (Lines is non-nil)
	box := &Box{X: 0, Y: 0, Width: 10, Height: 2, Content: "hello world", Lines: []string{"hello", "world"}}
	var buf bytes.Buffer

	// When
	ok := DirectWriteText(&buf, box, "hello earth", "hello world")

	// Then — should fail (wrapped text needs relayout)
	if ok {
		t.Error("expected DirectWriteText to fail for wrapped text")
	}
}

func TestDirectWriteTextNilBox(t *testing.T) {
	// Given — nil box
	var buf bytes.Buffer

	// When
	ok := DirectWriteText(&buf, nil, "new", "old")

	// Then
	if ok {
		t.Error("expected DirectWriteText to fail for nil box")
	}
}
