package render

// ThumbSize returns the scrollbar thumb height in cells, proportional to viewport/content.
// Minimum size is 1 cell. If content fits in viewport, returns viewportHeight.
func ThumbSize(contentHeight, viewportHeight int) int {
	if contentHeight <= viewportHeight {
		return viewportHeight
	}
	size := viewportHeight * viewportHeight / contentHeight
	if size < 1 {
		size = 1
	}
	return size
}

// ThumbPosition returns the row offset of the thumb within the scrollbar track.
func ThumbPosition(scrollY, contentHeight, viewportHeight int) int {
	maxScroll := contentHeight - viewportHeight
	if maxScroll <= 0 {
		return 0
	}
	thumbSize := ThumbSize(contentHeight, viewportHeight)
	trackSpace := viewportHeight - thumbSize
	return scrollY * trackSpace / maxScroll
}

// DrawHorizontalScrollbar draws a horizontal scrollbar at row y, from column x to x+width-1.
// Uses ░ for the track and █ for the thumb.
func DrawHorizontalScrollbar(buf *Buffer, x, y, width, contentWidth, scrollX int, style Style) {
	thumbSize := ThumbSize(contentWidth, width)
	thumbPos := ThumbPosition(scrollX, contentWidth, width)

	for col := 0; col < width; col++ {
		if col >= thumbPos && col < thumbPos+thumbSize {
			buf.SetStyledCell(y, x+col, '█', style)
		} else {
			buf.SetStyledCell(y, x+col, '░', style)
		}
	}
}

// DrawScrollbar draws a vertical scrollbar at column x, from row y to y+height-1.
// Uses ░ for the track and █ for the thumb.
func DrawScrollbar(buf *Buffer, x, y, height, contentHeight, scrollY int, style Style) {
	thumbSize := ThumbSize(contentHeight, height)
	thumbPos := ThumbPosition(scrollY, contentHeight, height)

	for row := 0; row < height; row++ {
		if row >= thumbPos && row < thumbPos+thumbSize {
			buf.SetStyledCell(y+row, x, '█', style)
		} else {
			buf.SetStyledCell(y+row, x, '░', style)
		}
	}
}
