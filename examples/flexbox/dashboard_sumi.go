package main

import (
	"os"

	sumi "github.com/tomyan/sumi/runtime/prelude"
)

func Run() {
	var app *sumi.App
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "root",
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				Tag:       "box",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"header"},
						Attrs:     map[string]string{"class": "header"},
						Justify:   "center",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
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
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press q to quit",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"panels"},
						Attrs:     map[string]string{"class": "panels"},
						Direction: "row",
						Gap:       1,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								Tag:       "box",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Left Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "This panel uses flex-grow",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "to fill available space.",
									},
								},
							},
							{
								Kind:      sumi.KindBox,
								Tag:       "box",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Right Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "Both panels share the",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "width equally.",
									},
								},
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"footer"},
						Attrs:     map[string]string{"class": "footer"},
						Direction: "row",
						Justify:   "space-between",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"status"},
								Attrs:   map[string]string{"class": "status"},
								Content: "Ready",
								Style: sumi.Style{
									FG: sumi.Color{Name: "green"},
								},
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"version"},
								Attrs:   map[string]string{"class": "version"},
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
		Tag:       "root",
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				Tag:       "box",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"header"},
						Attrs:     map[string]string{"class": "header"},
						Justify:   "center",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
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
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press q to quit",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"panels"},
						Attrs:     map[string]string{"class": "panels"},
						Direction: "row",
						Gap:       1,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								Tag:       "box",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Left Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "This panel uses flex-grow",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "to fill available space.",
									},
								},
							},
							{
								Kind:      sumi.KindBox,
								Tag:       "box",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								FlexGrow:  1,
								Padding:   sumi.ParsePadding("0 1"),
								Border:    "single",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Right Panel",
										Style: sumi.Style{
											FG:   sumi.Color{Name: "yellow"},
											Bold: true,
										},
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "Both panels share the",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "width equally.",
									},
								},
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"footer"},
						Attrs:     map[string]string{"class": "footer"},
						Direction: "row",
						Justify:   "space-between",
						Padding:   sumi.ParsePadding("0 2"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"status"},
								Attrs:   map[string]string{"class": "status"},
								Content: "Ready",
								Style: sumi.Style{
									FG: sumi.Color{Name: "green"},
								},
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"version"},
								Attrs:   map[string]string{"class": "version"},
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
