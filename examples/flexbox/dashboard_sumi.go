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
				Kind:      sumi.KindBox,
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Justify:   "center",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Sumi Flexbox Dashboard",
								Style: sumi.Style{
									FG:   sumi.Color{Name: "green"},
									Bold: true,
								},
							},
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Press q to quit",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:      sumi.KindBox,
						Direction: "row",
						Gap:       1,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Left Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "This panel uses flex-grow",
									},
									{
										Kind:    sumi.KindText,
										Content: "to fill available space.",
									},
								},
							},
							{
								Kind:      sumi.KindBox,
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Right Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "Both panels share the",
									},
									{
										Kind:    sumi.KindText,
										Content: "width equally.",
									},
								},
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Direction: "row",
						Justify:   "space-between",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Ready",
								Style: sumi.Style{
									FG: sumi.Color{Name: "green"},
								},
							},
							{
								Kind:    sumi.KindText,
								Content: "sumi v0.1",
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
				Kind:      sumi.KindBox,
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Justify:   "center",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Sumi Flexbox Dashboard",
								Style: sumi.Style{
									FG:   sumi.Color{Name: "green"},
									Bold: true,
								},
							},
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Press q to quit",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:      sumi.KindBox,
						Direction: "row",
						Gap:       1,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Left Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "This panel uses flex-grow",
									},
									{
										Kind:    sumi.KindText,
										Content: "to fill available space.",
									},
								},
							},
							{
								Kind:      sumi.KindBox,
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Content: "Right Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Content: "Both panels share the",
									},
									{
										Kind:    sumi.KindText,
										Content: "width equally.",
									},
								},
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Direction: "row",
						Justify:   "space-between",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Ready",
								Style: sumi.Style{
									FG: sumi.Color{Name: "green"},
								},
							},
							{
								Kind:    sumi.KindText,
								Content: "sumi v0.1",
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
