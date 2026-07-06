package main

import (
	"github.com/tomyan/sumi/examples/composition/counter"
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

func main() {
	// Create two counter components with different labels.
	c1 := counter.NewCounter(counter.CounterProps{Label: "Clicks"})
	c2 := counter.NewCounter(counter.CounterProps{Label: "Score"})

	// Build the app tree manually (no parent .sumi file needed).
	root := &layout.Input{
		Kind:      layout.KindBox,
		Border:    "single",
		Padding:   layout.Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		CursorCol: -1,
		CursorRow: -1,
		Style:     render.Style{FG: render.Color{Name: "cyan"}},
		Children: []*layout.Input{
			{Kind: layout.KindText, Content: "Composition Demo"},
			{Kind: layout.KindText, Content: "Press keys in each counter"},
			c1.Tree,
			c2.Tree,
		},
	}

	comp := &tui.Component{
		Tree: root,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventSignal {
				tui.Quit()
				return
			}
			if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
				tui.Quit()
				return
			}
			// Dispatch to first counter for now.
			if c1.OnEvent != nil {
				c1.OnEvent(evt)
			}
		},
	}

	tui.Run(comp)
}
