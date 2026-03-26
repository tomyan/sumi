package main

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type AppProps struct {
}

func NewApp(props AppProps) *tui.Component {
	name := signal.New("")

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			tui.Quit()
			return
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("You typed: %v", name.Get()),
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
						Content: "Text Input Demo",
						Style: render.Style{
							FG:   render.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Type to enter your name",
						Style: render.Style{
							FG:  render.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Name:",
						Style: render.Style{
							FG:   render.Color{Name: "yellow"},
							Bold: true,
						},
					},
					node0,
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("You typed: %v", name.Get())
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
