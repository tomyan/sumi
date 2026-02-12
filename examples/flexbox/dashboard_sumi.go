package main

import (
	"os"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

func Run() {
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					{
						Kind:      layout.KindBox,
						Justify:   "center",
						Padding:   layout.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
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
						CursorCol: -1,
						CursorRow: -1,
						Children: []*layout.Input{
							{
								Kind:      layout.KindBox,
								FlexGrow:  1,
								Padding:   layout.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
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
								Kind:      layout.KindBox,
								FlexGrow:  1,
								Padding:   layout.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
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
						CursorCol: -1,
						CursorRow: -1,
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

	app := &tui.App{
		OnRender: doRender,
	}
	app.Run()
}
