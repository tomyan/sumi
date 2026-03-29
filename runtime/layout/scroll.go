package layout

// ScrollState tracks the scroll position of a scrollable container.
// ContentHeight and ViewportHeight are populated by the layout engine
// when the ScrollState is attached to an Input via the Scroll field.
type ScrollState struct {
	ScrollY        int
	ScrollX        int
	ContentHeight  int
	ViewportHeight int
	Follow         bool // when true, layout automatically scrolls to bottom
}

// ClampY ensures ScrollY is within [0, contentHeight - viewportHeight].
func (s *ScrollState) ClampY(contentHeight, viewportHeight int) {
	maxScroll := contentHeight - viewportHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.ScrollY < 0 {
		s.ScrollY = 0
	}
	if s.ScrollY > maxScroll {
		s.ScrollY = maxScroll
	}
}

// ScrollDown moves the scroll position down by one line.
func (s *ScrollState) ScrollDown(contentHeight, viewportHeight int) {
	s.ScrollY++
	s.ClampY(contentHeight, viewportHeight)
}

// ScrollUp moves the scroll position up by one line.
func (s *ScrollState) ScrollUp() {
	s.ScrollY--
	if s.ScrollY < 0 {
		s.ScrollY = 0
	}
}

// PageDown moves the scroll position down by one page (viewport height).
func (s *ScrollState) PageDown(contentHeight, viewportHeight int) {
	s.ScrollY += viewportHeight
	s.ClampY(contentHeight, viewportHeight)
}

// PageUp moves the scroll position up by one page (viewport height).
func (s *ScrollState) PageUp(viewportHeight int) {
	s.ScrollY -= viewportHeight
	if s.ScrollY < 0 {
		s.ScrollY = 0
	}
}

// ClampX ensures ScrollX is within [0, contentWidth - viewportWidth].
func (s *ScrollState) ClampX(contentWidth, viewportWidth int) {
	maxScroll := contentWidth - viewportWidth
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.ScrollX < 0 {
		s.ScrollX = 0
	}
	if s.ScrollX > maxScroll {
		s.ScrollX = maxScroll
	}
}

// ScrollRight moves the scroll position right by one column.
func (s *ScrollState) ScrollRight(contentWidth, viewportWidth int) {
	s.ScrollX++
	s.ClampX(contentWidth, viewportWidth)
}

// ScrollLeft moves the scroll position left by one column.
func (s *ScrollState) ScrollLeft() {
	s.ScrollX--
	if s.ScrollX < 0 {
		s.ScrollX = 0
	}
}

// ScrollToBottom sets the scroll position to show the last entries.
func (s *ScrollState) ScrollToBottom() {
	max := s.ContentHeight - s.ViewportHeight
	if max < 0 {
		max = 0
	}
	s.ScrollY = max
}

// AtBottom returns true if the scroll position is at or past the bottom.
func (s *ScrollState) AtBottom() bool {
	max := s.ContentHeight - s.ViewportHeight
	if max <= 0 {
		return true
	}
	return s.ScrollY >= max
}
