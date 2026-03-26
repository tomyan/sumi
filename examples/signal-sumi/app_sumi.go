package main

import (
	"fmt"
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

func Run() {
	count := signal.New(0)

	var app *tui.App
	handleKey := func(evt input.Event) {

		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			app.Quit()
			return
		}
		if evt.Rune == 'q' {
			app.Quit()
			return
		}
		if evt.Rune == '+' || evt.Rune == '=' {
			count.Update(func(n int) int { return n + 1 })
		}
		if evt.Rune == '-' {
			count.Update(func(n int) int { return n - 1 })
		}

	}
	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Count: %v", count.Get()),
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					node0,
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("Count: %v", count.Get())
		if app != nil {
			app.Dirty = true
			app.Wake()
		}
	})

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		tree := layout.Layout(root, termW, termH)
		if prevTree == nil || termW != prevW || termH != prevH {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
			changes, _ := layout.DiffTrees(prevTree, tree)
			layout.ApplyChanges(os.Stdout, changes)
		}
		prevTree = tree
		prevW = termW
		prevH = termH
	}
	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			handleKey(evt)
		},
	}
	app.Run()
}

func CreateApp(w, h int) *tui.App {
	count := signal.New(0)

	var app *tui.App
	handleKey := func(evt input.Event) {

		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			app.Quit()
			return
		}
		if evt.Rune == 'q' {
			app.Quit()
			return
		}
		if evt.Rune == '+' || evt.Rune == '=' {
			count.Update(func(n int) int { return n + 1 })
		}
		if evt.Rune == '-' {
			count.Update(func(n int) int { return n - 1 })
		}

	}
	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Count: %v", count.Get()),
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					node0,
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("Count: %v", count.Get())
		if app != nil {
			app.Dirty = true
			app.Wake()
		}
	})

	var prevTree *layout.Box
	var prevW, prevH int
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdout.Fd()))
		tree := layout.Layout(root, termW, termH)
		if prevTree == nil || termW != prevW || termH != prevH {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			render.ClearScreen(os.Stdout)
			buf.RenderTo(os.Stdout)
		} else {
			changes, _ := layout.DiffTrees(prevTree, tree)
			layout.ApplyChanges(os.Stdout, changes)
		}
		prevTree = tree
		prevW = termW
		prevH = termH
	}
	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			handleKey(evt)
		},
	}
	return app
}
