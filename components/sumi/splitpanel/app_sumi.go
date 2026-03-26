package splitpanel

import (
	"os"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

func Run() {
	var app *tui.App
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:           layout.KindBox,
				Direction:      "row",
				BorderCollapse: true,
				CursorCol:      -1,
				CursorRow:      -1,
				Children: []*layout.Input{
					{
						Kind:        layout.KindBox,
						FixedHeight: 5,
						FlexGrow:    1,
						Padding:     layout.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Actual",
						CursorCol:   -1,
						CursorRow:   -1,
					},
					{
						Kind:        layout.KindBox,
						FixedHeight: 5,
						FlexGrow:    1,
						Padding:     layout.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Expected",
						CursorCol:   -1,
						CursorRow:   -1,
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
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		tree := layout.Layout(root, termW, termH)
		buf := render.NewBuffer(termW, termH)
		layout.RenderTree(buf, tree, nil)
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		} else {
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		}
	}

	app = &tui.App{
		OnRender: doRender,
	}
	app.Run()
}

func CreateApp(w, h int) *tui.App {
	var app *tui.App
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:           layout.KindBox,
				Direction:      "row",
				BorderCollapse: true,
				CursorCol:      -1,
				CursorRow:      -1,
				Children: []*layout.Input{
					{
						Kind:        layout.KindBox,
						FixedHeight: 5,
						FlexGrow:    1,
						Padding:     layout.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Actual",
						CursorCol:   -1,
						CursorRow:   -1,
					},
					{
						Kind:        layout.KindBox,
						FixedHeight: 5,
						FlexGrow:    1,
						Padding:     layout.ParsePadding("0 1"),
						Border:      "single",
						BorderTitle: "Expected",
						CursorCol:   -1,
						CursorRow:   -1,
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
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		tree := layout.Layout(root, termW, termH)
		buf := render.NewBuffer(termW, termH)
		layout.RenderTree(buf, tree, nil)
		if app.TestBuffer != nil {
			app.TestBuffer = buf
		} else {
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		}
	}

	app = &tui.App{
		OnRender: doRender,
	}
	app.TestWidth = w
	app.TestHeight = h
	app.TestBuffer = render.NewBuffer(w, h)
	app.Render()
	return app
}
