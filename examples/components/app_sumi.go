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
	dirty := true
	counter0_count := 0
	counter1_count := 0

	counter0_increment := func() {
		counter0_count = counter0_count + 1
		dirty = true
	}
	counter1_increment := func() {
		counter1_count = counter1_count + 1
		dirty = true
	}

	counter0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", counter0_count),
		Style: render.Style{
			FG:   render.Color{Name: "yellow"},
			Bold: true,
		},
	}
	counter1_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", counter1_count),
		Style: render.Style{
			FG:   render.Color{Name: "yellow"},
			Bold: true,
		},
	}
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
					{
						Kind: layout.KindBox,
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Clicks:",
								Style: render.Style{
									FG:   render.Color{Name: "cyan"},
									Bold: true,
								},
							},
							counter0_node0,
						},
					},
					{
						Kind: layout.KindBox,
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "Score:",
								Style: render.Style{
									FG:   render.Color{Name: "cyan"},
									Bold: true,
								},
							},
							counter1_node0,
						},
					},
				},
			},
		},
	}
	sync := func() {
		counter0_node0.Content = fmt.Sprintf("%v", counter0_count)
		counter1_node0.Content = fmt.Sprintf("%v", counter1_count)
	}

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		sync()
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
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
				counter0_increment()
				counter1_increment()
			}
		case <-resizeCh:
			dirty = true
		}
		if dirty {
			doRender()
		}
	}
}
