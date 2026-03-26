package main

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type ResponsiveProps struct {
}

func NewResponsive(props ResponsiveProps) *tui.Component {
	width := tui.Env[int]("width")
	height := tui.Env[int]("height")

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			tui.Quit()
			return
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Terminal: %vx%v", width.Get(), height.Get()),
		Style: render.Style{
			FG:   render.Color{Name: "yellow"},
			Bold: true,
		},
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		Overflow:  "auto",
		MinWidth:  48,
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Padding:   layout.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Style: render.Style{
					FG: render.Color{Name: "cyan"},
				},
				Children: []*layout.Input{
					{
						Kind:    layout.KindText,
						Content: "Sumi Responsive Demo",
						Style: render.Style{
							FG:   render.Color{Name: "green"},
							Bold: true,
						},
					},
					node0,
					{
						Kind:    layout.KindText,
						Content: "Resize your terminal to see this update! Press q to quit.",
						Style: render.Style{
							FG:  render.Color{Name: "cyan"},
							Dim: true,
						},
					},
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("Terminal: %vx%v", width.Get(), height.Get())
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
