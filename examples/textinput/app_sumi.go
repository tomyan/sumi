package main

import (
	"fmt"
	"os"
	"time"

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
	textinput0_scrollAnimating := false
	textinput0_scrollDragging := false
	textinput0_scrollHeld := false
	textinput0_scrollAnimStart := int64(0)
	textinput0_scrollAnimFrom := 0
	textinput0_scrollAnimTo := 0
	textinput0_placeholder := "Enter your name..."
	textinput0_selfW := 0
	textinput0_selfX := 0
	textinput0_selfY := 0
	textinput0_contentW := textinput0_selfW - 4
	focusIndex := -1
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
	textinput0_displayText := func() string {
		text := name
		if len(text) == 0 && len(textinput0_placeholder) > 0 {
			text = textinput0_placeholder
		}
		return text
	}
	textinput0_viewStart := func() int {
		vo := textinput0_viewOffset
		if len(name) == 0 {
			vo = 0
		}
		if vo < 0 {
			vo = 0
		}
		text := textinput0_displayText()
		if vo > len(text) {
			vo = len(text)
		}
		return vo
	}
	textinput0_viewEnd := func() int {
		vo := textinput0_viewStart()
		end := vo + textinput0_contentW
		text := textinput0_displayText()
		if end > len(text) {
			end = len(text)
		}
		return end
	}
	textinput0_leftIndicator := func() string {
		if textinput0_contentW <= 0 {
			return ""
		}
		if textinput0_viewStart() > 0 {
			return "<"
		}
		return " "
	}
	textinput0_visibleContent := func() string {
		if textinput0_contentW <= 0 {
			return ""
		}
		text := textinput0_displayText()
		visible := text[textinput0_viewStart():textinput0_viewEnd()]
		pad := ""
		for i := len(visible); i < textinput0_contentW; i++ {
			pad = pad + " "
		}
		return visible + pad
	}
	textinput0_rightIndicator := func() string {
		if textinput0_contentW <= 0 {
			return ""
		}
		if textinput0_viewEnd() < len(textinput0_displayText()) {
			return ">"
		}
		return " "
	}
	textinput0_scrollbarVisible := func() bool {
		return len(name) > textinput0_contentW && textinput0_contentW > 0
	}
	textinput0_scrollLeftArrow := func() string {
		if !textinput0_scrollbarVisible() {
			return ""
		}
		return "<"
	}
	textinput0_scrollTrackW := func() int {
		totalW := textinput0_contentW + 4
		return totalW - 2 - 4
	}
	textinput0_scrollTrack := func() string {
		if !textinput0_scrollbarVisible() {
			return ""
		}
		trackW := textinput0_scrollTrackW()
		if trackW < 1 {
			return ""
		}
		thumbSize := trackW * textinput0_contentW / len(name)
		if thumbSize < 1 {
			thumbSize = 1
		}
		maxOff := len(name) - textinput0_contentW
		thumbPos := 0
		if maxOff > 0 {
			thumbPos = (trackW - thumbSize) * textinput0_viewOffset / maxOff
		}
		if thumbPos < 0 {
			thumbPos = 0
		}
		if thumbPos+thumbSize > trackW {
			thumbPos = trackW - thumbSize
		}
		result := ""
		for i := 0; i < trackW; i++ {
			if i >= thumbPos && i < thumbPos+thumbSize {
				result = result + "#"
			} else {
				result = result + "-"
			}
		}
		return result
	}
	textinput0_scrollSpacer := func() string {
		if !textinput0_scrollbarVisible() {
			return ""
		}
		return "  "
	}
	textinput0_scrollRightArrow := func() string {
		if !textinput0_scrollbarVisible() {
			return ""
		}
		return ">"
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
	textinput0_clampCursorToView := func() {
		if textinput0_cursor < textinput0_viewOffset {
			textinput0_cursor = textinput0_viewOffset
			app.Dirty = true
		}
		if textinput0_cursor > textinput0_viewOffset+textinput0_contentW {
			textinput0_cursor = textinput0_viewOffset + textinput0_contentW
			app.Dirty = true
		}
		if textinput0_cursor > len(name) {
			textinput0_cursor = len(name)
			app.Dirty = true
		}
		if textinput0_cursor < 0 {
			textinput0_cursor = 0
			app.Dirty = true
		}
	}
	textinput0_maxOffset := func() int {
		return len(name) - textinput0_contentW
	}
	textinput0_scrollbarY := func() int {
		return textinput0_selfY + 1
	}
	textinput0_scrollTrackStart := func() int {
		return textinput0_selfX + 1 + 2
	}
	textinput0_viewOffsetForTrackX := func(trackX int) int {
		trackW := textinput0_scrollTrackW()
		if trackW <= 0 {
			return textinput0_viewOffset
		}
		maxOff := textinput0_maxOffset()
		if maxOff <= 0 {
			return 0
		}
		target := trackX * maxOff / trackW
		if target < 0 {
			target = 0
		}
		if target > maxOff {
			target = maxOff
		}
		return target
	}
	textinput0_cancelAnimation := func() {
		textinput0_scrollAnimating = false
		app.Dirty = true
		textinput0_scrollDragging = false
		app.Dirty = true
		textinput0_scrollHeld = false
		app.Dirty = true
	}
	textinput0_animateStep := func() {
		now := time.Now().UnixMilli()
		elapsed := now - textinput0_scrollAnimStart
		duration := int64(200)
		if elapsed >= duration {
			textinput0_viewOffset = textinput0_scrollAnimTo
			app.Dirty = true
			textinput0_scrollAnimating = false
			app.Dirty = true
			if textinput0_scrollHeld {
				textinput0_scrollDragging = true
				app.Dirty = true
			}
			textinput0_clampCursorToView()
			return
		}
		t := float64(elapsed) / float64(duration)
		inv := 1.0 - t
		eased := 1.0 - inv*inv*inv
		textinput0_viewOffset = textinput0_scrollAnimFrom + int(eased*float64(textinput0_scrollAnimTo-textinput0_scrollAnimFrom))
		app.Dirty = true
		textinput0_clampCursorToView()
		app.RequestFrame()
	}
	textinput0_startScrollAnimation := func(target int) {
		textinput0_scrollAnimating = true
		app.Dirty = true
		textinput0_scrollAnimStart = time.Now().UnixMilli()
		app.Dirty = true
		textinput0_scrollAnimFrom = textinput0_viewOffset
		app.Dirty = true
		textinput0_scrollAnimTo = target
		app.Dirty = true
		app.RequestFrame()
	}
	textinput0_handleEvent := func(evt input.Event) {
		if evt.Kind == input.EventFrame {
			if textinput0_scrollAnimating {
				textinput0_animateStep()
			}
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseRelease {
			if textinput0_scrollHeld || textinput0_scrollDragging {
				textinput0_scrollHeld = false
				app.Dirty = true
				textinput0_scrollDragging = false
				app.Dirty = true
				stopPropagation()
				return
			}
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_scrollDragging {
			trackX := evt.Mouse.X - textinput0_scrollTrackStart()
			textinput0_viewOffset = textinput0_viewOffsetForTrackX(trackX)
			app.Dirty = true
			textinput0_clampCursorToView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft && evt.Mouse.Y == textinput0_scrollbarY() && textinput0_scrollbarVisible() {
			trackX := evt.Mouse.X - textinput0_scrollTrackStart()
			trackW := textinput0_scrollTrackW()
			if trackX >= 0 && trackX < trackW {
				target := textinput0_viewOffsetForTrackX(trackX)
				textinput0_scrollHeld = true
				app.Dirty = true
				textinput0_startScrollAnimation(target)
				stopPropagation()
				return
			}
		}
		if textinput0_scrollAnimating {
			textinput0_cancelAnimation()
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			relX := evt.Mouse.X - textinput0_selfX - 2
			newCursor := textinput0_viewOffset + relX
			if newCursor < 0 {
				newCursor = 0
			}
			if newCursor > len(name) {
				newCursor = len(name)
			}
			textinput0_cursor = newCursor
			app.Dirty = true
			textinput0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseScroll {
			if evt.Mouse.Button == input.ScrollUp && textinput0_viewOffset > 0 {
				textinput0_viewOffset = textinput0_viewOffset - 3
				app.Dirty = true
				if textinput0_viewOffset < 0 {
					textinput0_viewOffset = 0
					app.Dirty = true
				}
				textinput0_clampCursorToView()
				stopPropagation()
				return
			}
			if evt.Mouse.Button == input.ScrollDown && textinput0_viewOffset < len(name)-textinput0_contentW {
				textinput0_viewOffset = textinput0_viewOffset + 3
				app.Dirty = true
				if textinput0_viewOffset > len(name)-textinput0_contentW {
					textinput0_viewOffset = len(name) - textinput0_contentW
					app.Dirty = true
				}
				textinput0_clampCursorToView()
				stopPropagation()
				return
			}
		}
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
		if evt.Kind == input.EventKey && !evt.Ctrl && evt.Rune >= 32 {
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
		Content: fmt.Sprintf("%v", textinput0_leftIndicator()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_visibleContent()),
	}
	textinput0_node2 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_rightIndicator()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_node3 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollLeftArrow()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_node4 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollSpacer()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_node5 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollTrack()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_node6 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollSpacer()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_node7 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollRightArrow()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_box0 := &layout.Input{
		Kind:      layout.KindBox,
		Focusable: true,
		CursorCol: textinput0_cursor - textinput0_viewOffset + 2,
		CursorRow: 0,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Direction: "row",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					{
						Kind:    layout.KindText,
						Content: "[",
					},
					textinput0_node0,
					textinput0_node1,
					textinput0_node2,
					{
						Kind:    layout.KindText,
						Content: "]",
					},
				},
			},
			{
				Kind:      layout.KindBox,
				Direction: "row",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					textinput0_node3,
					textinput0_node4,
					textinput0_node5,
					textinput0_node6,
					textinput0_node7,
				},
			},
		},
	}
	textinput0_box0.SelfW = &textinput0_selfW
	textinput0_box0.SelfX = &textinput0_selfX
	textinput0_box0.SelfY = &textinput0_selfY
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
		textinput0_contentW = textinput0_selfW - 4
		textinput0_node0.Content = fmt.Sprintf("%v", textinput0_leftIndicator())
		textinput0_node1.Content = fmt.Sprintf("%v", textinput0_visibleContent())
		textinput0_node2.Content = fmt.Sprintf("%v", textinput0_rightIndicator())
		textinput0_node3.Content = fmt.Sprintf("%v", textinput0_scrollLeftArrow())
		textinput0_node4.Content = fmt.Sprintf("%v", textinput0_scrollSpacer())
		textinput0_node5.Content = fmt.Sprintf("%v", textinput0_scrollTrack())
		textinput0_node6.Content = fmt.Sprintf("%v", textinput0_scrollSpacer())
		textinput0_node7.Content = fmt.Sprintf("%v", textinput0_scrollRightArrow())
		node0.Content = fmt.Sprintf("You typed: %v", name)
		if focusIndex == 0 {
			textinput0_box0.CursorCol = textinput0_cursor - textinput0_viewOffset + 2
		} else {
			textinput0_box0.CursorCol = -1
		}
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var prevTextinput0_selfW int
	var prevTextinput0_selfX int
	var prevTextinput0_selfY int
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
		if textinput0_selfX != prevTextinput0_selfX {
			prevTextinput0_selfX = textinput0_selfX
			app.Dirty = true
		}
		if textinput0_selfY != prevTextinput0_selfY {
			prevTextinput0_selfY = textinput0_selfY
			app.Dirty = true
		}
		if cursorBox := layout.FindCursor(tree); cursorBox != nil {
			render.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)
		} else {
			render.HideCursor(os.Stdout)
		}
	}

	_ = textinput0_adjustView
	_ = textinput0_displayText
	_ = textinput0_viewStart
	_ = textinput0_viewEnd
	_ = textinput0_leftIndicator
	_ = textinput0_visibleContent
	_ = textinput0_rightIndicator
	_ = textinput0_scrollbarVisible
	_ = textinput0_scrollLeftArrow
	_ = textinput0_scrollTrackW
	_ = textinput0_scrollTrack
	_ = textinput0_scrollSpacer
	_ = textinput0_scrollRightArrow
	_ = textinput0_wordLeft
	_ = textinput0_wordRight
	_ = textinput0_clampCursorToView
	_ = textinput0_maxOffset
	_ = textinput0_scrollbarY
	_ = textinput0_scrollTrackStart
	_ = textinput0_viewOffsetForTrackX
	_ = textinput0_cancelAnimation
	_ = textinput0_animateStep
	_ = textinput0_startScrollAnimation
	_ = stopPropagation
	app = &tui.App{
		HasMouse: true,
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventSpecial {
				if evt.Special == input.KeyTab {
					focusIndex = (focusIndex + 1) % focusCount
					app.Dirty = true
					return
				}
				if evt.Special == input.KeyShiftTab {
					if focusIndex < 0 {
						focusIndex = focusCount - 1
					} else {
						focusIndex = (focusIndex + focusCount - 1) % focusCount
					}
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
