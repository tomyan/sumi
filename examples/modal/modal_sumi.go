package main

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

func Run() {
	showModal := false

	dirty := true
	handleKey := func() {
		showModal = !showModal
		dirty = true
	}

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		root := &layout.Input{
			Kind:      layout.KindBox,
			Direction: "column",
			Children: func() []*layout.Input {
				var cs []*layout.Input
				cs = append(cs, &layout.Input{
					Kind:    layout.KindBox,
					Padding: layout.ParsePadding("1 2"),
					Border:  "single",
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: "Modal Demo",
							Style: render.Style{
								FG:   render.Color{Name: "green"},
								Bold: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: "Press any key to toggle modal, q to quit",
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: "Background content here",
						},
					},
				})
				if showModal {
					cs = append(cs, &layout.Input{
						Kind:        layout.KindBox,
						FixedWidth:  40,
						FixedHeight: 8,
						Padding:     layout.ParsePadding("1 2"),
						Border:      "single",
						Position:    "fixed",
						Top:         5,
						Left:        10,
						ZIndex:      2,
						Style: render.Style{
							FG: render.Color{Name: "yellow"},
							BG: render.Color{Name: "black"},
						},
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Modal Dialog",
								Style: render.Style{
									FG:   render.Color{Name: "yellow"},
									Bold: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "This is a fixed-position modal overlay.",
							},
							{
								Kind:    layout.KindText,
								Content: "Press any key to close",
								Style: render.Style{
									FG:  render.Color{Name: "cyan"},
									Dim: true,
								},
							},
						},
					})
				}
				return cs
			}(),
		}
		tree := layout.Layout(root, termW, termH)
		if prevTree == nil || termW != prevW || termH != prevH || layout.HasScrollChanged(prevTree, tree) || layout.HasOverlappingElements(tree) || layout.HasOverlappingElements(prevTree) {
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
				handleKey()
			}
		case <-resizeCh:
			dirty = true
		}
		if dirty {
			doRender()
		}
	}
}
