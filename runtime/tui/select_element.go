package tui

import (
	"unicode/utf8"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

// selectOptions returns a select's option elements, including options
// nested inside optgroups.
func selectOptions(sel *layout.Input) []*layout.Input {
	var opts []*layout.Input
	var walk func(n *layout.Input)
	walk = func(n *layout.Input) {
		for _, c := range n.Children {
			if c == nil {
				continue
			}
			if c.Tag == "option" {
				opts = append(opts, c)
				continue
			}
			if c.Tag == "optgroup" {
				walk(c)
			}
		}
	}
	walk(sel)
	return opts
}

// selectedOptionIndex finds the option carrying the selected attribute,
// defaulting to the first.
func selectedOptionIndex(opts []*layout.Input) int {
	for i, o := range opts {
		if v, ok := o.Attrs["selected"]; ok && v != "false" {
			return i
		}
	}
	return 0
}

func setSelectedOption(opts []*layout.Input, idx int) {
	for i, o := range opts {
		if o.Attrs == nil {
			o.Attrs = map[string]string{}
		}
		if i == idx {
			o.Attrs["selected"] = "true"
		} else {
			delete(o.Attrs, "selected")
		}
	}
}

// optionValue is the option's value attribute, falling back to its label.
func optionValue(o *layout.Input) string {
	if v, ok := o.Attrs["value"]; ok {
		return v
	}
	return o.Content
}

// syncSelectElement hides the options from layout and projects the
// selected label plus a ▾ marker into the implicit child, sizing the
// element to its longest option.
func syncSelectElement(n *layout.Input) {
	opts := selectOptions(n)
	maxWidth := 0
	for _, o := range opts {
		o.Display = "none"
		if w := utf8.RuneCountInString(o.Content); w > maxWidth {
			maxWidth = w
		}
	}
	label := ""
	if len(opts) > 0 {
		label = opts[selectedOptionIndex(opts)].Content
	}
	ensureValueChild(n).Content = label + " ▾"
	if _, authorWidth := n.Attrs["width"]; !authorWidth {
		n.FixedWidth = maxWidth + 2
	}
	n.CursorCol = -1
}

// moveSelect shifts the selection by delta with wraparound and
// dispatches a change event carrying the new value.
func moveSelect(comp *Component, path []*layout.Input, sel *layout.Input, delta int, evt input.Event) {
	opts := selectOptions(sel)
	if len(opts) == 0 {
		return
	}
	idx := (selectedOptionIndex(opts) + delta + len(opts)) % len(opts)
	setSelectedOption(opts, idx)
	syncSelectElement(sel)
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "change", Key: evt,
		Data: map[string]any{"value": optionValue(opts[idx])}})
}

// selectKeydown is the keydown default action for a focused select:
// arrows move with wraparound, Space and Enter advance.
func selectKeydown(comp *Component, evt input.Event) bool {
	path := layout.FocusablePath(comp.Tree, comp.FocusIndex)
	if len(path) == 0 {
		return false
	}
	target := path[len(path)-1]
	if target.Tag != "select" {
		return false
	}
	switch {
	case evt.Kind == input.EventSpecial && evt.Special == input.KeyUp:
		moveSelect(comp, path, target, -1, evt)
	case evt.Kind == input.EventSpecial && evt.Special == input.KeyDown:
		moveSelect(comp, path, target, 1, evt)
	case evt.Kind == input.EventKey && !evt.Ctrl && evt.Rune == ' ':
		moveSelect(comp, path, target, 1, evt)
	case evt.Kind == input.EventSpecial && evt.Special == input.KeyEnter:
		moveSelect(comp, path, target, 1, evt)
	default:
		return false
	}
	return true
}
