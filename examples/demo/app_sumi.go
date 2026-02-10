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
	renderTree(buf, tree)
	buf.RenderTo(os.Stdout)
	bufio.NewScanner(os.Stdin).Scan()
	render.ExitAlternateScreen(os.Stdout)
}

func renderTree(buf *render.Buffer, box *layout.Box) {
	if box.Border != "" && box.Border != "none" {
		buf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)
	}
	if box.Lines != nil {
		for i, line := range box.Lines {
			buf.WriteStyledText(box.Y+i, box.X, line, box.Style)
		}
	} else if box.Content != "" {
		buf.WriteStyledText(box.Y, box.X, box.Content, box.Style)
	}
	for _, child := range box.Children {
		renderTree(buf, child)
	}
}
