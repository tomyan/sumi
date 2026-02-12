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
	items := []string{"Buy groceries", "Write tests", "Review PR"}
	selected := 0

	var app *tui.App
	handleKey := func() {
		selected = (selected + 1) % len(items)
		app.Dirty = true
	}

	box0 := &layout.Input{
		Kind:    layout.KindBox,
		Padding: layout.ParsePadding("1 2"),
		Border:  "single",
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		Children: []*layout.Input{
			box0,
		},
	}
	sync := func() {
		box0.Children = func() []*layout.Input {
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
		}()
	}

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		sync()
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		tree := layout.Layout(root, termW, termH)
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
			if evt.Kind == input.EventKey && evt.Rune == 3 {
				app.Quit()
				return
			}
			if evt.Kind == input.EventSignal {
				app.Quit()
				return
			}
			if evt.Kind == input.EventKey {
				handleKey()
			}
		},
	}
	app.Run()
}
