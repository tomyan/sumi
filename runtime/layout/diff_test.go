package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestDiffTreesSameTree(t *testing.T) {
	// Given
	tree := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		Children: []*Box{
			{X: 0, Y: 0, Width: 5, Height: 1, Content: "hello"},
		},
	}

	// When
	changes, _ := DiffTrees(tree, tree)

	// Then — no changes
	if len(changes) != 0 {
		t.Errorf("len(changes) = %d, want 0", len(changes))
	}
}

func TestDiffTreesContentChanged(t *testing.T) {
	// Given
	old := &Box{
		X: 0, Y: 0, Width: 10, Height: 3,
		Children: []*Box{
			{X: 1, Y: 1, Width: 3, Height: 1, Content: "old"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 10, Height: 3,
		Children: []*Box{
			{X: 1, Y: 1, Width: 3, Height: 1, Content: "new"},
		},
	}

	// When
	changes, _ := DiffTrees(old, new)

	// Then — one change for the text node
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if changes[0].Old.Content != "old" {
		t.Errorf("Old.Content = %q, want %q", changes[0].Old.Content, "old")
	}
	if changes[0].New.Content != "new" {
		t.Errorf("New.Content = %q, want %q", changes[0].New.Content, "new")
	}
}

func TestDiffTreesPositionShifted(t *testing.T) {
	// Given
	old := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		Children: []*Box{
			{X: 0, Y: 0, Width: 5, Height: 1, Content: "text"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		Children: []*Box{
			{X: 2, Y: 1, Width: 5, Height: 1, Content: "text"},
		},
	}

	// When
	changes, _ := DiffTrees(old, new)

	// Then — change for the moved node
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if changes[0].Old.X != 0 || changes[0].New.X != 2 {
		t.Errorf("X: old=%d new=%d, want old=0 new=2", changes[0].Old.X, changes[0].New.X)
	}
}

func TestDiffTreesSizeChanged(t *testing.T) {
	// Given
	old := &Box{X: 0, Y: 0, Width: 10, Height: 5}
	new := &Box{X: 0, Y: 0, Width: 12, Height: 5}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
}

func TestDiffTreesStyleChanged(t *testing.T) {
	// Given
	old := &Box{
		X: 0, Y: 0, Width: 5, Height: 1,
		Content: "hi",
		Style:   render.Style{FG: render.Color{Name: "red"}},
	}
	new := &Box{
		X: 0, Y: 0, Width: 5, Height: 1,
		Content: "hi",
		Style:   render.Style{FG: render.Color{Name: "blue"}},
	}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
}

func TestDiffTreesBorderChanged(t *testing.T) {
	// Given
	old := &Box{X: 0, Y: 0, Width: 10, Height: 5, Border: "single"}
	new := &Box{X: 0, Y: 0, Width: 10, Height: 5, Border: "none"}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
}

func TestDiffTreesNestedChanges(t *testing.T) {
	// Given — deep tree, only leaf changed
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{
				X: 1, Y: 1, Width: 18, Height: 8,
				Children: []*Box{
					{X: 2, Y: 2, Width: 5, Height: 1, Content: "old"},
				},
			},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{
				X: 1, Y: 1, Width: 18, Height: 8,
				Children: []*Box{
					{X: 2, Y: 2, Width: 5, Height: 1, Content: "new"},
				},
			},
		},
	}

	// When
	changes, _ := DiffTrees(old, new)

	// Then — only the leaf node changed
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if changes[0].Old.Content != "old" {
		t.Errorf("changed node Content = %q, want %q", changes[0].Old.Content, "old")
	}
}

func TestDiffDetectsScrollYChange(t *testing.T) {
	// Given
	old := &Box{X: 0, Y: 0, Width: 10, Height: 5, ScrollY: 0}
	new := &Box{X: 0, Y: 0, Width: 10, Height: 5, ScrollY: 3}

	// When
	changes, scrollChanged := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if !scrollChanged {
		t.Error("expected scrollChanged=true when ScrollY differs")
	}
}

func TestDiffDetectsScrollXChange(t *testing.T) {
	// Given
	old := &Box{X: 0, Y: 0, Width: 10, Height: 5, ScrollX: 0}
	new := &Box{X: 0, Y: 0, Width: 10, Height: 5, ScrollX: 5}

	// When
	changes, scrollChanged := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if !scrollChanged {
		t.Error("expected scrollChanged=true when ScrollX differs")
	}
}

func TestDiffScrollChangedFalseWhenNoScroll(t *testing.T) {
	// Given — content change but no scroll change
	old := &Box{X: 0, Y: 0, Width: 10, Height: 5, Content: "old"}
	new := &Box{X: 0, Y: 0, Width: 10, Height: 5, Content: "new"}

	// When
	_, scrollChanged := DiffTrees(old, new)

	// Then
	if scrollChanged {
		t.Error("expected scrollChanged=false when no scroll offset differs")
	}
}

func TestDiffScrollChangedDetectsNestedScroll(t *testing.T) {
	// Given — scroll change in nested child
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 20, Height: 10, ScrollY: 0},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 20, Height: 10, ScrollY: 5},
		},
	}

	// When
	_, scrollChanged := DiffTrees(old, new)

	// Then
	if !scrollChanged {
		t.Error("expected scrollChanged=true for nested scroll change")
	}
}

func TestDiffDetectsScrollbarChange(t *testing.T) {
	// Given
	old := &Box{X: 0, Y: 0, Width: 10, Height: 5, NeedsScrollbar: false}
	new := &Box{X: 0, Y: 0, Width: 10, Height: 5, NeedsScrollbar: true}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
}

func TestDiffDetectsCollapsedChange(t *testing.T) {
	// Given
	old := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		Border:    "single",
		Collapsed: render.CollapsedEdges{},
	}
	new := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		Border:    "single",
		Collapsed: render.CollapsedEdges{Top: true},
	}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
}

func TestDiffDetectsBorderTitleChange(t *testing.T) {
	// Given
	old := &Box{X: 0, Y: 0, Width: 20, Height: 5, Border: "single", BorderTitle: "Old"}
	new := &Box{X: 0, Y: 0, Width: 20, Height: 5, Border: "single", BorderTitle: "New"}

	// When
	changes, _ := DiffTrees(old, new)

	// Then
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
}

func TestDiffTreesNilOldIsFullRedraw(t *testing.T) {
	// Given
	new := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		Children: []*Box{
			{X: 0, Y: 0, Width: 5, Height: 1, Content: "hi"},
		},
	}

	// When
	changes, _ := DiffTrees(nil, new)

	// Then — all nodes reported as additions
	if len(changes) < 1 {
		t.Errorf("expected at least 1 change for nil old tree")
	}
	for _, c := range changes {
		if c.Old != nil {
			t.Errorf("Old should be nil for additions")
		}
	}
}
