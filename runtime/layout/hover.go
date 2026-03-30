package layout

// UpdateHover walks the box tree and sets Hovered on nodes whose bounds
// contain the mouse position.
func UpdateHover(input *Input, box *Box, mouseX, mouseY int) {
	updateHoverInput(input, box, mouseX, mouseY)
}

func updateHoverInput(input *Input, box *Box, mx, my int) {
	if box == nil || input == nil {
		return
	}
	hit := mx >= box.X && mx < box.X+box.Width && my >= box.Y && my < box.Y+box.Height
	input.Hovered = hit && !input.HoverStyle.IsZero()

	for i, child := range input.Children {
		if child == nil {
			continue
		}
		var childBox *Box
		if i < len(box.Children) && box.Children[i] != nil {
			childBox = box.Children[i]
		}
		updateHoverInput(child, childBox, mx, my)
	}
}
