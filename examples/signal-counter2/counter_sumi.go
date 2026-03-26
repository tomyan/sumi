package main

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type CounterProps struct {
}

func NewCounter(props CounterProps) *tui.Component {
	count := signal.New(0)

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			tui.Quit()
			return
		}
		if evt.Kind == input.EventKey {
			count.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Count: %v", count.Get()),
		Style: render.Style{
			FG:   render.Color{Name: "yellow"},
			Bold: true,
		},
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Padding:   layout.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					{
						Kind:    layout.KindText,
						Content: "Signal Counter",
						Style: render.Style{
							FG:   render.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Press any key to increment, q to quit",
					},
					node0,
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("Count: %v", count.Get())
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
