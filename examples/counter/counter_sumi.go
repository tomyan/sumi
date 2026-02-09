package main

import (
	"fmt"
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
)

func Run() {
	count := 0

	dirty := true
	increment := func() {
		count = count + 1
		dirty = true
	}

	var prevBuf *render.Buffer
	doRender := func() {
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
							Content: fmt.Sprintf("Count: %v", count),
						},
						{
							Kind:    layout.KindText,
							Content: "Press any key to increment, q to quit",
						},
					},
				},
			},
		}
		tree := layout.Layout(root, 80, 24)
		buf := render.NewBuffer(80, 24)
		renderTree(buf, tree)
		if prevBuf != nil {
			buf.RenderTo(os.Stdout)
		} else {
			buf.RenderTo(os.Stdout)
		}
		prevBuf = buf
		dirty = false
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(os.Stdout)
	defer render.ExitAlternateScreen(os.Stdout)

	doRender()

	for {
		key, err := input.ReadKey(os.Stdin)
		if err != nil || key == 'q' {
			break
		}
		increment()
		if dirty {
			doRender()
		}
	}
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
