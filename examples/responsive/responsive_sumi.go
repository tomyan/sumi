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
	var scroll0 layout.ScrollState

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		root := &layout.Input{
			Kind:      layout.KindBox,
			Direction: "column",
			Overflow:  "auto",
			MinWidth:  48,
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
		tree.ScrollY = scroll0.ScrollY
		tree.ScrollX = scroll0.ScrollX
		if prevTree == nil || termW != prevW || termH != prevH {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
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

	eventCh := make(chan input.Event)
	go func() {
		for {
			evt, err := input.ReadEvent(os.Stdin)
			if err != nil {
				close(eventCh)
				return
			}
			eventCh <- evt
		}
	}()

	resizeCh, stopResize := term.WatchResize()
	defer stopResize()

	doRender()

	for {
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return
			}
			if evt.Kind == input.EventKey {
				if evt.Rune == 'q' || evt.Rune == 3 {
					return
				}
			}
			if evt.Kind == input.EventSpecial && prevTree != nil {
				switch evt.Special {
				case input.KeyDown:
					scroll0.ScrollDown(prevTree.ContentHeight, prevTree.Height)
					dirty = true
				case input.KeyUp:
					scroll0.ScrollUp()
					dirty = true
				case input.KeyPgDn:
					scroll0.PageDown(prevTree.ContentHeight, prevTree.Height)
					dirty = true
				case input.KeyPgUp:
					scroll0.PageUp(prevTree.Height)
					dirty = true
				case input.KeyRight:
					scroll0.ScrollRight(prevTree.ContentWidth, prevTree.Width)
					dirty = true
				case input.KeyLeft:
					scroll0.ScrollLeft()
					dirty = true
				}
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
