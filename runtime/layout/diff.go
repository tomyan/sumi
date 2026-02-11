package layout

// Change represents a difference between two corresponding nodes in the layout tree.
// Old is nil for additions; New is nil for removals; both are set for modifications.
type Change struct {
	Old *Box
	New *Box
}

// DiffTrees walks old and new layout trees in lockstep, returning changes
// and whether any scroll offset changed.
func DiffTrees(old, new *Box) ([]Change, bool) {
	if old == nil && new == nil {
		return nil, false
	}
	var changes []Change
	var scrollChanged bool
	diffNode(old, new, &changes, &scrollChanged)
	return changes, scrollChanged
}

func diffNode(old, new *Box, changes *[]Change, scrollChanged *bool) {
	if old == nil && new == nil {
		return
	}
	if old == nil {
		// Addition: new node with no old counterpart
		*changes = append(*changes, Change{Old: nil, New: new})
		for _, child := range new.Children {
			diffNode(nil, child, changes, scrollChanged)
		}
		return
	}
	if new == nil {
		// Removal: old node with no new counterpart
		*changes = append(*changes, Change{Old: old, New: nil})
		return
	}

	// Track scroll changes
	if old.ScrollX != new.ScrollX || old.ScrollY != new.ScrollY {
		*scrollChanged = true
	}

	// Both exist — compare
	if nodeChanged(old, new) {
		*changes = append(*changes, Change{Old: old, New: new})
	}

	// Walk children: use keyed diffing when keys are present, positional otherwise
	if hasKeyedChildren(old.Children) || hasKeyedChildren(new.Children) {
		diffKeyedChildren(old.Children, new.Children, changes, scrollChanged)
	} else {
		diffPositionalChildren(old.Children, new.Children, changes, scrollChanged)
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
	if old.BorderTitle != new.BorderTitle {
		return true
	}
	if old.Collapsed != new.Collapsed {
		return true
	}
	if old.Style != new.Style {
		return true
	}
	if old.Position != new.Position {
		return true
	}
	if old.Top != new.Top || old.Left != new.Left || old.Right != new.Right || old.Bottom != new.Bottom {
		return true
	}
	if old.ZIndex != new.ZIndex {
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

// diffPositionalChildren walks old and new children in lockstep by index.
func diffPositionalChildren(oldChildren, newChildren []*Box, changes *[]Change, scrollChanged *bool) {
	maxChildren := len(oldChildren)
	if len(newChildren) > maxChildren {
		maxChildren = len(newChildren)
	}
	for i := 0; i < maxChildren; i++ {
		var oldChild, newChild *Box
		if i < len(oldChildren) {
			oldChild = oldChildren[i]
		}
		if i < len(newChildren) {
			newChild = newChildren[i]
		}
		diffNode(oldChild, newChild, changes, scrollChanged)
	}
}

// hasKeyedChildren returns true if any child has a non-empty Key.
func hasKeyedChildren(children []*Box) bool {
	for _, c := range children {
		if c.Key != "" {
			return true
		}
	}
	return false
}

// diffKeyedChildren matches old and new children by Key for efficient diffing.
// Keyed children are matched by key; unkeyed children are matched positionally
// against other unkeyed children. Unmatched children are reported as additions/removals.
func diffKeyedChildren(oldChildren, newChildren []*Box, changes *[]Change, scrollChanged *bool) {
	oldByKey := make(map[string]*Box, len(oldChildren))
	var oldUnkeyed []*Box
	for _, c := range oldChildren {
		if c.Key != "" {
			oldByKey[c.Key] = c
		} else {
			oldUnkeyed = append(oldUnkeyed, c)
		}
	}

	matched := make(map[string]bool, len(newChildren))
	unkeyedIdx := 0
	for _, newChild := range newChildren {
		if newChild.Key == "" {
			// Match unkeyed children positionally
			var oldChild *Box
			if unkeyedIdx < len(oldUnkeyed) {
				oldChild = oldUnkeyed[unkeyedIdx]
				unkeyedIdx++
			}
			diffNode(oldChild, newChild, changes, scrollChanged)
			continue
		}
		if oldChild, ok := oldByKey[newChild.Key]; ok {
			matched[newChild.Key] = true
			diffNode(oldChild, newChild, changes, scrollChanged)
		} else {
			diffNode(nil, newChild, changes, scrollChanged)
		}
	}

	// Report unmatched keyed old children as removals
	for _, oldChild := range oldChildren {
		if oldChild.Key != "" && !matched[oldChild.Key] {
			diffNode(oldChild, nil, changes, scrollChanged)
		}
	}
	// Report excess unkeyed old children as removals
	for i := unkeyedIdx; i < len(oldUnkeyed); i++ {
		diffNode(oldUnkeyed[i], nil, changes, scrollChanged)
	}
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
