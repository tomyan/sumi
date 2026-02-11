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
	items := []string{"Buy groceries", "Write tests", "Review PR"}
	selected := 0

	dirty := true
	handleKey := func() {
		selected = (selected + 1) % len(items)
		dirty = true
	}

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
					Children: func() []*layout.Input {
						var cs []*layout.Input
						cs = append(cs, &layout.Input{
							Kind:    layout.KindText,
							Content: "Todo List",
							Style: render.Style{
								FG:   render.Color{Name: "green"},
								Bold: true,
							},
						})
						cs = append(cs, &layout.Input{
							Kind:    layout.KindText,
							Content: "Press any key to cycle, q to quit",
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						})
						for i, item := range items {
							if i == selected {
								cs = append(cs, &layout.Input{
									Kind:    layout.KindText,
									Content: fmt.Sprintf("> %v", item),
									Style: render.Style{
										FG:   render.Color{Name: "yellow"},
										Bold: true,
									},
								})
							} else {
								cs = append(cs, &layout.Input{
									Kind:    layout.KindText,
									Content: fmt.Sprintf("  %v", item),
								})
							}
							cs[len(cs)-1].Key = fmt.Sprint(item)
						}
						return cs
					}(),
				},
			},
		}
		tree := layout.Layout(root, termW, termH)
		if prevTree == nil || termW != prevW || termH != prevH || layout.HasScrollChanged(prevTree, tree) {
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
