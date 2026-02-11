package layout

// HitTestScroll finds the deepest scrollable box containing the point (x, y).
// Returns the index into a flat list of scrollable boxes (depth-first order),
// or -1 if no scrollable box contains the point.
func HitTestScroll(tree *Box, x, y int) int {
	idx := -1
	counter := 0
	hitTestScrollRecursive(tree, x, y, &idx, &counter)
	return idx
}

func hitTestScrollRecursive(box *Box, x, y int, bestIdx *int, counter *int) {
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		if !containsPoint(child, x, y) {
			countScrollable(child, counter)
			continue
		}
		if isScrollable(child) {
			*bestIdx = *counter
			*counter++
		}
		hitTestScrollRecursive(child, x, y, bestIdx, counter)
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
