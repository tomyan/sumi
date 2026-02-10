package render

import "testing"

func TestScrollbarThumbSizeMinimumOne(t *testing.T) {
	// Given — very large content, small viewport
	contentHeight := 1000
	viewportHeight := 5

	// When
	thumbSize := ThumbSize(contentHeight, viewportHeight)

	// Then — minimum thumb size is 1
	if thumbSize < 1 {
		t.Errorf("ThumbSize = %d, want >= 1", thumbSize)
	}
}

func TestScrollbarThumbSizeProportional(t *testing.T) {
	// Given — content=20, viewport=10 → thumb = 10/20 * 10 = 5
	contentHeight := 20
	viewportHeight := 10

	// When
	thumbSize := ThumbSize(contentHeight, viewportHeight)

	// Then
	if thumbSize != 5 {
		t.Errorf("ThumbSize = %d, want 5", thumbSize)
	}
}

func TestScrollbarThumbSizeFullViewport(t *testing.T) {
	// Given — content fits in viewport
	contentHeight := 5
	viewportHeight := 10

	// When
	thumbSize := ThumbSize(contentHeight, viewportHeight)

	// Then — thumb fills entire track
	if thumbSize != viewportHeight {
		t.Errorf("ThumbSize = %d, want %d (full viewport)", thumbSize, viewportHeight)
	}
}

func TestScrollbarThumbPositionAtTop(t *testing.T) {
	// Given — scrollY=0
	scrollY := 0
	contentHeight := 20
	viewportHeight := 10

	// When
	pos := ThumbPosition(scrollY, contentHeight, viewportHeight)

	// Then
	if pos != 0 {
		t.Errorf("ThumbPosition = %d, want 0", pos)
	}
}

func TestScrollbarThumbPositionAtBottom(t *testing.T) {
	// Given — scrollY at max (content-viewport)
	contentHeight := 20
	viewportHeight := 10
	scrollY := contentHeight - viewportHeight // 10, max scroll

	// When
	pos := ThumbPosition(scrollY, contentHeight, viewportHeight)
	thumbSize := ThumbSize(contentHeight, viewportHeight)

	// Then — thumb should be at bottom: pos + thumbSize = viewportHeight
	if pos+thumbSize != viewportHeight {
		t.Errorf("ThumbPosition + ThumbSize = %d, want %d (at bottom)", pos+thumbSize, viewportHeight)
	}
}

func TestDrawScrollbar(t *testing.T) {
	// Given
	buf := NewBuffer(20, 10)
	style := Style{}

	// When — draw scrollbar at column 19, rows 0-9, content=20, scroll=0
	DrawScrollbar(buf, 19, 0, 10, 20, 0, style)

	// Then — track characters should be present
	trackFound := false
	thumbFound := false
	for row := 0; row < 10; row++ {
		ch := buf.Cell(row, 19).Ch
		if ch == '░' {
			trackFound = true
		}
		if ch == '█' {
			thumbFound = true
		}
	}
	if !trackFound {
		t.Error("expected track characters (░) in scrollbar")
	}
	if !thumbFound {
		t.Error("expected thumb characters (█) in scrollbar")
	}
}

func TestDrawScrollbarThumbPosition(t *testing.T) {
	// Given — scrollbar at top (scrollY=0)
	buf := NewBuffer(20, 10)

	// When
	DrawScrollbar(buf, 19, 0, 10, 20, 0, Style{})

	// Then — thumb should be at the top
	if ch := buf.Cell(0, 19); ch.Ch != '█' {
		t.Errorf("Cell(0, 19).Ch = %c, want '█' (thumb at top)", ch.Ch)
	}
}

func TestDrawScrollbarThumbAtBottom(t *testing.T) {
	// Given — scrolled to bottom
	buf := NewBuffer(20, 10)
	maxScroll := 10 // content=20, viewport=10

	// When
	DrawScrollbar(buf, 19, 0, 10, 20, maxScroll, Style{})

	// Then — last cell should be thumb
	if ch := buf.Cell(9, 19); ch.Ch != '█' {
		t.Errorf("Cell(9, 19).Ch = %c, want '█' (thumb at bottom)", ch.Ch)
	}
}
