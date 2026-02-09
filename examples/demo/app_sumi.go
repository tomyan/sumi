package main

import (
	"bufio"
	"os"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
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
	tree := layout.Layout(root, 80, 24)
	buf := render.NewBuffer(80, 24)
	render.EnterAlternateScreen(os.Stdout)
	renderTree(buf, tree)
	buf.RenderTo(os.Stdout)
	bufio.NewScanner(os.Stdin).Scan()
	render.ExitAlternateScreen(os.Stdout)
}

func renderTree(buf *render.Buffer, box *layout.Box) {
	if box.Border != "" && box.Border != "none" {
		buf.DrawBorder(box.Y, box.X, box.Width, box.Height, box.Border)
	}
	if box.Content != "" {
		buf.WriteText(box.Y, box.X, box.Content)
	}
	for _, child := range box.Children {
		renderTree(buf, child)
	}
}
