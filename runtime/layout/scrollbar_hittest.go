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
	return scrollFromDrag(dragRow, contentHeight, viewportHeight)
}

// ScrollXFromDrag computes the scrollX value for a thumb drag position.
// dragCol is the column offset from the scrollbar left where the thumb's left edge is.
func ScrollXFromDrag(dragCol, contentWidth, viewportWidth int) int {
	return scrollFromDrag(dragCol, contentWidth, viewportWidth)
}

// scrollFromDrag computes a scroll offset from a drag position (works for both axes).
func scrollFromDrag(dragPos, contentSize, viewportSize int) int {
	thumbSize := render.ThumbSize(contentSize, viewportSize)
	trackSpace := viewportSize - thumbSize
	if trackSpace <= 0 {
		return 0
	}
	maxScroll := contentSize - viewportSize
	scroll := dragPos * maxScroll / trackSpace
	if scroll < 0 {
		return 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}
