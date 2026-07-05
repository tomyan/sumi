package layout

// HitTestPath returns the chain of Inputs from root to the deepest node
// whose Box contains (x, y). Ancestors stay on the path even when the
// point lies outside their own bounds (fixed-position descendants have
// viewport-relative coordinates). Returns nil when nothing is hit.
func HitTestPath(input *Input, box *Box, x, y int) []*Input {
	if input == nil || box == nil {
		return nil
	}

	// Check children depth-first — deepest match wins.
	for i, child := range input.Children {
		if child == nil {
			continue
		}
		var childBox *Box
		if i < len(box.Children) && box.Children[i] != nil {
			childBox = box.Children[i]
		}
		if path := HitTestPath(child, childBox, x, y); path != nil {
			return append([]*Input{input}, path...)
		}
	}

	if boxOccupies(box, x, y) {
		return []*Input{input}
	}
	return nil
}

// boxOccupies reports whether the box's own content covers the cell.
// Fragment boxes occupy only their line rectangles; union boxes (inline
// elements, display:contents placeholders) occupy nothing themselves —
// only their children do.
func boxOccupies(box *Box, x, y int) bool {
	if !containsPoint(box, x, y) {
		return false
	}
	if box.UnionBox {
		for _, child := range box.Children {
			if child != nil && boxOccupies(child, x, y) {
				return true
			}
		}
		return false
	}
	if box.Fragments == nil {
		return true
	}
	for _, f := range box.Fragments {
		if y == box.Y+f.Y && x >= box.X+f.X && x < box.X+f.X+runeLen(f.Text) {
			return true
		}
	}
	return false
}

// PathTo returns the chain of Inputs from root to target, or nil when
// target is not in the tree.
func PathTo(root, target *Input) []*Input {
	if root == nil {
		return nil
	}
	if root == target {
		return []*Input{root}
	}
	for _, child := range root.Children {
		if p := PathTo(child, target); p != nil {
			return append([]*Input{root}, p...)
		}
	}
	return nil
}

// HasClickHandlers returns true if any node in the tree handles "click".
func HasClickHandlers(input *Input) bool {
	if input == nil {
		return false
	}
	if input.On["click"] != nil {
		return true
	}
	for _, child := range input.Children {
		if HasClickHandlers(child) {
			return true
		}
	}
	return false
}
