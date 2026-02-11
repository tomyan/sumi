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

	// Walk children: use keyed diffing when keys are present, positional otherwise
	if hasKeyedChildren(old.Children) || hasKeyedChildren(new.Children) {
		diffKeyedChildren(old.Children, new.Children, changes)
	} else {
		diffPositionalChildren(old.Children, new.Children, changes)
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
		if old.Children[i] == nil || new.Children[i] == nil {
			continue
		}
		if HasScrollChanged(old.Children[i], new.Children[i]) {
			return true
		}
	}
	return false
}

// diffPositionalChildren walks old and new children in lockstep by index.
func diffPositionalChildren(oldChildren, newChildren []*Box, changes *[]Change) {
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
		diffNode(oldChild, newChild, changes)
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
func diffKeyedChildren(oldChildren, newChildren []*Box, changes *[]Change) {
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
			diffNode(oldChild, newChild, changes)
			continue
		}
		if oldChild, ok := oldByKey[newChild.Key]; ok {
			matched[newChild.Key] = true
			diffNode(oldChild, newChild, changes)
		} else {
			diffNode(nil, newChild, changes)
		}
	}

	// Report unmatched keyed old children as removals
	for _, oldChild := range oldChildren {
		if oldChild.Key != "" && !matched[oldChild.Key] {
			diffNode(oldChild, nil, changes)
		}
	}
	// Report excess unkeyed old children as removals
	for i := unkeyedIdx; i < len(oldUnkeyed); i++ {
		diffNode(oldUnkeyed[i], nil, changes)
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
