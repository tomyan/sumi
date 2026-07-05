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
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
								Content: "Sumi Flexbox Dashboard",
							},
						},
					},
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press q to quit",
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"panels"},
						Attrs:     map[string]string{"class": "panels"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								Tag:       "box",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Left Panel",
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
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Right Panel",
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
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"status"},
								Attrs:   map[string]string{"class": "status"},
								Content: "Ready",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"version"},
								Attrs:   map[string]string{"class": "version"},
								Content: "sumi v0.1",
							},
						},
					},
				},
			},
		},
	}
	stylesheet := sumi.MustParseStylesheet(".header {\n\tborder: single;\n\tjustify-content: center;\n\tpadding: 0 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.panels {\n\tflex-direction: row;\n\tgap: 1;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.panel-title {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n.footer {\n\tborder: single;\n\tflex-direction: row;\n\tjustify-content: space-between;\n\tpadding: 0 2;\n}\n.status {\n\tcolor: green;\n}\n.version {\n\tcolor: cyan;\n\topacity: dim;\n}\n")
	doRender := func() {
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = sumi.GetSize(int(os.Stdin.Fd()))
		}
		sumi.ResolveStyles(root, stylesheet)
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
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
								Content: "Sumi Flexbox Dashboard",
							},
						},
					},
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press q to quit",
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"panels"},
						Attrs:     map[string]string{"class": "panels"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								Tag:       "box",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Left Panel",
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
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Right Panel",
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
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"status"},
								Attrs:   map[string]string{"class": "status"},
								Content: "Ready",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Classes: []string{"version"},
								Attrs:   map[string]string{"class": "version"},
								Content: "sumi v0.1",
							},
						},
					},
				},
			},
		},
	}
	stylesheet := sumi.MustParseStylesheet(".header {\n\tborder: single;\n\tjustify-content: center;\n\tpadding: 0 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.panels {\n\tflex-direction: row;\n\tgap: 1;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.panel-title {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n.footer {\n\tborder: single;\n\tflex-direction: row;\n\tjustify-content: space-between;\n\tpadding: 0 2;\n}\n.status {\n\tcolor: green;\n}\n.version {\n\tcolor: cyan;\n\topacity: dim;\n}\n")
	doRender := func() {
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = sumi.GetSize(int(os.Stdin.Fd()))
		}
		sumi.ResolveStyles(root, stylesheet)
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
