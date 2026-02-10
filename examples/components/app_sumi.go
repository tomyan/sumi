package main

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

func Run() {
	counter0 := NewCounterComponent("Clicks")
	counter1 := NewCounterComponent("Score")

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
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: "Sumi Components Demo",
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
						counter0.Layout(),
						counter1.Layout(),
					},
				},
			},
		}
		tree := layout.Layout(root, termW, termH)
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
				counter0.HandleKey(evt.Rune)
				counter1.HandleKey(evt.Rune)
			}
		case <-resizeCh:
			dirty = true
		}
		if dirty || counter0.Dirty() || counter1.Dirty() {
			doRender()
		}
	}
}
