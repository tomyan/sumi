package main

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

func Run() {
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		Children: []*layout.Input{
			{
				Kind: layout.KindBox,
				Children: []*layout.Input{
					{
						Kind:    layout.KindBox,
						Justify: "center",
						Padding: layout.ParsePadding("0 2"),
						Border:  "single",
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Sumi Flexbox Dashboard",
								Style: render.Style{
									FG:   render.Color{Name: "green"},
									Bold: true,
								},
							},
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Press q to quit",
						Style: render.Style{
							FG:  render.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:      layout.KindBox,
						Direction: "row",
						Gap:       1,
						Children: []*layout.Input{
							{
								Kind:     layout.KindBox,
								FlexGrow: 1,
								Padding:  layout.ParsePadding("0 1"),
								Border:   "single",
								Children: []*layout.Input{
									{
										Kind:    layout.KindText,
										Content: "Left Panel",
										Style: render.Style{
											FG:   render.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    layout.KindText,
										Content: "This panel uses flex-grow",
									},
									{
										Kind:    layout.KindText,
										Content: "to fill available space.",
									},
								},
							},
							{
								Kind:     layout.KindBox,
								FlexGrow: 1,
								Padding:  layout.ParsePadding("0 1"),
								Border:   "single",
								Children: []*layout.Input{
									{
										Kind:    layout.KindText,
										Content: "Right Panel",
										Style: render.Style{
											FG:   render.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    layout.KindText,
										Content: "Both panels share the",
									},
									{
										Kind:    layout.KindText,
										Content: "width equally.",
									},
								},
							},
						},
					},
					{
						Kind:      layout.KindBox,
						Direction: "row",
						Justify:   "space-between",
						Padding:   layout.ParsePadding("0 2"),
						Border:    "single",
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Ready",
								Style: render.Style{
									FG: render.Color{Name: "green"},
								},
							},
							{
								Kind:    layout.KindText,
								Content: "sumi v0.1",
								Style: render.Style{
									FG:  render.Color{Name: "cyan"},
									Dim: true,
								},
							},
						},
					},
				},
			},
		},
	}
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		tree := layout.Layout(root, termW, termH)
		buf := render.NewBuffer(termW, termH)
		layout.RenderTree(buf, tree, nil)
		render.ClearScreen(os.Stdout)
		buf.RenderTo(os.Stdout)
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(os.Stdout)
	defer render.ExitAlternateScreen(os.Stdout)

	eventCh := make(chan input.Event)
	go func() {
		for {
			evt, err := input.ReadEvent(os.Stdin)
			if err != nil {
				close(eventCh)
				return
			}
			eventCh <- evt
		}
	}()

	resizeCh, stopResize := term.WatchResize()
	defer stopResize()

	doRender()

	for {
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return
			}
			if evt.Kind == input.EventKey {
				if evt.Rune == 'q' || evt.Rune == 3 {
					return
				}
			}
		case <-resizeCh:
			doRender()
		}
	}
}
