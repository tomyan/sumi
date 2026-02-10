package layout

import "testing"

func TestLayoutBorderTitlePassthrough(t *testing.T) {
	// Given
	input := &Input{
		Kind:        KindBox,
		Border:      "single",
		BorderTitle: "My Panel",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 40, 10)

	// Then
	if box.BorderTitle != "My Panel" {
		t.Errorf("BorderTitle = %q, want %q", box.BorderTitle, "My Panel")
	}
}
