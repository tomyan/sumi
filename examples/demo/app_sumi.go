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
				Kind:    layout.KindBox,
				Padding: layout.ParsePadding("1 2"),
				Border:  "single",
				Children: []*layout.Input{
					{
						Kind:    layout.KindText,
						Content: "Welcome to Sumi!",
					},
					{
						Kind:    layout.KindText,
						Content: "A declarative TTY framework for Go.",
					},
					{
						Kind:    layout.KindBox,
						Padding: layout.ParsePadding("0 1"),
						Border:  "single",
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Press Enter to exit.",
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
