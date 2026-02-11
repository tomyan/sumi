package layout

import "math"

// HitTestScroll finds the deepest scrollable box containing the point (x, y).
// When multiple scrollable boxes overlap, the one with the highest z-index wins.
// Returns the index into a flat list of scrollable boxes (depth-first order),
// or -1 if no scrollable box contains the point.
func HitTestScroll(tree *Box, x, y int) int {
	idx := -1
	bestZ := math.MinInt
	counter := 0
	hitTestScrollRecursive(tree, x, y, &idx, &bestZ, &counter)
	return idx
}

func hitTestScrollRecursive(box *Box, x, y int, bestIdx *int, bestZ *int, counter *int) {
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		if !containsPoint(child, x, y) {
			countScrollable(child, counter)
			continue
		}
		if isScrollable(child) {
			if child.ZIndex >= *bestZ {
				*bestIdx = *counter
				*bestZ = child.ZIndex
			}
			*counter++
		}
		hitTestScrollRecursive(child, x, y, bestIdx, bestZ, counter)
	}
}

// countScrollable counts all scrollable boxes in a subtree.
func countScrollable(box *Box, counter *int) {
	if isScrollable(box) {
		*counter++
	}
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		countScrollable(child, counter)
	}
}

func containsPoint(box *Box, x, y int) bool {
	return x >= box.X && x < box.X+box.Width && y >= box.Y && y < box.Y+box.Height
}

func isScrollable(box *Box) bool {
	return box.ContentHeight > 0
}
