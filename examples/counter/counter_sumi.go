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
	count := 0

	dirty := true
	increment := func() {
		count = count + 1
		dirty = true
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Count: %v", count),
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
					node0,
				},
			},
		},
	}
	sync := func() []*layout.Input {
		var changed []*layout.Input
		if v := fmt.Sprintf("Count: %v", count); v != node0.Content {
			node0.Content = v
			changed = append(changed, node0)
		}
		return changed
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var nodeBoxMap map[*layout.Input]*layout.Box
	doRender := func() {
		changed := sync()
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		if prevTree != nil && len(changed) == 0 && termW == prevW && termH == prevH {
			dirty = false
			return
		}
		if prevTree != nil && len(changed) > 0 && termW == prevW && termH == prevH && !prevTree.HasOverlap && nodeBoxMap != nil {
			allDirect := true
			for _, inp := range changed {
				box := nodeBoxMap[inp]
				if !layout.DirectWriteText(os.Stdout, box, inp.Content, box.Content) {
					allDirect = false
					break
				}
				box.Content = inp.Content
			}
			if allDirect {
				dirty = false
				return
			}
		}
		tree := layout.Layout(root, termW, termH)
		nodeBoxMap = layout.MapInputToBox(root, tree)
		changes, scrollChanged := layout.DiffTrees(prevTree, tree)
		if prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
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
				increment()
			}
		case <-resizeCh:
			dirty = true
		}
		if dirty {
			doRender()
		}
	}
}
