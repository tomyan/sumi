package layout

import "testing"

func TestDiffKeyedReorder(t *testing.T) {
	// Given — two keyed children in different order, same content
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "beta", Key: "b"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "beta", Key: "b"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "alpha", Key: "a"},
		},
	}

	// When
	changes := DiffTrees(old, new)

	// Then — no content changes, only position changes from reorder
	for _, c := range changes {
		if c.Old == nil || c.New == nil {
			t.Error("reorder should not produce additions or removals")
		}
		if c.Old.Content != c.New.Content {
			t.Errorf("content should match after keyed reorder: old=%q new=%q",
				c.Old.Content, c.New.Content)
		}
	}
}

func TestDiffKeyedInsertion(t *testing.T) {
	// Given — new list has an extra keyed child
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "beta", Key: "b"},
		},
	}

	// When
	changes := DiffTrees(old, new)

	// Then — the new child "beta" is reported as an addition
	additions := 0
	for _, c := range changes {
		if c.Old == nil && c.New != nil {
			additions++
			if c.New.Content != "beta" {
				t.Errorf("addition Content = %q, want %q", c.New.Content, "beta")
			}
		}
	}
	if additions != 1 {
		t.Errorf("expected 1 addition, got %d (total changes: %d)", additions, len(changes))
	}
}

func TestDiffKeyedRemoval(t *testing.T) {
	// Given — old list has a child that was removed
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "beta", Key: "b"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
		},
	}

	// When
	changes := DiffTrees(old, new)

	// Then — "beta" is reported as a removal
	removals := 0
	for _, c := range changes {
		if c.Old != nil && c.New == nil {
			removals++
			if c.Old.Content != "beta" {
				t.Errorf("removal Content = %q, want %q", c.Old.Content, "beta")
			}
		}
	}
	if removals != 1 {
		t.Errorf("expected 1 removal, got %d (total changes: %d)", removals, len(changes))
	}
}

func TestDiffKeyedModification(t *testing.T) {
	// Given — same keys, but one child's content changed
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "beta", Key: "b"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "gamma", Key: "b"},
		},
	}

	// When
	changes := DiffTrees(old, new)

	// Then — only "b" is reported as modified
	modifications := 0
	for _, c := range changes {
		if c.Old != nil && c.New != nil && c.Old.Content != c.New.Content {
			modifications++
			if c.Old.Content != "beta" || c.New.Content != "gamma" {
				t.Errorf("modification: old=%q new=%q, want old=beta new=gamma",
					c.Old.Content, c.New.Content)
			}
		}
	}
	if modifications != 1 {
		t.Errorf("expected 1 modification, got %d", modifications)
	}
}

func TestDiffUnkeyedFallback(t *testing.T) {
	// Given — no keys, positional diffing applies
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "beta"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "beta"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "alpha"},
		},
	}

	// When
	changes := DiffTrees(old, new)

	// Then — positional: both children appear changed (content swapped)
	contentChanges := 0
	for _, c := range changes {
		if c.Old != nil && c.New != nil && c.Old.Content != c.New.Content {
			contentChanges++
		}
	}
	if contentChanges != 2 {
		t.Errorf("expected 2 content changes for unkeyed reorder, got %d", contentChanges)
	}
}

func TestDiffMixedKeyedUnkeyed(t *testing.T) {
	// Given — some children keyed, some not
	old := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "unkeyed"},
		},
	}
	new := &Box{
		X: 0, Y: 0, Width: 20, Height: 10,
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "alpha", Key: "a"},
			{X: 0, Y: 1, Width: 10, Height: 1, Content: "unkeyed"},
		},
	}

	// When
	changes := DiffTrees(old, new)

	// Then — no changes, mixed keyed/unkeyed produces reasonable result
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical mixed list, got %d", len(changes))
	}
}
