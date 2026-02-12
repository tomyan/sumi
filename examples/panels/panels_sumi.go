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
		Children: []*layout.Input{
			{
				Kind:           layout.KindBox,
				Direction:      "row",
				BorderCollapse: true,
				Children: []*layout.Input{
					{
						Kind:           layout.KindBox,
						FlexGrow:       1,
						Border:         "single",
						BorderCollapse: true,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Panel 1",
								Children: []*layout.Input{
									{
										Kind:    layout.KindText,
										Content: "Top-left panel",
										Style: render.Style{
											FG:   render.Color{Name: "green"},
											Bold: true,
										},
									},
									{
										Kind:    layout.KindText,
										Content: "Content goes here",
									},
								},
							},
							{
								Kind:        layout.KindBox,
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Panel 2",
								Children: []*layout.Input{
									{
										Kind:    layout.KindText,
										Content: "Bottom-left panel",
										Style: render.Style{
											FG:   render.Color{Name: "green"},
											Bold: true,
										},
									},
									{
										Kind:    layout.KindText,
										Content: "More content here",
									},
								},
							},
						},
					},
					{
						Kind:        layout.KindBox,
						FlexGrow:    1,
						Padding:     layout.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Panel 3",
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Right panel",
								Style: render.Style{
									FG:   render.Color{Name: "green"},
									Bold: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "This panel spans the full height",
							},
							{
								Kind:    layout.KindText,
								Content: "Press q to quit",
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
