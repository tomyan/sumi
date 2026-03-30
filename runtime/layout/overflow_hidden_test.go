package layout

import (
	"strings"
	"testing"
)

func TestOverflowHiddenFillsAvailableWidth(t *testing.T) {
	rule := strings.Repeat("─", 500)
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{
				Kind:        KindBox,
				FixedHeight: 1,
				Overflow:    "hidden",
				Children: []*Input{
					{Kind: KindText, Content: rule},
				},
			},
		},
	}
	box := Layout(input, 40, 5)
	ruleBox := box.Children[0]
	if ruleBox.Width != 40 {
		t.Errorf("overflow:hidden box Width = %d, want 40", ruleBox.Width)
	}
}
