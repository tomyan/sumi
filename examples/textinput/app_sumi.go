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
	textinput0_viewOffset := 0
	textinput0_placeholder := "Enter your name..."
	textinput0_selfW := 0
	textinput0_contentW := textinput0_selfW - 4
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

	textinput0_adjustView := func() {
		if textinput0_contentW <= 0 {
			return
		}
		if textinput0_cursor < textinput0_viewOffset {
			textinput0_viewOffset = textinput0_cursor
			app.Dirty = true
		}
		if textinput0_cursor > textinput0_viewOffset+textinput0_contentW {
			textinput0_viewOffset = textinput0_cursor - textinput0_contentW
			app.Dirty = true
		}
		if textinput0_viewOffset < 0 {
			textinput0_viewOffset = 0
			app.Dirty = true
		}
	}
	textinput0_buildDisplayLine := func() string {
		cw := textinput0_contentW
		if cw <= 0 {
			return "[]"
		}
		text := name
		if len(text) == 0 && len(textinput0_placeholder) > 0 {
			text = textinput0_placeholder
		}
		vo := textinput0_viewOffset
		if len(name) == 0 {
			vo = 0
		}
		if vo < 0 {
			vo = 0
		}
		if vo > len(text) {
			vo = len(text)
		}
		end := vo + cw
		if end > len(text) {
			end = len(text)
		}
		visible := text[vo:end]
		left := " "
		if vo > 0 {
			left = "<"
		}
		right := " "
		if end < len(text) {
			right = ">"
		}
		pad := ""
		for i := len(visible); i < cw; i++ {
			pad = pad + " "
		}
		return "[" + left + visible + pad + right + "]"
	}
	textinput0_wordLeft := func() int {
		pos := textinput0_cursor
		for pos > 0 && name[pos-1] == ' ' {
			pos = pos - 1
		}
		for pos > 0 && name[pos-1] != ' ' {
			pos = pos - 1
		}
		return pos
	}
	textinput0_wordRight := func() int {
		pos := textinput0_cursor
		for pos < len(name) && name[pos] != ' ' {
			pos = pos + 1
		}
		for pos < len(name) && name[pos] == ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_handleEvent := func(evt input.Event) {
		if evt.Special == input.KeyBackspace && evt.Ctrl && textinput0_cursor > 0 {
			pos := textinput0_wordLeft()
			name = name[:pos] + name[textinput0_cursor:]
			app.Dirty = true
			textinput0_cursor = pos
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyBackspace && textinput0_cursor > 0 {
			name = name[:textinput0_cursor-1] + name[textinput0_cursor:]
			app.Dirty = true
			textinput0_cursor = textinput0_cursor - 1
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyDelete && evt.Ctrl && textinput0_cursor < len(name) {
			pos := textinput0_wordRight()
			name = name[:textinput0_cursor] + name[pos:]
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyDelete && textinput0_cursor < len(name) {
			name = name[:textinput0_cursor] + name[textinput0_cursor+1:]
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Ctrl && textinput0_cursor > 0 {
			textinput0_cursor = textinput0_wordLeft()
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && textinput0_cursor > 0 {
			textinput0_cursor = textinput0_cursor - 1
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Ctrl && textinput0_cursor < len(name) {
			textinput0_cursor = textinput0_wordRight()
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && textinput0_cursor < len(name) {
			textinput0_cursor = textinput0_cursor + 1
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyHome && textinput0_cursor > 0 {
			textinput0_cursor = 0
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyEnd && textinput0_cursor < len(name) {
			textinput0_cursor = len(name)
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'w' && textinput0_cursor > 0 {
			pos := textinput0_wordLeft()
			name = name[:pos] + name[textinput0_cursor:]
			app.Dirty = true
			textinput0_cursor = pos
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Rune >= 32 {
			name = name[:textinput0_cursor] + string(evt.Rune) + name[textinput0_cursor:]
			app.Dirty = true
			textinput0_cursor = textinput0_cursor + 1
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
	}

	textinput0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_buildDisplayLine()),
	}
	textinput0_box0 := &layout.Input{
		Kind:      layout.KindBox,
		Focusable: true,
		CursorCol: textinput0_cursor - textinput0_viewOffset + 2,
		CursorRow: 0,
		Children: []*layout.Input{
			textinput0_node0,
		},
	}
	textinput0_box0.SelfW = &textinput0_selfW
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
	root.SelfW = &textinput0_selfW
	sync := func() {
		textinput0_contentW = textinput0_selfW - 4
		textinput0_node0.Content = fmt.Sprintf("%v", textinput0_buildDisplayLine())
		node0.Content = fmt.Sprintf("You typed: %v", name)
		textinput0_box0.CursorCol = textinput0_cursor - textinput0_viewOffset + 2
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var prevTextinput0_selfW int
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
		if textinput0_selfW != prevTextinput0_selfW {
			prevTextinput0_selfW = textinput0_selfW
			app.Dirty = true
		}
		if cursorBox := layout.FindCursor(tree); cursorBox != nil {
			render.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)
		} else {
			render.HideCursor(os.Stdout)
		}
	}

	_ = textinput0_adjustView
	_ = textinput0_buildDisplayLine
	_ = textinput0_wordLeft
	_ = textinput0_wordRight
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
