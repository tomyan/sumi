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

func main() {
	// Reactive state.
	count := signal.New(0)

	// Derived value.
	label := signal.From(func() string {
		return fmt.Sprintf("Count: %d", count.Get())
	})

	// Build-once layout tree with a mutable text node.
	textNode := &layout.Input{
		Kind:      layout.KindText,
		Content:   label.Get(),
		CursorCol: -1,
		CursorRow: -1,
	}
	root := &layout.Input{
		Kind:        layout.KindBox,
		Border:      "single",
		BorderTitle: "Signal Counter",
		Padding:     layout.Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		CursorCol:   -1,
		CursorRow:   -1,
		Style:       render.Style{FG: render.Color{Name: "cyan"}},
		Children:    []*layout.Input{textNode},
	}

	// Render state.
	var prevTree *layout.Box
	var prevW, prevH int

	app := &tui.App{}

	// Sync: when label changes, update the tree node and trigger re-render.
	signal.Effect(func() {
		textNode.Content = label.Get()
		app.Dirty = true
		app.Wake()
	})

	app.OnRender = func() {
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

	app.OnEvent = func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if evt.Kind != input.EventKey {
			return
		}
		switch {
		case evt.Ctrl && evt.Rune == 'c':
			app.Quit()
		case evt.Rune == 'q':
			app.Quit()
		case evt.Rune == '+' || evt.Rune == '=' || evt.Rune == 'l':
			count.Update(func(n int) int { return n + 1 })
		case evt.Rune == '-' || evt.Rune == 'h':
			count.Update(func(n int) int { return n - 1 })
		}
	}

	app.Run()
}
