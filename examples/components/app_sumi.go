package main

import (
	"fmt"
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

func Run() {
	var app *tui.App
	counter0_count := 0
	counter1_count := 0

	counter0_increment := func() {
		counter0_count = counter0_count + 1
		app.Dirty = true
	}
	counter1_increment := func() {
		counter1_count = counter1_count + 1
		app.Dirty = true
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
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Padding:   layout.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
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
						Kind:      layout.KindBox,
						CursorCol: -1,
						CursorRow: -1,
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
						Kind:      layout.KindBox,
						CursorCol: -1,
						CursorRow: -1,
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
	sync := func() []*layout.Input {
		var changed []*layout.Input
		if v := fmt.Sprintf("%v", counter0_count); v != counter0_node0.Content {
			counter0_node0.Content = v
			changed = append(changed, counter0_node0)
		}
		if v := fmt.Sprintf("%v", counter1_count); v != counter1_node0.Content {
			counter1_node0.Content = v
			changed = append(changed, counter1_node0)
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
	}

	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'c' {
				app.Quit()
				return
			}
			if evt.Kind == input.EventSignal {
				app.Quit()
				return
			}
			if evt.Kind == input.EventKey {
				counter0_increment()
				counter1_increment()
			}
		},
	}
	app.Run()
}
