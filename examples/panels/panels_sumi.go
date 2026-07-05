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
				Classes:   []string{"layout"},
				Attrs:     map[string]string{"class": "layout"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"left-col"},
						Attrs:     map[string]string{"class": "left-col"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"panel"},
								Attrs:       map[string]string{"border-title": "Panel 1", "class": "panel"},
								BorderTitle: "Panel 1",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"title"},
										Attrs:   map[string]string{"class": "title"},
										Content: "Top-left panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "Content goes here",
									},
								},
							},
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"panel"},
								Attrs:       map[string]string{"border-title": "Panel 2", "class": "panel"},
								BorderTitle: "Panel 2",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"title"},
										Attrs:   map[string]string{"class": "title"},
										Content: "Bottom-left panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "More content here",
									},
								},
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Panel 3", "class": "panel"},
						BorderTitle: "Panel 3",
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
								Content: "Right panel",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "This panel spans the full height",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"hint"},
								Attrs:   map[string]string{"class": "hint"},
								Content: "Press q to quit",
							},
						},
					},
				},
			},
		},
	}
	stylesheet := sumi.MustParseStylesheet(".layout {\n\tborder-collapse: collapse;\n\tflex-direction: row;\n}\n.left-col {\n\tborder: single;\n\tborder-collapse: collapse;\n\tflex-grow: 1;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n")
	doRender := func() {
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = sumi.GetSize(int(os.Stdin.Fd()))
		}
		sumi.ResolveStyles(root, stylesheet, termW, termH)
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
				Classes:   []string{"layout"},
				Attrs:     map[string]string{"class": "layout"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"left-col"},
						Attrs:     map[string]string{"class": "left-col"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"panel"},
								Attrs:       map[string]string{"border-title": "Panel 1", "class": "panel"},
								BorderTitle: "Panel 1",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"title"},
										Attrs:   map[string]string{"class": "title"},
										Content: "Top-left panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "Content goes here",
									},
								},
							},
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"panel"},
								Attrs:       map[string]string{"border-title": "Panel 2", "class": "panel"},
								BorderTitle: "Panel 2",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"title"},
										Attrs:   map[string]string{"class": "title"},
										Content: "Bottom-left panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "More content here",
									},
								},
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Panel 3", "class": "panel"},
						BorderTitle: "Panel 3",
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
								Content: "Right panel",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "This panel spans the full height",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"hint"},
								Attrs:   map[string]string{"class": "hint"},
								Content: "Press q to quit",
							},
						},
					},
				},
			},
		},
	}
	stylesheet := sumi.MustParseStylesheet(".layout {\n\tborder-collapse: collapse;\n\tflex-direction: row;\n}\n.left-col {\n\tborder: single;\n\tborder-collapse: collapse;\n\tflex-grow: 1;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n")
	doRender := func() {
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = sumi.GetSize(int(os.Stdin.Fd()))
		}
		sumi.ResolveStyles(root, stylesheet, termW, termH)
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
