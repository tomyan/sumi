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
	width, height := term.GetSize(int(os.Stdin.Fd()))
	var app *tui.App
	var scroll0 layout.ScrollState
	draggingHScroll := false

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Terminal: %vx%v", width, height),
		Style: render.Style{
			FG:   render.Color{Name: "yellow"},
			Bold: true,
		},
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		Overflow:  "auto",
		MinWidth:  48,
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Padding:   layout.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
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
					node0,
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
	sync := func() {
		node0.Content = fmt.Sprintf("Terminal: %vx%v", width, height)
	}

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		sync()
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		tree := layout.Layout(root, termW, termH)
		tree.ScrollY = scroll0.ScrollY
		tree.ScrollX = scroll0.ScrollX
		changes, scrollChanged := layout.DiffTrees(prevTree, tree)
		if prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
			layout.ApplyChanges(os.Stdout, changes)
		}
		fmt.Fprintf(os.Stdout, "\033]2;Sumi %vx%v\007", width, height)
		prevTree = tree
		prevW = termW
		prevH = termH
	}

	app = &tui.App{
		HasMouse:  true,
		SaveTitle: true,
		OnRender:  doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventSpecial && prevTree != nil {
				switch evt.Special {
				case input.KeyDown:
					scroll0.ScrollDown(prevTree.ContentHeight, prevTree.Height)
					app.Dirty = true
				case input.KeyUp:
					scroll0.ScrollUp()
					app.Dirty = true
				case input.KeyPgDn:
					scroll0.PageDown(prevTree.ContentHeight, prevTree.Height)
					app.Dirty = true
				case input.KeyPgUp:
					scroll0.PageUp(prevTree.Height)
					app.Dirty = true
				case input.KeyRight:
					scroll0.ScrollRight(prevTree.ContentWidth, prevTree.Width)
					app.Dirty = true
				case input.KeyLeft:
					scroll0.ScrollLeft()
					app.Dirty = true
				}
			}
			if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseScroll && prevTree != nil {
				switch evt.Mouse.Button {
				case input.ScrollDown:
					scroll0.ScrollDown(prevTree.ContentHeight, prevTree.Height)
					app.Dirty = true
				case input.ScrollUp:
					scroll0.ScrollUp()
					app.Dirty = true
				}
			}
			if evt.Kind == input.EventMouse && prevTree != nil {
				if evt.Mouse.Action == input.MousePress {
					if prevTree.NeedsHorizontalScrollbar && prevTree.Clip != nil && evt.Mouse.Y == prevTree.Clip.Bottom {
						draggingHScroll = true
						scroll0.ScrollX = layout.ScrollXFromDrag(evt.Mouse.X-prevTree.Clip.Left, prevTree.ContentWidth, prevTree.Clip.Right-prevTree.Clip.Left+1)
						app.Dirty = true
					}
				}
				if evt.Mouse.Action == input.MouseMotion && draggingHScroll {
					scroll0.ScrollX = layout.ScrollXFromDrag(evt.Mouse.X-prevTree.Clip.Left, prevTree.ContentWidth, prevTree.Clip.Right-prevTree.Clip.Left+1)
					app.Dirty = true
				}
				if evt.Mouse.Action == input.MouseRelease {
					draggingHScroll = false
				}
			}
		},
		OnResize: func() {
			width, height = term.GetSize(int(os.Stdin.Fd()))
		},
	}
	app.Run()
}
