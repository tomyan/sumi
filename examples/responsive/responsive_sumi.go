package main

import (
	"fmt"
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

func Run() {
	width, height := term.GetSize(int(os.Stdin.Fd()))
	dirty := true

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		root := &layout.Input{
			Kind:      layout.KindBox,
			Direction: "column",
			Children: []*layout.Input{
				{
					Kind:    layout.KindBox,
					Padding: layout.ParsePadding("1 2"),
					Border:  "single",
					Style: render.Style{
						FG: render.Color{Name: "cyan"},
					},
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: "Sumi Responsive Demo",
							Style: render.Style{
								FG:   render.Color{Name: "green"},
								Bold: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: fmt.Sprintf("Terminal: %vx%v", width, height),
							Style: render.Style{
								FG:   render.Color{Name: "yellow"},
								Bold: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: "Resize your terminal to see this update! Press q to quit.",
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
					},
				},
			},
		}
		tree := layout.Layout(root, termW, termH)
		if prevTree == nil || termW != prevW || termH != prevH {
			buf := render.NewBuffer(termW, termH)
			renderTree(buf, tree)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
			changes := layout.DiffTrees(prevTree, tree)
			layout.ApplyChanges(os.Stdout, changes)
		}
		prevTree = tree
		prevW = termW
		prevH = termH
		dirty = false
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(os.Stdout)
	defer render.ExitAlternateScreen(os.Stdout)

	keyCh := make(chan rune)
	go func() {
		for {
			key, err := input.ReadKey(os.Stdin)
			if err != nil {
				close(keyCh)
				return
			}
			keyCh <- key
		}
	}()

	resizeCh, stopResize := term.WatchResize()
	defer stopResize()

	doRender()

	for {
		select {
		case key, ok := <-keyCh:
			if !ok || key == 'q' {
				return
			}
		case <-resizeCh:
			width, height = term.GetSize(int(os.Stdin.Fd()))
			dirty = true
		}
		if dirty {
			doRender()
		}
	}
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
