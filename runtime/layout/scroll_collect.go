package layout

// CollectScrollStates returns all ScrollState pointers from the Input tree
// in depth-first order, matching the index order used by HitTestScroll.
func CollectScrollStates(input *Input) []*ScrollState {
	var states []*ScrollState
	collectScrollStates(input, &states)
	return states
}

func collectScrollStates(input *Input, states *[]*ScrollState) {
	if input == nil {
		return
	}
	if input.Scroll != nil {
		*states = append(*states, input.Scroll)
	}
	for _, child := range input.Children {
		collectScrollStates(child, states)
	}
}
