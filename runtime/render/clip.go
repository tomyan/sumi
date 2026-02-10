package render

// Clip defines a rectangular clipping region in buffer coordinates.
type Clip struct {
	Top, Left, Bottom, Right int // inclusive bounds
}

// Contains returns true if the cell at (row, col) is inside the clip region.
func (c *Clip) Contains(row, col int) bool {
	return row >= c.Top && row <= c.Bottom && col >= c.Left && col <= c.Right
}

// Intersect returns the intersection of two clip regions, or nil if they don't overlap.
func (c *Clip) Intersect(other *Clip) *Clip {
	top := max(c.Top, other.Top)
	left := max(c.Left, other.Left)
	bottom := min(c.Bottom, other.Bottom)
	right := min(c.Right, other.Right)
	if top > bottom || left > right {
		return nil
	}
	return &Clip{Top: top, Left: left, Bottom: bottom, Right: right}
}
