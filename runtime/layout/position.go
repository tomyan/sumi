package layout

// HiddenFromLayout reports whether a node is excluded from layout:
// display:none from the cascade, or the runtime Hidden override.
func (in *Input) HiddenFromLayout() bool {
	return in.Display == "none" || in.Hidden
}

// filterVisible partitions children into visible and returns a mapping.
// Returns the visible children and a slice of original indices for each visible child.
func filterVisible(children []*Input) ([]*Input, []int) {
	var visible []*Input
	var indices []int
	for i, c := range children {
		if !c.HiddenFromLayout() {
			visible = append(visible, c)
			indices = append(indices, i)
		}
	}
	return visible, indices
}

// spliceChildren creates a full-length children slice with nil placeholders
// for hidden elements. visibleBoxes[i] is placed at indices[i] in the result.
func spliceChildren(total int, visibleBoxes []*Box, indices []int) []*Box {
	result := make([]*Box, total)
	for i, idx := range indices {
		result[idx] = visibleBoxes[i]
	}
	return result
}

// isPositioned returns true if the element is taken out of normal flow.
func isPositioned(pos string) bool {
	return pos == "absolute" || pos == "fixed"
}

// partitionPositioned separates visible children into flow and positioned groups.
// Returns flow children/indices and positioned children/indices (relative to the visible slice).
func partitionPositioned(visible []*Input) (flow []*Input, flowIdx []int, pos []*Input, posIdx []int) {
	for i, c := range visible {
		if isPositioned(c.Position) {
			pos = append(pos, c)
			posIdx = append(posIdx, i)
		} else {
			flow = append(flow, c)
			flowIdx = append(flowIdx, i)
		}
	}
	return
}

// mergePartitioned combines flow and positioned boxes back into visible-order.
func mergePartitioned(flowBoxes []*Box, flowIdx []int, posBoxes []*Box, posIdx []int, total int) []*Box {
	result := make([]*Box, total)
	for i, idx := range flowIdx {
		result[idx] = flowBoxes[i]
	}
	for i, idx := range posIdx {
		result[idx] = posBoxes[i]
	}
	return result
}

// layoutPositionedChildren lays out absolute/fixed children relative to parent content area.
func layoutPositionedChildren(children []*Input, offsetX, offsetY, parentW, parentH int) []*Box {
	boxes := make([]*Box, len(children))
	for i, child := range children {
		boxes[i] = layoutAbsolute(child, offsetX, offsetY, parentW, parentH)
	}
	return boxes
}

// layoutAbsolute positions an absolute child relative to the parent content area.
// When both top and bottom (or left and right) are set without a fixed size,
// the element stretches to fill the gap.
func layoutAbsolute(child *Input, offsetX, offsetY, parentW, parentH int) *Box {
	childW := parentW
	childH := parentH

	// Determine stretch dimensions when opposing offsets are set
	if child.Top != 0 || child.Bottom != 0 {
		if child.FixedHeight == 0 && child.Top != 0 && child.Bottom != 0 {
			childH = parentH - child.Top - child.Bottom
		}
	}
	if child.Left != 0 || child.Right != 0 {
		if child.FixedWidth == 0 && child.Left != 0 && child.Right != 0 {
			childW = parentW - child.Left - child.Right
		}
	}

	box := layoutNode(child, childW, childH)

	// Apply stretch when opposing offsets are both set and no fixed size
	if child.FixedHeight == 0 && child.Top != 0 && child.Bottom != 0 {
		box.Height = parentH - child.Top - child.Bottom
	}
	if child.FixedWidth == 0 && child.Left != 0 && child.Right != 0 {
		box.Width = parentW - child.Left - child.Right
	}

	// Position: top/left take priority over bottom/right
	if child.Left != 0 {
		box.X = offsetX + child.Left
	} else if child.Right != 0 {
		box.X = offsetX + parentW - child.Right - box.Width
	} else {
		box.X = offsetX
	}

	if child.Top != 0 {
		box.Y = offsetY + child.Top
	} else if child.Bottom != 0 {
		box.Y = offsetY + parentH - child.Bottom - box.Height
	} else {
		box.Y = offsetY
	}

	return box
}

// repositionFixed walks the tree and repositions any fixed box to viewport-relative
// coordinates using its stored offsets. Called as a post-pass after absolutePositions.
func repositionFixed(box *Box, viewW, viewH int) {
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		if child.Position == "fixed" {
			// Position: top/left take priority over bottom/right
			if child.Left != 0 {
				child.X = child.Left
			} else if child.Right != 0 {
				child.X = viewW - child.Right - child.Width
			} else {
				child.X = 0
			}
			if child.Top != 0 {
				child.Y = child.Top
			} else if child.Bottom != 0 {
				child.Y = viewH - child.Bottom - child.Height
			} else {
				child.Y = 0
			}
			// Stretch when opposing offsets are both set
			if child.Left != 0 && child.Right != 0 {
				child.Width = viewW - child.Left - child.Right
				// Re-check X in case stretch affects it
			}
			if child.Top != 0 && child.Bottom != 0 {
				child.Height = viewH - child.Top - child.Bottom
			}
		}
		repositionFixed(child, viewW, viewH)
	}
}

// applyRelativeOffsets shifts relatively-positioned boxes by their offsets.
// The parent's size is computed from flow positions before this is called,
// matching CSS behavior where relative offsets are purely visual.
func applyRelativeOffsets(boxes []*Box) {
	for _, b := range boxes {
		if b == nil || b.Position != "relative" {
			continue
		}
		// Top wins over Bottom; Left wins over Right
		if b.Top != 0 {
			b.Y += b.Top
		} else if b.Bottom != 0 {
			b.Y -= b.Bottom
		}
		if b.Left != 0 {
			b.X += b.Left
		} else if b.Right != 0 {
			b.X -= b.Right
		}
	}
}
