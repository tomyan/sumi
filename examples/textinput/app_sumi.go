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
	name := ""

	var app *tui.App
	textinput0_cursor := 0
	focusIndex := 0
	focusCount := 1
	propagationStopped := false
	stopPropagation := func() { propagationStopped = true }

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			app.Quit()
			return
		}
	}

	textinput0_handleEvent := func(evt input.Event) {
		if evt.Special == input.KeyBackspace && textinput0_cursor > 0 {
			name = name[:textinput0_cursor-1] + name[textinput0_cursor:]
			app.Dirty = true
			textinput0_cursor = textinput0_cursor - 1
			app.Dirty = true
			stopPropagation()
		}
		if evt.Special == input.KeyLeft && textinput0_cursor > 0 {
			textinput0_cursor = textinput0_cursor - 1
			app.Dirty = true
			stopPropagation()
		}
		if evt.Special == input.KeyRight && textinput0_cursor < len(name) {
			textinput0_cursor = textinput0_cursor + 1
			app.Dirty = true
			stopPropagation()
		}
		if evt.Kind == input.EventKey {
			name = name[:textinput0_cursor] + string(evt.Rune) + name[textinput0_cursor:]
			app.Dirty = true
			textinput0_cursor = textinput0_cursor + 1
			app.Dirty = true
			stopPropagation()
		}
	}

	textinput0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", name),
	}
	textinput0_box0 := &layout.Input{
		Kind:      layout.KindBox,
		Focusable: true,
		CursorCol: textinput0_cursor,
		CursorRow: 0,
		Children: []*layout.Input{
			textinput0_node0,
		},
	}
	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("You typed: %v", name),
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
						Content: "Text Input Demo",
						Style: render.Style{
							FG:   render.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Type to enter your name, Tab to focus",
						Style: render.Style{
							FG:  render.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: "Name:",
						Style: render.Style{
							FG:   render.Color{Name: "yellow"},
							Bold: true,
						},
					},
					textinput0_box0,
					node0,
				},
			},
		},
	}
	sync := func() {
		textinput0_node0.Content = fmt.Sprintf("%v", name)
		node0.Content = fmt.Sprintf("You typed: %v", name)
		textinput0_box0.CursorCol = textinput0_cursor
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
		if cursorBox := layout.FindCursor(tree); cursorBox != nil {
			render.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)
		} else {
			render.HideCursor(os.Stdout)
		}
	}

	_ = stopPropagation
	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventSpecial {
				if evt.Special == input.KeyTab {
					focusIndex = (focusIndex + 1) % focusCount
					app.Dirty = true
					return
				}
				if evt.Special == input.KeyShiftTab {
					focusIndex = (focusIndex + focusCount - 1) % focusCount
					app.Dirty = true
					return
				}
			}
			propagationStopped = false
			switch focusIndex {
			case 0:
				textinput0_handleEvent(evt)
			}
			if !propagationStopped {
				handleKey(evt)
			}
		},
	}
	app.Run()
}
