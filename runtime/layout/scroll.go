package layout

// ScrollState tracks the scroll position of a scrollable container.
type ScrollState struct {
	ScrollY int
	ScrollX int
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
