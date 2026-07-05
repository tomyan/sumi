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
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Content: "Welcome to Sumi!",
					},
					{
						Kind:    sumi.KindText,
						Content: "A declarative TTY framework for Go.",
					},
					{
						Kind:      sumi.KindBox,
						Padding:   sumi.ParsePadding("0 1"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Press Enter to exit.",
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
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Content: "Welcome to Sumi!",
					},
					{
						Kind:    sumi.KindText,
						Content: "A declarative TTY framework for Go.",
					},
					{
						Kind:      sumi.KindBox,
						Padding:   sumi.ParsePadding("0 1"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Content: "Press Enter to exit.",
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
