package main

import (
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type ModalProps struct {
}

func NewModal(props ModalProps) *tui.Component {
	showModal := signal.New(false)

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
			showModal.Set(!showModal.Get())
		}
	}

	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
	}

	signal.Effect(func() {
		root.Children = func() []*layout.Input {
			var cs []*layout.Input
			cs = append(cs, &layout.Input{
				Kind:      layout.KindBox,
				Padding:   layout.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					{
						Kind:    layout.KindText,
						Content: "Modal Demo",
						Style: render.Style{
							FG:   render.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Press any key to toggle modal, q to quit",
						Style: render.Style{
							FG:  render.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Background content here",
					},
				},
			})
			if showModal.Get() {
				cs = append(cs, &layout.Input{
					Kind:        layout.KindBox,
					FixedWidth:  40,
					FixedHeight: 8,
					Padding:     layout.ParsePadding("1 2"),
					Border:      "single",
					Position:    "fixed",
					Top:         5,
					Left:        10,
					ZIndex:      2,
					CursorCol:   -1,
					CursorRow:   -1,
					Style: render.Style{
						FG: render.Color{Name: "yellow"},
						BG: render.Color{Name: "black"},
					},
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: "Modal Dialog",
							Style: render.Style{
								FG:   render.Color{Name: "yellow"},
								Bold: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: "This is a fixed-position modal overlay.",
						},
						{
							Kind:    layout.KindText,
							Content: "Press any key to close",
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
					},
				})
			}
			return cs
		}()
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
