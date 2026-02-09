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
							Content: "Sumi Counter",
							Style: render.Style{
								FG:   render.Color{Name: "green"},
								Bold: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: "Press any key to increment, q to quit",
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: fmt.Sprintf("Count: %v", count),
							Style: render.Style{
								FG:   render.Color{Name: "yellow"},
								Bold: true,
							},
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
		buf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)
	}
	if box.Content != "" {
		buf.WriteStyledText(box.Y, box.X, box.Content, box.Style)
	}
	for _, child := range box.Children {
		renderTree(buf, child)
	}
}
