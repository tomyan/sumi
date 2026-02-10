package layout

import "github.com/tomyan/sumi/runtime/render"

// ScrollbarHitResult identifies where on the scrollbar a click landed.
type ScrollbarHitResult int

const (
	ScrollbarAboveThumb ScrollbarHitResult = iota
	ScrollbarOnThumb
	ScrollbarBelowThumb
)

// ScrollbarHit determines if a click at a given row offset (relative to
// scrollbar top) is above, on, or below the thumb.
func ScrollbarHit(row, thumbPos, thumbSize, trackHeight int) ScrollbarHitResult {
	if row < thumbPos {
		return ScrollbarAboveThumb
	}
	if row < thumbPos+thumbSize {
		return ScrollbarOnThumb
	}
	return ScrollbarBelowThumb
}

// ScrollYFromDrag computes the scrollY value for a thumb drag position.
// dragRow is the row offset from the scrollbar top where the thumb's top edge is.
func ScrollYFromDrag(dragRow, contentHeight, viewportHeight int) int {
	thumbSize := render.ThumbSize(contentHeight, viewportHeight)
	trackSpace := viewportHeight - thumbSize
	if trackSpace <= 0 {
		return 0
	}
	maxScroll := contentHeight - viewportHeight
	scrollY := dragRow * maxScroll / trackSpace
	if scrollY < 0 {
		return 0
	}
	if scrollY > maxScroll {
		return maxScroll
	}
	return scrollY
}
