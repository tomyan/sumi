package layout

// Change represents a difference between two corresponding nodes in the layout tree.
// Old is nil for additions; New is nil for removals; both are set for modifications.
type Change struct {
	Old *Box
	New *Box
}

// DiffTrees walks old and new layout trees in lockstep, returning changes.
// If old is nil, all nodes in new are reported as additions.
func DiffTrees(old, new *Box) []Change {
	if old == nil && new == nil {
		return nil
	}
	var changes []Change
	diffNode(old, new, &changes)
	return changes
}

func diffNode(old, new *Box, changes *[]Change) {
	if old == nil && new == nil {
		return
	}
	if old == nil {
		// Addition: new node with no old counterpart
		*changes = append(*changes, Change{Old: nil, New: new})
		for _, child := range new.Children {
			diffNode(nil, child, changes)
		}
		return
	}
	if new == nil {
		// Removal: old node with no new counterpart
		*changes = append(*changes, Change{Old: old, New: nil})
		return
	}

	// Both exist — compare
	if nodeChanged(old, new) {
		*changes = append(*changes, Change{Old: old, New: new})
	}

	// Walk children in lockstep
	maxChildren := len(old.Children)
	if len(new.Children) > maxChildren {
		maxChildren = len(new.Children)
	}
	for i := 0; i < maxChildren; i++ {
		var oldChild, newChild *Box
		if i < len(old.Children) {
			oldChild = old.Children[i]
		}
		if i < len(new.Children) {
			newChild = new.Children[i]
		}
		diffNode(oldChild, newChild, changes)
	}
}

// nodeChanged returns true if any visual property differs between old and new.
func nodeChanged(old, new *Box) bool {
	if old.X != new.X || old.Y != new.Y {
		return true
	}
	if old.Width != new.Width || old.Height != new.Height {
		return true
	}
	if old.Content != new.Content {
		return true
	}
	if old.Border != new.Border {
		return true
	}
	if old.Style != new.Style {
		return true
	}
	if !linesEqual(old.Lines, new.Lines) {
		return true
	}
	if old.ScrollY != new.ScrollY || old.ScrollX != new.ScrollX {
		return true
	}
	if old.NeedsScrollbar != new.NeedsScrollbar {
		return true
	}
	if old.NeedsHorizontalScrollbar != new.NeedsHorizontalScrollbar {
		return true
	}
	return false
}

// HasScrollChanged returns true if any node's scroll offset differs between old and new trees.
func HasScrollChanged(old, new *Box) bool {
	if old == nil || new == nil {
		return false
	}
	if old.ScrollX != new.ScrollX || old.ScrollY != new.ScrollY {
		return true
	}
	for i := 0; i < len(old.Children) && i < len(new.Children); i++ {
		if HasScrollChanged(old.Children[i], new.Children[i]) {
			return true
		}
	}
	return false
}

// linesEqual compares two string slices for equality.
func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
