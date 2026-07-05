package splitpanel

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
				Classes:   []string{"root"},
				Attrs:     map[string]string{"class": "root", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Actual", "class": "panel", "height": "5"},
						FixedHeight: 5,
						BorderTitle: "Actual",
						CursorCol:   -1,
						CursorRow:   -1,
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Expected", "class": "panel", "height": "5"},
						FixedHeight: 5,
						BorderTitle: "Expected",
						CursorCol:   -1,
						CursorRow:   -1,
					},
				},
			},
		},
	}
	stylesheet := sumi.MustParseStylesheet(".root {\n\tborder-collapse: collapse;\n\tflex-direction: row;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n")
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
				Classes:   []string{"root"},
				Attrs:     map[string]string{"class": "root", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Actual", "class": "panel", "height": "5"},
						FixedHeight: 5,
						BorderTitle: "Actual",
						CursorCol:   -1,
						CursorRow:   -1,
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Expected", "class": "panel", "height": "5"},
						FixedHeight: 5,
						BorderTitle: "Expected",
						CursorCol:   -1,
						CursorRow:   -1,
					},
				},
			},
		},
	}
	stylesheet := sumi.MustParseStylesheet(".root {\n\tborder-collapse: collapse;\n\tflex-direction: row;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n")
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
