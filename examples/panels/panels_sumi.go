package main

import (
	"os"

	sumi "github.com/tomyan/sumi/runtime/prelude"
)

func Run() {
	var app *sumi.App
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:           sumi.KindBox,
				Direction:      "row",
				BorderCollapse: true,
				CursorCol:      -1,
				CursorRow:      -1,
				Children: []*sumi.Input{
					{
						Kind:           sumi.KindBox,
						FlexGrow:       1,
						Border:         "single",
						BorderCollapse: true,
						CursorCol:      -1,
						CursorRow:      -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								FlexGrow:    1,
								Padding:     sumi.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Panel 1",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Top-left panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "green"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "Content goes here",
									},
								},
							},
							{
								Kind:        sumi.KindBox,
								FlexGrow:    1,
								Padding:     sumi.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Panel 2",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Bottom-left panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "green"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "More content here",
									},
								},
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						FlexGrow:    1,
						Padding:     sumi.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Panel 3",
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Right panel",
								Style: sumi.Style{
									FG:   sumi.Color{Name: "green"},
									Bold: true,
								},
							},
							{
								Kind:    sumi.KindText,
								Content: "This panel spans the full height",
							},
							{
								Kind:    sumi.KindText,
								Content: "Press q to quit",
								Style: sumi.Style{
									FG:  sumi.Color{Name: "cyan"},
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
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = sumi.GetSize(int(os.Stdin.Fd()))
		}
		tree := sumi.Layout(root, termW, termH)
		buf := sumi.NewBuffer(termW, termH)
		sumi.RenderTree(buf, tree, nil)
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		} else {
			sumi.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		}
	}

	app = &sumi.App{
		OnRender: doRender,
	}
	app.Run()
}

func CreateApp(w, h int) *sumi.App {
	var app *sumi.App
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:           sumi.KindBox,
				Direction:      "row",
				BorderCollapse: true,
				CursorCol:      -1,
				CursorRow:      -1,
				Children: []*sumi.Input{
					{
						Kind:           sumi.KindBox,
						FlexGrow:       1,
						Border:         "single",
						BorderCollapse: true,
						CursorCol:      -1,
						CursorRow:      -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								FlexGrow:    1,
								Padding:     sumi.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Panel 1",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Top-left panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "green"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "Content goes here",
									},
								},
							},
							{
								Kind:        sumi.KindBox,
								FlexGrow:    1,
								Padding:     sumi.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Panel 2",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Bottom-left panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "green"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "More content here",
									},
								},
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						FlexGrow:    1,
						Padding:     sumi.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Panel 3",
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Right panel",
								Style: sumi.Style{
									FG:   sumi.Color{Name: "green"},
									Bold: true,
								},
							},
							{
								Kind:    sumi.KindText,
								Content: "This panel spans the full height",
							},
							{
								Kind:    sumi.KindText,
								Content: "Press q to quit",
								Style: sumi.Style{
									FG:  sumi.Color{Name: "cyan"},
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
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = sumi.GetSize(int(os.Stdin.Fd()))
		}
		tree := sumi.Layout(root, termW, termH)
		buf := sumi.NewBuffer(termW, termH)
		sumi.RenderTree(buf, tree, nil)
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		} else {
			sumi.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		}
	}

	app = &sumi.App{
		OnRender: doRender,
	}
	app.TestWidth = w
	app.TestHeight = h
	app.TestBuffer = sumi.NewBuffer(w, h)
	app.Render()
	return app
}
