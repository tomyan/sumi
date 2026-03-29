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

func TestScrollStateClampXClampsToZero(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollX: -5}

	// When
	ss.ClampX(20, 10)

	// Then
	if ss.ScrollX != 0 {
		t.Errorf("ScrollX = %d, want 0 (clamped to min)", ss.ScrollX)
	}
}

func TestScrollStateClampXClampsToMax(t *testing.T) {
	// Given — contentWidth=20, viewportWidth=10, max scroll = 10
	ss := &ScrollState{ScrollX: 100}

	// When
	ss.ClampX(20, 10)

	// Then
	if ss.ScrollX != 10 {
		t.Errorf("ScrollX = %d, want 10 (clamped to max)", ss.ScrollX)
	}
}

func TestScrollStateClampXNoScrollNeeded(t *testing.T) {
	// Given — content fits in viewport
	ss := &ScrollState{ScrollX: 3}

	// When
	ss.ClampX(5, 10)

	// Then — content fits, so scroll = 0
	if ss.ScrollX != 0 {
		t.Errorf("ScrollX = %d, want 0 (content fits)", ss.ScrollX)
	}
}

func TestScrollStateScrollRight(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollX: 0}

	// When
	ss.ScrollRight(20, 10)

	// Then
	if ss.ScrollX != 1 {
		t.Errorf("ScrollX = %d, want 1", ss.ScrollX)
	}
}

func TestScrollStateScrollRightClamped(t *testing.T) {
	// Given — already at max
	ss := &ScrollState{ScrollX: 10}

	// When
	ss.ScrollRight(20, 10)

	// Then — should stay at max
	if ss.ScrollX != 10 {
		t.Errorf("ScrollX = %d, want 10 (clamped)", ss.ScrollX)
	}
}

func TestScrollStateScrollLeft(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollX: 5}

	// When
	ss.ScrollLeft()

	// Then
	if ss.ScrollX != 4 {
		t.Errorf("ScrollX = %d, want 4", ss.ScrollX)
	}
}

func TestScrollStateScrollLeftClamped(t *testing.T) {
	// Given
	ss := &ScrollState{ScrollX: 0}

	// When
	ss.ScrollLeft()

	// Then
	if ss.ScrollX != 0 {
		t.Errorf("ScrollX = %d, want 0 (clamped)", ss.ScrollX)
	}
}

func TestScrollToBottom(t *testing.T) {
	// Given a scroll state not at the bottom
	ss := &ScrollState{ScrollY: 0, ContentHeight: 20, ViewportHeight: 10}

	// When
	ss.ScrollToBottom()

	// Then scroll should be at the bottom
	if ss.ScrollY != 10 {
		t.Errorf("ScrollY = %d, want 10", ss.ScrollY)
	}
}

func TestScrollToBottomContentFits(t *testing.T) {
	// Given content that fits in the viewport
	ss := &ScrollState{ScrollY: 0, ContentHeight: 5, ViewportHeight: 10}

	// When
	ss.ScrollToBottom()

	// Then scroll stays at 0
	if ss.ScrollY != 0 {
		t.Errorf("ScrollY = %d, want 0", ss.ScrollY)
	}
}

func TestAtBottomTrue(t *testing.T) {
	// Given scroll is at the bottom
	ss := &ScrollState{ScrollY: 10, ContentHeight: 20, ViewportHeight: 10}

	// Then
	if !ss.AtBottom() {
		t.Error("expected AtBottom() = true")
	}
}

func TestAtBottomFalse(t *testing.T) {
	// Given scroll is not at the bottom
	ss := &ScrollState{ScrollY: 5, ContentHeight: 20, ViewportHeight: 10}

	// Then
	if ss.AtBottom() {
		t.Error("expected AtBottom() = false")
	}
}

func TestLayoutPopulatesScrollState(t *testing.T) {
	// Given a scroll container with a ScrollState attached
	ss := &ScrollState{}
	input := &Input{
		Kind:     KindBox,
		Overflow: "auto",
		Scroll:   ss,
		Children: []*Input{
			{Kind: KindText, Content: "line1"},
			{Kind: KindText, Content: "line2"},
			{Kind: KindText, Content: "line3"},
			{Kind: KindText, Content: "line4"},
			{Kind: KindText, Content: "line5"},
		},
	}

	// When laid out with height 3
	Layout(input, 20, 3)

	// Then ScrollState should have the box dimensions
	if ss.ContentHeight != 5 {
		t.Errorf("ContentHeight = %d, want 5", ss.ContentHeight)
	}
	if ss.ViewportHeight != 3 {
		t.Errorf("ViewportHeight = %d, want 3", ss.ViewportHeight)
	}
}

func TestLayoutFollowScrollsToBottom(t *testing.T) {
	// Given a scroll state with Follow=true
	ss := &ScrollState{Follow: true}
	input := &Input{
		Kind:     KindBox,
		Overflow: "auto",
		Scroll:   ss,
		Children: []*Input{
			{Kind: KindText, Content: "line1"},
			{Kind: KindText, Content: "line2"},
			{Kind: KindText, Content: "line3"},
			{Kind: KindText, Content: "line4"},
			{Kind: KindText, Content: "line5"},
		},
	}

	// When laid out with height 3
	box := Layout(input, 20, 3)

	// Then scroll should be at the bottom
	if box.ScrollY != 2 {
		t.Errorf("box.ScrollY = %d, want 2 (5 content - 3 viewport)", box.ScrollY)
	}
	if ss.ScrollY != 2 {
		t.Errorf("ss.ScrollY = %d, want 2", ss.ScrollY)
	}
}

func TestLayoutAppliesScrollStateToBox(t *testing.T) {
	// Given a scroll state scrolled to position 2
	ss := &ScrollState{ScrollY: 2}
	input := &Input{
		Kind:     KindBox,
		Overflow: "auto",
		Scroll:   ss,
		Children: []*Input{
			{Kind: KindText, Content: "line1"},
			{Kind: KindText, Content: "line2"},
			{Kind: KindText, Content: "line3"},
			{Kind: KindText, Content: "line4"},
			{Kind: KindText, Content: "line5"},
		},
	}

	// When
	box := Layout(input, 20, 3)

	// Then the box's ScrollY should match the ScrollState
	if box.ScrollY != 2 {
		t.Errorf("box.ScrollY = %d, want 2", box.ScrollY)
	}
}
