package main

import (
	"bufio"
	"os"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
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
	termW, termH := term.GetSize(int(os.Stdin.Fd()))
	tree := layout.Layout(root, termW, termH)
	buf := render.NewBuffer(termW, termH)
	render.EnterAlternateScreen(os.Stdout)
	layout.RenderTree(buf, tree, nil)
	buf.RenderTo(os.Stdout)
	bufio.NewScanner(os.Stdin).Scan()
	render.ExitAlternateScreen(os.Stdout)
}
