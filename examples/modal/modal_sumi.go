package main

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

func Run() {
	showModal := false

	var app *tui.App
	handleKey := func() {
		showModal = !showModal
		app.Dirty = true
	}

	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
	}
	sync := func() {
		root.Children = func() []*layout.Input {
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
			if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'c' {
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
