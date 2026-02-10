package layout

import "testing"

func TestScrollStateInitialValues(t *testing.T) {
	// Given
	ss := &ScrollState{}

	// Then
	if ss.ScrollY != 0 {
		t.Errorf("ScrollY = %d, want 0", ss.ScrollY)
	}
	if ss.ScrollX != 0 {
		t.Errorf("ScrollX = %d, want 0", ss.ScrollX)
	}
}

func TestScrollStateClampYClampsToZero(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: -5}

	// When
	ss.ClampY(10, 5)

	// Then
	if ss.ScrollY != 0 {
		t.Errorf("ScrollY = %d, want 0 (clamped to min)", ss.ScrollY)
	}
}

func TestScrollStateClampYClampsToMax(t *testing.T) {
	// Given — content=10, viewport=5, max scroll = 5
	ss := &ScrollState{ScrollY: 100}

	// When
	ss.ClampY(10, 5)

	// Then
	if ss.ScrollY != 5 {
		t.Errorf("ScrollY = %d, want 5 (clamped to max)", ss.ScrollY)
	}
}

func TestScrollStateClampYNoScrollNeeded(t *testing.T) {
	// Given — content fits in viewport
	ss := &ScrollState{ScrollY: 3}

	// When
	ss.ClampY(3, 5)

	// Then — content fits, so scroll = 0
	if ss.ScrollY != 0 {
		t.Errorf("ScrollY = %d, want 0 (content fits)", ss.ScrollY)
	}
}

func TestScrollStateScrollDown(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: 0}

	// When
	ss.ScrollDown(10, 5)

	// Then
	if ss.ScrollY != 1 {
		t.Errorf("ScrollY = %d, want 1", ss.ScrollY)
	}
}

func TestScrollStateScrollDownClamped(t *testing.T) {
	// Given — already at max
	ss := &ScrollState{ScrollY: 5}

	// When
	ss.ScrollDown(10, 5)

	// Then — should stay at max
	if ss.ScrollY != 5 {
		t.Errorf("ScrollY = %d, want 5 (clamped)", ss.ScrollY)
	}
}

func TestScrollStateScrollUp(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: 3}

	// When
	ss.ScrollUp()

	// Then
	if ss.ScrollY != 2 {
		t.Errorf("ScrollY = %d, want 2", ss.ScrollY)
	}
}

func TestScrollStateScrollUpClamped(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: 0}

	// When
	ss.ScrollUp()

	// Then
	if ss.ScrollY != 0 {
		t.Errorf("ScrollY = %d, want 0 (clamped)", ss.ScrollY)
	}
}

func TestScrollStatePageDown(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: 0}

	// When — viewport=5, jumps by viewport
	ss.PageDown(20, 5)

	// Then
	if ss.ScrollY != 5 {
		t.Errorf("ScrollY = %d, want 5", ss.ScrollY)
	}
}

func TestScrollStatePageUp(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: 10}

	// When
	ss.PageUp(5)

	// Then
	if ss.ScrollY != 5 {
		t.Errorf("ScrollY = %d, want 5", ss.ScrollY)
	}
}

func TestScrollStatePageUpClamped(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollY: 2}

	// When
	ss.PageUp(5)

	// Then
	if ss.ScrollY != 0 {
		t.Errorf("ScrollY = %d, want 0 (clamped)", ss.ScrollY)
	}
}
