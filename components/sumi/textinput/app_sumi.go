package textinput

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
	value := ""

	var app *tui.App
	textinput0_viewOffset := 0
	textinput0_focused := false
	textinput0_placeholder := "Type here..."
	textinput0_inputType := ""
	textinput0_maxlength := 0
	textinput0_readonly := false
	textinput0_strip := false
	textinput0_selfW := 0
	textinput0_contentW := textinput0_selfW - 4
	textinput0_textedit0_cursor := 0
	textinput0_textedit0_killBuffer := ""
	textinput0_textedit0_undoValues := []string{}
	textinput0_textedit0_undoCursors := []int{}
	textinput0_textedit0_selAnchor := -1
	textinput0_textedit0_lastClickTime := int64(0)
	textinput0_textedit0_lastClickX := -1
	textinput0_textedit0_mouseDragging := false
	textinput0_textedit0_wordDragging := false
	textinput0_textedit0_wordDragStart := 0
	textinput0_textedit0_wordDragEnd := 0
	textinput0_textedit0_selfW := 0
	textinput0_textedit0_selfX := 0
	textinput0_textedit0_selfY := 0
	textinput0_textedit0_contentW := textinput0_textedit0_selfW
	textinput0_scrollbar0_animating := false
	textinput0_scrollbar0_dragging := false
	textinput0_scrollbar0_held := false
	textinput0_scrollbar0_animStart := int64(0)
	textinput0_scrollbar0_animFrom := 0
	textinput0_scrollbar0_animTo := 0
	textinput0_scrollbar0_direction := "horizontal"
	textinput0_scrollbar0_selfX := 0
	textinput0_scrollbar0_selfY := 0
	textinput0_scrollbar0_selfW := 0
	textinput0_scrollbar0_selfH := 0
	focusIndex := -1
	focusCount := 2
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

	textinput0_leftIndicator := func() string {
		if textinput0_viewOffset > 0 {
			return "<"
		}
		return " "
	}
	textinput0_rightIndicator := func() string {
		if textinput0_contentW > 0 && textinput0_viewOffset+textinput0_contentW < len(value) {
			return ">"
		}
		return " "
	}
	textinput0_scrollbarVisible := func() bool {
		return textinput0_focused && len(value) > textinput0_contentW && textinput0_contentW > 0
	}
	textinput0_textedit0_adjustView := func() {
		if textinput0_textedit0_contentW <= 0 {
			return
		}
		if textinput0_textedit0_cursor < textinput0_viewOffset {
			textinput0_viewOffset = textinput0_textedit0_cursor
			app.Dirty = true
		}
		if textinput0_textedit0_cursor > textinput0_viewOffset+textinput0_textedit0_contentW {
			textinput0_viewOffset = textinput0_textedit0_cursor - textinput0_textedit0_contentW
			app.Dirty = true
		}
		if textinput0_viewOffset < 0 {
			textinput0_viewOffset = 0
			app.Dirty = true
		}
	}
	textinput0_textedit0_clearSelection := func() {
		textinput0_textedit0_selAnchor = -1
		app.Dirty = true
	}
	textinput0_textedit0_selStart := func() int {
		if textinput0_textedit0_selAnchor < textinput0_textedit0_cursor {
			return textinput0_textedit0_selAnchor
		}
		return textinput0_textedit0_cursor
	}
	textinput0_textedit0_selEnd := func() int {
		if textinput0_textedit0_selAnchor > textinput0_textedit0_cursor {
			return textinput0_textedit0_selAnchor
		}
		return textinput0_textedit0_cursor
	}
	textinput0_textedit0_startSelection := func() {
		if textinput0_textedit0_selAnchor == -1 {
			textinput0_textedit0_selAnchor = textinput0_textedit0_cursor
			app.Dirty = true
		}
	}
	textinput0_textedit0_maskedValue := func() string {
		result := ""
		for i := 0; i < len(value); i++ {
			result = result + "*"
		}
		return result
	}
	textinput0_textedit0_displayValue := func() string {
		if textinput0_inputType == "password" {
			return textinput0_textedit0_maskedValue()
		}
		return value
	}
	textinput0_textedit0_displayText := func() string {
		text := textinput0_textedit0_displayValue()
		if len(text) == 0 && len(textinput0_placeholder) > 0 {
			text = textinput0_placeholder
		}
		return text
	}
	textinput0_textedit0_viewStart := func() int {
		vo := textinput0_viewOffset
		if len(value) == 0 {
			vo = 0
		}
		if vo < 0 {
			vo = 0
		}
		text := textinput0_textedit0_displayText()
		if vo > len(text) {
			vo = len(text)
		}
		return vo
	}
	textinput0_textedit0_viewEnd := func() int {
		vo := textinput0_textedit0_viewStart()
		end := vo + textinput0_textedit0_contentW
		text := textinput0_textedit0_displayText()
		if end > len(text) {
			end = len(text)
		}
		return end
	}
	textinput0_textedit0_visiblePreSel := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) == 0 {
			return ""
		}
		dv := textinput0_textedit0_displayValue()
		vs := textinput0_textedit0_viewStart()
		ve := textinput0_textedit0_viewEnd()
		if textinput0_textedit0_selAnchor == -1 {
			visible := dv[vs:ve]
			pad := ""
			for i := len(visible); i < textinput0_textedit0_contentW; i++ {
				pad = pad + " "
			}
			return visible + pad
		}
		ss := textinput0_textedit0_selStart()
		end := ss
		if end > ve {
			end = ve
		}
		if end <= vs {
			return ""
		}
		return dv[vs:end]
	}
	textinput0_textedit0_visibleSel := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) == 0 || textinput0_textedit0_selAnchor == -1 {
			return ""
		}
		dv := textinput0_textedit0_displayValue()
		vs := textinput0_textedit0_viewStart()
		ve := textinput0_textedit0_viewEnd()
		ss := textinput0_textedit0_selStart()
		se := textinput0_textedit0_selEnd()
		start := ss
		if start < vs {
			start = vs
		}
		end := se
		if end > ve {
			end = ve
		}
		if start >= end {
			return ""
		}
		return dv[start:end]
	}
	textinput0_textedit0_visiblePostSel := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) == 0 || textinput0_textedit0_selAnchor == -1 {
			return ""
		}
		dv := textinput0_textedit0_displayValue()
		vs := textinput0_textedit0_viewStart()
		ve := textinput0_textedit0_viewEnd()
		se := textinput0_textedit0_selEnd()
		start := se
		if start < vs {
			start = vs
		}
		if start >= ve {
			totalVisible := ve - vs
			pad := ""
			for i := totalVisible; i < textinput0_textedit0_contentW; i++ {
				pad = pad + " "
			}
			return pad
		}
		text := dv[start:ve]
		totalVisible := ve - vs
		pad := ""
		for i := totalVisible; i < textinput0_textedit0_contentW; i++ {
			pad = pad + " "
		}
		return text + pad
	}
	textinput0_textedit0_visiblePlaceholder := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) > 0 || len(textinput0_placeholder) == 0 {
			return ""
		}
		text := textinput0_placeholder
		if len(text) > textinput0_textedit0_contentW {
			text = text[:textinput0_textedit0_contentW]
		}
		pad := ""
		for i := len(text); i < textinput0_textedit0_contentW; i++ {
			pad = pad + " "
		}
		return text + pad
	}
	textinput0_textedit0_cursorX := func() int {
		if textinput0_textedit0_selAnchor >= 0 {
			return -1
		}
		return textinput0_textedit0_cursor - textinput0_viewOffset
	}
	textinput0_textedit0_wordLeft := func() int {
		pos := textinput0_textedit0_cursor
		for pos > 0 && value[pos-1] == ' ' {
			pos = pos - 1
		}
		for pos > 0 && value[pos-1] != ' ' {
			pos = pos - 1
		}
		return pos
	}
	textinput0_textedit0_wordRight := func() int {
		pos := textinput0_textedit0_cursor
		for pos < len(value) && value[pos] != ' ' {
			pos = pos + 1
		}
		for pos < len(value) && value[pos] == ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_textedit0_wordLeftFrom := func(pos int) int {
		for pos > 0 && value[pos-1] == ' ' {
			pos = pos - 1
		}
		for pos > 0 && value[pos-1] != ' ' {
			pos = pos - 1
		}
		return pos
	}
	textinput0_textedit0_wordRightFrom := func(pos int) int {
		for pos < len(value) && value[pos] != ' ' {
			pos = pos + 1
		}
		for pos < len(value) && value[pos] == ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_textedit0_wordEndFrom := func(pos int) int {
		for pos < len(value) && value[pos] != ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_textedit0_mousePosToTextPos := func(mouseX int) int {
		relX := mouseX - textinput0_textedit0_selfX
		pos := textinput0_viewOffset + relX
		if pos < 0 {
			pos = 0
		}
		if pos > len(value) {
			pos = len(value)
		}
		return pos
	}
	textinput0_textedit0_saveUndo := func() {
		textinput0_textedit0_undoValues = append(textinput0_textedit0_undoValues, value)
		app.Dirty = true
		textinput0_textedit0_undoCursors = append(textinput0_textedit0_undoCursors, textinput0_textedit0_cursor)
		app.Dirty = true
		if len(textinput0_textedit0_undoValues) > 100 {
			textinput0_textedit0_undoValues = textinput0_textedit0_undoValues[1:]
			app.Dirty = true
			textinput0_textedit0_undoCursors = textinput0_textedit0_undoCursors[1:]
			app.Dirty = true
		}
	}
	textinput0_textedit0_undo := func() {
		if len(textinput0_textedit0_undoValues) == 0 {
			return
		}
		last := len(textinput0_textedit0_undoValues) - 1
		value = textinput0_textedit0_undoValues[last]
		app.Dirty = true
		textinput0_textedit0_cursor = textinput0_textedit0_undoCursors[last]
		app.Dirty = true
		textinput0_textedit0_undoValues = textinput0_textedit0_undoValues[:last]
		app.Dirty = true
		textinput0_textedit0_undoCursors = textinput0_textedit0_undoCursors[:last]
		app.Dirty = true
		textinput0_textedit0_adjustView()
	}
	textinput0_textedit0_deleteSelection := func() {
		textinput0_textedit0_saveUndo()
		value = value[:textinput0_textedit0_selStart()] + value[textinput0_textedit0_selEnd():]
		app.Dirty = true
		textinput0_textedit0_cursor = textinput0_textedit0_selStart()
		app.Dirty = true
		textinput0_textedit0_selAnchor = -1
		app.Dirty = true
		textinput0_textedit0_adjustView()
	}
	textinput0_textedit0_killText := func(text string) {
		textinput0_textedit0_killBuffer = text
		app.Dirty = true
	}
	textinput0_textedit0_transposeChars := func() {
		if textinput0_textedit0_cursor < 2 {
			return
		}
		a := value[textinput0_textedit0_cursor-2]
		b := value[textinput0_textedit0_cursor-1]
		value = value[:textinput0_textedit0_cursor-2] + string(b) + string(a) + value[textinput0_textedit0_cursor:]
		app.Dirty = true
	}
	textinput0_textedit0_handleEvent := func(evt input.Event) {
		if evt.Kind == input.EventFocus {
			textinput0_focused = true
			app.Dirty = true
			stopPropagation()
			return
		}
		if evt.Kind == input.EventBlur {
			textinput0_focused = false
			app.Dirty = true
			textinput0_textedit0_clearSelection()
			if textinput0_strip && !textinput0_readonly {
				for len(value) > 0 && value[0] == ' ' {
					value = value[1:]
					app.Dirty = true
				}
				for len(value) > 0 && value[len(value)-1] == ' ' {
					value = value[:len(value)-1]
					app.Dirty = true
				}
				if textinput0_textedit0_cursor > len(value) {
					textinput0_textedit0_cursor = len(value)
					app.Dirty = true
				}
				textinput0_textedit0_adjustView()
			}
			stopPropagation()
			return
		}
		if evt.Kind == input.EventPaste {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			paste := evt.PasteText
			if textinput0_textedit0_selAnchor >= 0 {
				textinput0_textedit0_deleteSelection()
			}
			if textinput0_maxlength > 0 {
				room := textinput0_maxlength - len(value)
				if room <= 0 {
					stopPropagation()
					return
				}
				if len(paste) > room {
					paste = paste[:room]
				}
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + paste + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + len(paste)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseRelease {
			if textinput0_textedit0_mouseDragging || textinput0_textedit0_wordDragging {
				textinput0_textedit0_mouseDragging = false
				app.Dirty = true
				textinput0_textedit0_wordDragging = false
				app.Dirty = true
				if textinput0_textedit0_selAnchor == textinput0_textedit0_cursor {
					textinput0_textedit0_selAnchor = -1
					app.Dirty = true
				}
				stopPropagation()
				return
			}
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_textedit0_wordDragging {
			pos := textinput0_textedit0_mousePosToTextPos(evt.Mouse.X)
			wl := textinput0_textedit0_wordLeftFrom(pos)
			wr := textinput0_textedit0_wordEndFrom(pos)
			if wl < textinput0_textedit0_wordDragStart {
				textinput0_textedit0_selAnchor = textinput0_textedit0_wordDragEnd
				app.Dirty = true
				textinput0_textedit0_cursor = wl
				app.Dirty = true
			} else {
				textinput0_textedit0_selAnchor = textinput0_textedit0_wordDragStart
				app.Dirty = true
				textinput0_textedit0_cursor = wr
				app.Dirty = true
			}
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_textedit0_mouseDragging {
			pos := textinput0_textedit0_mousePosToTextPos(evt.Mouse.X)
			textinput0_textedit0_cursor = pos
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			clickPos := textinput0_textedit0_mousePosToTextPos(evt.Mouse.X)
			now := time.Now().UnixMilli()
			if evt.Shift {
				if textinput0_textedit0_selAnchor == -1 {
					textinput0_textedit0_selAnchor = textinput0_textedit0_cursor
					app.Dirty = true
				}
				textinput0_textedit0_cursor = clickPos
				app.Dirty = true
				textinput0_textedit0_mouseDragging = true
				app.Dirty = true
				textinput0_textedit0_adjustView()
				stopPropagation()
				return
			}
			if now-textinput0_textedit0_lastClickTime < 500 && evt.Mouse.X == textinput0_textedit0_lastClickX {
				wl := textinput0_textedit0_wordLeftFrom(clickPos)
				wr := textinput0_textedit0_wordEndFrom(clickPos)
				textinput0_textedit0_selAnchor = wl
				app.Dirty = true
				textinput0_textedit0_cursor = wr
				app.Dirty = true
				textinput0_textedit0_wordDragging = true
				app.Dirty = true
				textinput0_textedit0_wordDragStart = wl
				app.Dirty = true
				textinput0_textedit0_wordDragEnd = wr
				app.Dirty = true
				textinput0_textedit0_adjustView()
				textinput0_textedit0_lastClickTime = 0
				app.Dirty = true
				stopPropagation()
				return
			}
			textinput0_textedit0_lastClickTime = now
			app.Dirty = true
			textinput0_textedit0_lastClickX = evt.Mouse.X
			app.Dirty = true
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = clickPos
			app.Dirty = true
			textinput0_textedit0_selAnchor = clickPos
			app.Dirty = true
			textinput0_textedit0_mouseDragging = true
			app.Dirty = true
			textinput0_textedit0_adjustView()
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
				stopPropagation()
				return
			}
			if evt.Mouse.Button == input.ScrollDown && textinput0_viewOffset < len(value)-textinput0_textedit0_contentW {
				textinput0_viewOffset = textinput0_viewOffset + 3
				app.Dirty = true
				if textinput0_viewOffset > len(value)-textinput0_textedit0_contentW {
					textinput0_viewOffset = len(value) - textinput0_textedit0_contentW
					app.Dirty = true
				}
				stopPropagation()
				return
			}
		}
		if (evt.Special == input.KeyBackspace || evt.Special == input.KeyDelete) && textinput0_textedit0_selAnchor >= 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_deleteSelection()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyBackspace && evt.Ctrl && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordLeft()
			textinput0_textedit0_killText(value[pos:textinput0_textedit0_cursor])
			value = value[:pos] + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = pos
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyBackspace && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor-1] + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyDelete && evt.Ctrl && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordRight()
			textinput0_textedit0_killText(value[textinput0_textedit0_cursor:pos])
			value = value[:textinput0_textedit0_cursor] + value[pos:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyDelete && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + value[textinput0_textedit0_cursor+1:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Shift && evt.Ctrl {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordLeft()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Shift && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Shift && evt.Ctrl {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordRight()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Shift && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyHome && evt.Shift {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = 0
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyEnd && evt.Shift {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Ctrl && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordLeft()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Ctrl && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordRight()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyHome {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = 0
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyEnd {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'a' {
			textinput0_textedit0_selAnchor = 0
			app.Dirty = true
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'e' {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'f' && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'b' && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'c' && textinput0_textedit0_selAnchor >= 0 {
			selected := value[textinput0_textedit0_selStart():textinput0_textedit0_selEnd()]
			textinput0_textedit0_killText(selected)
			render.CopyToClipboard(os.Stdout, selected)
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'x' && textinput0_textedit0_selAnchor >= 0 {
			if textinput0_readonly {
				selected := value[textinput0_textedit0_selStart():textinput0_textedit0_selEnd()]
				textinput0_textedit0_killText(selected)
				render.CopyToClipboard(os.Stdout, selected)
				stopPropagation()
				return
			}
			selected := value[textinput0_textedit0_selStart():textinput0_textedit0_selEnd()]
			textinput0_textedit0_killText(selected)
			render.CopyToClipboard(os.Stdout, selected)
			textinput0_textedit0_deleteSelection()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'd' && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + value[textinput0_textedit0_cursor+1:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'k' && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			textinput0_textedit0_killText(value[textinput0_textedit0_cursor:])
			value = value[:textinput0_textedit0_cursor]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'u' && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			textinput0_textedit0_killText(value[:textinput0_textedit0_cursor])
			value = value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = 0
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'w' && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordLeft()
			textinput0_textedit0_killText(value[pos:textinput0_textedit0_cursor])
			value = value[:pos] + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = pos
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 't' && textinput0_textedit0_cursor >= 2 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			textinput0_textedit0_transposeChars()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'y' && len(textinput0_textedit0_killBuffer) > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			yank := textinput0_textedit0_killBuffer
			if textinput0_maxlength > 0 {
				room := textinput0_maxlength - len(value)
				if room <= 0 {
					stopPropagation()
					return
				}
				if len(yank) > room {
					yank = yank[:room]
				}
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + yank + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + len(yank)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == '/' {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_undo()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Alt && evt.Rune == 'f' && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordRight()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Alt && evt.Rune == 'b' && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordLeft()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Alt && evt.Rune == 'd' && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordRight()
			textinput0_textedit0_killText(value[textinput0_textedit0_cursor:pos])
			value = value[:textinput0_textedit0_cursor] + value[pos:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && !evt.Ctrl && !evt.Alt && evt.Rune >= 32 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			if textinput0_textedit0_selAnchor >= 0 {
				textinput0_textedit0_deleteSelection()
			}
			if textinput0_maxlength > 0 && len(value) >= textinput0_maxlength {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + string(evt.Rune) + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
	}
	textinput0_scrollbar0_isVisible := func() bool {
		return textinput0_scrollbarVisible() && len(value) > textinput0_contentW && textinput0_contentW > 0
	}
	textinput0_scrollbar0_trackChar := func() string {
		if textinput0_scrollbar0_direction == "vertical" {
			return "|"
		}
		return "-"
	}
	textinput0_scrollbar0_thumbChar := func() string {
		return "#"
	}
	textinput0_scrollbar0_startArrow := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		if textinput0_scrollbar0_direction == "vertical" {
			return "^"
		}
		return "<"
	}
	textinput0_scrollbar0_endArrow := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		if textinput0_scrollbar0_direction == "vertical" {
			return "v"
		}
		return ">"
	}
	textinput0_scrollbar0_trackSize := func() int {
		if textinput0_scrollbar0_direction == "vertical" {
			return textinput0_scrollbar0_selfH - 2
		}
		return textinput0_scrollbar0_selfW - 2
	}
	textinput0_scrollbar0_thumbSize := func() int {
		ts := textinput0_scrollbar0_trackSize()
		if ts < 1 {
			return 1
		}
		size := ts * textinput0_contentW / len(value)
		if size < 1 {
			size = 1
		}
		return size
	}
	textinput0_scrollbar0_thumbPos := func() int {
		ts := textinput0_scrollbar0_trackSize()
		ths := textinput0_scrollbar0_thumbSize()
		maxOff := len(value) - textinput0_contentW
		if maxOff <= 0 {
			return 0
		}
		pos := (ts - ths) * textinput0_viewOffset / maxOff
		if pos < 0 {
			pos = 0
		}
		if pos+ths > ts {
			pos = ts - ths
		}
		return pos
	}
	textinput0_scrollbar0_trackText := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		ts := textinput0_scrollbar0_trackSize()
		if ts < 1 {
			return ""
		}
		tp := textinput0_scrollbar0_thumbPos()
		ths := textinput0_scrollbar0_thumbSize()
		tc := textinput0_scrollbar0_trackChar()
		thc := textinput0_scrollbar0_thumbChar()
		result := ""
		for i := 0; i < ts; i++ {
			if i >= tp && i < tp+ths {
				result = result + thc
			} else {
				result = result + tc
			}
		}
		return result
	}
	textinput0_scrollbar0_spacer := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		return " "
	}
	textinput0_scrollbar0_maxOffset := func() int {
		return len(value) - textinput0_contentW
	}
	textinput0_scrollbar0_offsetForTrackPos := func(pos int) int {
		ts := textinput0_scrollbar0_trackSize()
		if ts <= 0 {
			return textinput0_viewOffset
		}
		maxOff := textinput0_scrollbar0_maxOffset()
		if maxOff <= 0 {
			return 0
		}
		target := pos * maxOff / ts
		if target < 0 {
			target = 0
		}
		if target > maxOff {
			target = maxOff
		}
		return target
	}
	textinput0_scrollbar0_trackStart := func() int {
		if textinput0_scrollbar0_direction == "vertical" {
			return textinput0_scrollbar0_selfY + 1
		}
		return textinput0_scrollbar0_selfX + 1 + 1
	}
	textinput0_scrollbar0_mouseToTrackPos := func(mousePos int) int {
		return mousePos - textinput0_scrollbar0_trackStart()
	}
	textinput0_scrollbar0_isOnTrack := func(evt input.Event) bool {
		if textinput0_scrollbar0_direction == "vertical" {
			return evt.Mouse.X == textinput0_scrollbar0_selfX
		}
		return evt.Mouse.Y == textinput0_scrollbar0_selfY
	}
	textinput0_scrollbar0_mouseTrackPos := func(evt input.Event) int {
		if textinput0_scrollbar0_direction == "vertical" {
			return textinput0_scrollbar0_mouseToTrackPos(evt.Mouse.Y)
		}
		return textinput0_scrollbar0_mouseToTrackPos(evt.Mouse.X)
	}
	textinput0_scrollbar0_cancelAnimation := func() {
		textinput0_scrollbar0_animating = false
		app.Dirty = true
		textinput0_scrollbar0_dragging = false
		app.Dirty = true
		textinput0_scrollbar0_held = false
		app.Dirty = true
	}
	textinput0_scrollbar0_animateStep := func() {
		now := time.Now().UnixMilli()
		elapsed := now - textinput0_scrollbar0_animStart
		duration := int64(200)
		if elapsed >= duration {
			textinput0_viewOffset = textinput0_scrollbar0_animTo
			app.Dirty = true
			textinput0_scrollbar0_animating = false
			app.Dirty = true
			if textinput0_scrollbar0_held {
				textinput0_scrollbar0_dragging = true
				app.Dirty = true
			}
			return
		}
		t := float64(elapsed) / float64(duration)
		inv := 1.0 - t
		eased := 1.0 - inv*inv*inv
		textinput0_viewOffset = textinput0_scrollbar0_animFrom + int(eased*float64(textinput0_scrollbar0_animTo-textinput0_scrollbar0_animFrom))
		app.Dirty = true
		app.RequestFrame()
	}
	textinput0_scrollbar0_startAnimation := func(target int) {
		textinput0_scrollbar0_animating = true
		app.Dirty = true
		textinput0_scrollbar0_animStart = time.Now().UnixMilli()
		app.Dirty = true
		textinput0_scrollbar0_animFrom = textinput0_viewOffset
		app.Dirty = true
		textinput0_scrollbar0_animTo = target
		app.Dirty = true
		app.RequestFrame()
	}
	textinput0_scrollbar0_handleEvent := func(evt input.Event) {
		if evt.Kind == input.EventFrame {
			if textinput0_scrollbar0_animating {
				textinput0_scrollbar0_animateStep()
			}
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseRelease {
			if textinput0_scrollbar0_held || textinput0_scrollbar0_dragging {
				textinput0_scrollbar0_held = false
				app.Dirty = true
				textinput0_scrollbar0_dragging = false
				app.Dirty = true
				stopPropagation()
				return
			}
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_scrollbar0_dragging {
			pos := textinput0_scrollbar0_mouseTrackPos(evt)
			textinput0_viewOffset = textinput0_scrollbar0_offsetForTrackPos(pos)
			app.Dirty = true
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft && textinput0_scrollbar0_isOnTrack(evt) {
			pos := textinput0_scrollbar0_mouseTrackPos(evt)
			ts := textinput0_scrollbar0_trackSize()
			if pos >= 0 && pos < ts {
				target := textinput0_scrollbar0_offsetForTrackPos(pos)
				textinput0_scrollbar0_held = true
				app.Dirty = true
				textinput0_scrollbar0_startAnimation(target)
				stopPropagation()
				return
			}
		}
	}

	dispatchToFocusable := func(idx int, evt input.Event) {
		switch idx {
		case 0:
			textinput0_textedit0_handleEvent(evt)
		case 1:
			textinput0_scrollbar0_handleEvent(evt)
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Value: %v", value),
	}
	textinput0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_leftIndicator()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_textedit0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visiblePreSel()),
	}
	textinput0_textedit0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visibleSel()),
		Style: render.Style{
			Inverse: true,
		},
	}
	textinput0_textedit0_node2 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visiblePostSel()),
	}
	textinput0_textedit0_node3 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visiblePlaceholder()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_textedit0_box0 := &layout.Input{
		Kind:      layout.KindBox,
		Focusable: true,
		CursorCol: textinput0_textedit0_cursorX(),
		CursorRow: 0,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Direction: "row",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					textinput0_textedit0_node0,
					textinput0_textedit0_node1,
					textinput0_textedit0_node2,
					textinput0_textedit0_node3,
				},
			},
		},
	}
	textinput0_textedit0_box0.SelfW = &textinput0_textedit0_selfW
	textinput0_textedit0_box0.SelfX = &textinput0_textedit0_selfX
	textinput0_textedit0_box0.SelfY = &textinput0_textedit0_selfY
	textinput0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_rightIndicator()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_spacer()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_startArrow()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node2 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_trackText()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node3 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_endArrow()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node4 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_spacer()),
		Style: render.Style{
			Dim: true,
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
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					node0,
					{
						Kind:       layout.KindBox,
						FixedWidth: 24,
						CursorCol:  -1,
						CursorRow:  -1,
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
									textinput0_textedit0_box0,
									textinput0_node1,
									{
										Kind:    layout.KindText,
										Content: "]",
									},
								},
							},
							{
								Kind:      layout.KindBox,
								Direction: "row",
								Focusable: true,
								CursorCol: -1,
								CursorRow: -1,
								Children: []*layout.Input{
									textinput0_scrollbar0_node0,
									textinput0_scrollbar0_node1,
									textinput0_scrollbar0_node2,
									textinput0_scrollbar0_node3,
									textinput0_scrollbar0_node4,
								},
							},
						},
					},
				},
			},
		},
	}
	sync := func() {
		textinput0_contentW = textinput0_selfW - 4
		textinput0_textedit0_contentW = textinput0_textedit0_selfW
		textinput0_focused = focusIndex == 0
		node0.Content = fmt.Sprintf("Value: %v", value)
		textinput0_node0.Content = fmt.Sprintf("%v", textinput0_leftIndicator())
		textinput0_textedit0_node0.Content = fmt.Sprintf("%v", textinput0_textedit0_visiblePreSel())
		textinput0_textedit0_node1.Content = fmt.Sprintf("%v", textinput0_textedit0_visibleSel())
		textinput0_textedit0_node2.Content = fmt.Sprintf("%v", textinput0_textedit0_visiblePostSel())
		textinput0_textedit0_node3.Content = fmt.Sprintf("%v", textinput0_textedit0_visiblePlaceholder())
		textinput0_node1.Content = fmt.Sprintf("%v", textinput0_rightIndicator())
		textinput0_scrollbar0_node0.Content = fmt.Sprintf("%v", textinput0_scrollbar0_spacer())
		textinput0_scrollbar0_node1.Content = fmt.Sprintf("%v", textinput0_scrollbar0_startArrow())
		textinput0_scrollbar0_node2.Content = fmt.Sprintf("%v", textinput0_scrollbar0_trackText())
		textinput0_scrollbar0_node3.Content = fmt.Sprintf("%v", textinput0_scrollbar0_endArrow())
		textinput0_scrollbar0_node4.Content = fmt.Sprintf("%v", textinput0_scrollbar0_spacer())
		if focusIndex == 0 {
			textinput0_textedit0_box0.CursorCol = textinput0_textedit0_cursorX()
		} else {
			textinput0_textedit0_box0.CursorCol = -1
		}
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var prevTextinput0_selfW int
	var prevTextinput0_textedit0_selfW int
	var prevTextinput0_textedit0_selfX int
	var prevTextinput0_textedit0_selfY int
	var prevTextinput0_scrollbar0_selfX int
	var prevTextinput0_scrollbar0_selfY int
	var prevTextinput0_scrollbar0_selfW int
	var prevTextinput0_scrollbar0_selfH int
	doRender := func() {
		sync()
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		tree := layout.Layout(root, termW, termH)
		changes, scrollChanged := layout.DiffTrees(prevTree, tree)
		if prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			if app.TestBuffer != nil {
				app.TestBuffer = buf
			} else {
				render.ClearScreen(os.Stdout)
				buf.RenderTo(os.Stdout)
			}
		} else if app.TestBuffer != nil {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			app.TestBuffer = buf
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
		if textinput0_textedit0_selfW != prevTextinput0_textedit0_selfW {
			prevTextinput0_textedit0_selfW = textinput0_textedit0_selfW
			app.Dirty = true
		}
		if textinput0_textedit0_selfX != prevTextinput0_textedit0_selfX {
			prevTextinput0_textedit0_selfX = textinput0_textedit0_selfX
			app.Dirty = true
		}
		if textinput0_textedit0_selfY != prevTextinput0_textedit0_selfY {
			prevTextinput0_textedit0_selfY = textinput0_textedit0_selfY
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfX != prevTextinput0_scrollbar0_selfX {
			prevTextinput0_scrollbar0_selfX = textinput0_scrollbar0_selfX
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfY != prevTextinput0_scrollbar0_selfY {
			prevTextinput0_scrollbar0_selfY = textinput0_scrollbar0_selfY
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfW != prevTextinput0_scrollbar0_selfW {
			prevTextinput0_scrollbar0_selfW = textinput0_scrollbar0_selfW
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfH != prevTextinput0_scrollbar0_selfH {
			prevTextinput0_scrollbar0_selfH = textinput0_scrollbar0_selfH
			app.Dirty = true
		}
		if app.TestBuffer == nil {
			if cursorBox := layout.FindCursor(tree); cursorBox != nil {
				render.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)
			} else {
				render.HideCursor(os.Stdout)
			}
		}
	}

	_ = textinput0_leftIndicator
	_ = textinput0_rightIndicator
	_ = textinput0_scrollbarVisible
	_ = textinput0_textedit0_adjustView
	_ = textinput0_textedit0_clearSelection
	_ = textinput0_textedit0_selStart
	_ = textinput0_textedit0_selEnd
	_ = textinput0_textedit0_startSelection
	_ = textinput0_textedit0_maskedValue
	_ = textinput0_textedit0_displayValue
	_ = textinput0_textedit0_displayText
	_ = textinput0_textedit0_viewStart
	_ = textinput0_textedit0_viewEnd
	_ = textinput0_textedit0_visiblePreSel
	_ = textinput0_textedit0_visibleSel
	_ = textinput0_textedit0_visiblePostSel
	_ = textinput0_textedit0_visiblePlaceholder
	_ = textinput0_textedit0_cursorX
	_ = textinput0_textedit0_wordLeft
	_ = textinput0_textedit0_wordRight
	_ = textinput0_textedit0_wordLeftFrom
	_ = textinput0_textedit0_wordRightFrom
	_ = textinput0_textedit0_wordEndFrom
	_ = textinput0_textedit0_mousePosToTextPos
	_ = textinput0_textedit0_saveUndo
	_ = textinput0_textedit0_undo
	_ = textinput0_textedit0_deleteSelection
	_ = textinput0_textedit0_killText
	_ = textinput0_textedit0_transposeChars
	_ = textinput0_scrollbar0_isVisible
	_ = textinput0_scrollbar0_trackChar
	_ = textinput0_scrollbar0_thumbChar
	_ = textinput0_scrollbar0_startArrow
	_ = textinput0_scrollbar0_endArrow
	_ = textinput0_scrollbar0_trackSize
	_ = textinput0_scrollbar0_thumbSize
	_ = textinput0_scrollbar0_thumbPos
	_ = textinput0_scrollbar0_trackText
	_ = textinput0_scrollbar0_spacer
	_ = textinput0_scrollbar0_maxOffset
	_ = textinput0_scrollbar0_offsetForTrackPos
	_ = textinput0_scrollbar0_trackStart
	_ = textinput0_scrollbar0_mouseToTrackPos
	_ = textinput0_scrollbar0_isOnTrack
	_ = textinput0_scrollbar0_mouseTrackPos
	_ = textinput0_scrollbar0_cancelAnimation
	_ = textinput0_scrollbar0_animateStep
	_ = textinput0_scrollbar0_startAnimation
	_ = stopPropagation
	app = &tui.App{
		HasMouse: true,
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventSpecial {
				if evt.Special == input.KeyTab {
					prev := focusIndex
					focusIndex = (focusIndex+2)%(focusCount+1) - 1
					if prev >= 0 {
						dispatchToFocusable(prev, input.Event{Kind: input.EventBlur})
					}
					if focusIndex >= 0 {
						dispatchToFocusable(focusIndex, input.Event{Kind: input.EventFocus})
					}
					app.Dirty = true
					return
				}
				if evt.Special == input.KeyShiftTab {
					prev := focusIndex
					focusIndex = (focusIndex+focusCount+1)%(focusCount+1) - 1
					if prev >= 0 {
						dispatchToFocusable(prev, input.Event{Kind: input.EventBlur})
					}
					if focusIndex >= 0 {
						dispatchToFocusable(focusIndex, input.Event{Kind: input.EventFocus})
					}
					app.Dirty = true
					return
				}
			}
			propagationStopped = false
			switch focusIndex {
			case 0:
				textinput0_textedit0_handleEvent(evt)
			case 1:
				textinput0_scrollbar0_handleEvent(evt)
			}
			if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && !propagationStopped && focusIndex >= 0 {
				prev := focusIndex
				focusIndex = -1
				dispatchToFocusable(prev, input.Event{Kind: input.EventBlur})
				app.Dirty = true
			}
			if !propagationStopped {
				handleKey(evt)
			}
		},
	}
	app.Run()
}

func CreateApp(w, h int) *tui.App {
	value := ""

	var app *tui.App
	textinput0_viewOffset := 0
	textinput0_focused := false
	textinput0_placeholder := "Type here..."
	textinput0_inputType := ""
	textinput0_maxlength := 0
	textinput0_readonly := false
	textinput0_strip := false
	textinput0_selfW := 0
	textinput0_contentW := textinput0_selfW - 4
	textinput0_textedit0_cursor := 0
	textinput0_textedit0_killBuffer := ""
	textinput0_textedit0_undoValues := []string{}
	textinput0_textedit0_undoCursors := []int{}
	textinput0_textedit0_selAnchor := -1
	textinput0_textedit0_lastClickTime := int64(0)
	textinput0_textedit0_lastClickX := -1
	textinput0_textedit0_mouseDragging := false
	textinput0_textedit0_wordDragging := false
	textinput0_textedit0_wordDragStart := 0
	textinput0_textedit0_wordDragEnd := 0
	textinput0_textedit0_selfW := 0
	textinput0_textedit0_selfX := 0
	textinput0_textedit0_selfY := 0
	textinput0_textedit0_contentW := textinput0_textedit0_selfW
	textinput0_scrollbar0_animating := false
	textinput0_scrollbar0_dragging := false
	textinput0_scrollbar0_held := false
	textinput0_scrollbar0_animStart := int64(0)
	textinput0_scrollbar0_animFrom := 0
	textinput0_scrollbar0_animTo := 0
	textinput0_scrollbar0_direction := "horizontal"
	textinput0_scrollbar0_selfX := 0
	textinput0_scrollbar0_selfY := 0
	textinput0_scrollbar0_selfW := 0
	textinput0_scrollbar0_selfH := 0
	focusIndex := -1
	focusCount := 2
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

	textinput0_leftIndicator := func() string {
		if textinput0_viewOffset > 0 {
			return "<"
		}
		return " "
	}
	textinput0_rightIndicator := func() string {
		if textinput0_contentW > 0 && textinput0_viewOffset+textinput0_contentW < len(value) {
			return ">"
		}
		return " "
	}
	textinput0_scrollbarVisible := func() bool {
		return textinput0_focused && len(value) > textinput0_contentW && textinput0_contentW > 0
	}
	textinput0_textedit0_adjustView := func() {
		if textinput0_textedit0_contentW <= 0 {
			return
		}
		if textinput0_textedit0_cursor < textinput0_viewOffset {
			textinput0_viewOffset = textinput0_textedit0_cursor
			app.Dirty = true
		}
		if textinput0_textedit0_cursor > textinput0_viewOffset+textinput0_textedit0_contentW {
			textinput0_viewOffset = textinput0_textedit0_cursor - textinput0_textedit0_contentW
			app.Dirty = true
		}
		if textinput0_viewOffset < 0 {
			textinput0_viewOffset = 0
			app.Dirty = true
		}
	}
	textinput0_textedit0_clearSelection := func() {
		textinput0_textedit0_selAnchor = -1
		app.Dirty = true
	}
	textinput0_textedit0_selStart := func() int {
		if textinput0_textedit0_selAnchor < textinput0_textedit0_cursor {
			return textinput0_textedit0_selAnchor
		}
		return textinput0_textedit0_cursor
	}
	textinput0_textedit0_selEnd := func() int {
		if textinput0_textedit0_selAnchor > textinput0_textedit0_cursor {
			return textinput0_textedit0_selAnchor
		}
		return textinput0_textedit0_cursor
	}
	textinput0_textedit0_startSelection := func() {
		if textinput0_textedit0_selAnchor == -1 {
			textinput0_textedit0_selAnchor = textinput0_textedit0_cursor
			app.Dirty = true
		}
	}
	textinput0_textedit0_maskedValue := func() string {
		result := ""
		for i := 0; i < len(value); i++ {
			result = result + "*"
		}
		return result
	}
	textinput0_textedit0_displayValue := func() string {
		if textinput0_inputType == "password" {
			return textinput0_textedit0_maskedValue()
		}
		return value
	}
	textinput0_textedit0_displayText := func() string {
		text := textinput0_textedit0_displayValue()
		if len(text) == 0 && len(textinput0_placeholder) > 0 {
			text = textinput0_placeholder
		}
		return text
	}
	textinput0_textedit0_viewStart := func() int {
		vo := textinput0_viewOffset
		if len(value) == 0 {
			vo = 0
		}
		if vo < 0 {
			vo = 0
		}
		text := textinput0_textedit0_displayText()
		if vo > len(text) {
			vo = len(text)
		}
		return vo
	}
	textinput0_textedit0_viewEnd := func() int {
		vo := textinput0_textedit0_viewStart()
		end := vo + textinput0_textedit0_contentW
		text := textinput0_textedit0_displayText()
		if end > len(text) {
			end = len(text)
		}
		return end
	}
	textinput0_textedit0_visiblePreSel := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) == 0 {
			return ""
		}
		dv := textinput0_textedit0_displayValue()
		vs := textinput0_textedit0_viewStart()
		ve := textinput0_textedit0_viewEnd()
		if textinput0_textedit0_selAnchor == -1 {
			visible := dv[vs:ve]
			pad := ""
			for i := len(visible); i < textinput0_textedit0_contentW; i++ {
				pad = pad + " "
			}
			return visible + pad
		}
		ss := textinput0_textedit0_selStart()
		end := ss
		if end > ve {
			end = ve
		}
		if end <= vs {
			return ""
		}
		return dv[vs:end]
	}
	textinput0_textedit0_visibleSel := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) == 0 || textinput0_textedit0_selAnchor == -1 {
			return ""
		}
		dv := textinput0_textedit0_displayValue()
		vs := textinput0_textedit0_viewStart()
		ve := textinput0_textedit0_viewEnd()
		ss := textinput0_textedit0_selStart()
		se := textinput0_textedit0_selEnd()
		start := ss
		if start < vs {
			start = vs
		}
		end := se
		if end > ve {
			end = ve
		}
		if start >= end {
			return ""
		}
		return dv[start:end]
	}
	textinput0_textedit0_visiblePostSel := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) == 0 || textinput0_textedit0_selAnchor == -1 {
			return ""
		}
		dv := textinput0_textedit0_displayValue()
		vs := textinput0_textedit0_viewStart()
		ve := textinput0_textedit0_viewEnd()
		se := textinput0_textedit0_selEnd()
		start := se
		if start < vs {
			start = vs
		}
		if start >= ve {
			totalVisible := ve - vs
			pad := ""
			for i := totalVisible; i < textinput0_textedit0_contentW; i++ {
				pad = pad + " "
			}
			return pad
		}
		text := dv[start:ve]
		totalVisible := ve - vs
		pad := ""
		for i := totalVisible; i < textinput0_textedit0_contentW; i++ {
			pad = pad + " "
		}
		return text + pad
	}
	textinput0_textedit0_visiblePlaceholder := func() string {
		if textinput0_textedit0_contentW <= 0 || len(value) > 0 || len(textinput0_placeholder) == 0 {
			return ""
		}
		text := textinput0_placeholder
		if len(text) > textinput0_textedit0_contentW {
			text = text[:textinput0_textedit0_contentW]
		}
		pad := ""
		for i := len(text); i < textinput0_textedit0_contentW; i++ {
			pad = pad + " "
		}
		return text + pad
	}
	textinput0_textedit0_cursorX := func() int {
		if textinput0_textedit0_selAnchor >= 0 {
			return -1
		}
		return textinput0_textedit0_cursor - textinput0_viewOffset
	}
	textinput0_textedit0_wordLeft := func() int {
		pos := textinput0_textedit0_cursor
		for pos > 0 && value[pos-1] == ' ' {
			pos = pos - 1
		}
		for pos > 0 && value[pos-1] != ' ' {
			pos = pos - 1
		}
		return pos
	}
	textinput0_textedit0_wordRight := func() int {
		pos := textinput0_textedit0_cursor
		for pos < len(value) && value[pos] != ' ' {
			pos = pos + 1
		}
		for pos < len(value) && value[pos] == ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_textedit0_wordLeftFrom := func(pos int) int {
		for pos > 0 && value[pos-1] == ' ' {
			pos = pos - 1
		}
		for pos > 0 && value[pos-1] != ' ' {
			pos = pos - 1
		}
		return pos
	}
	textinput0_textedit0_wordRightFrom := func(pos int) int {
		for pos < len(value) && value[pos] != ' ' {
			pos = pos + 1
		}
		for pos < len(value) && value[pos] == ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_textedit0_wordEndFrom := func(pos int) int {
		for pos < len(value) && value[pos] != ' ' {
			pos = pos + 1
		}
		return pos
	}
	textinput0_textedit0_mousePosToTextPos := func(mouseX int) int {
		relX := mouseX - textinput0_textedit0_selfX
		pos := textinput0_viewOffset + relX
		if pos < 0 {
			pos = 0
		}
		if pos > len(value) {
			pos = len(value)
		}
		return pos
	}
	textinput0_textedit0_saveUndo := func() {
		textinput0_textedit0_undoValues = append(textinput0_textedit0_undoValues, value)
		app.Dirty = true
		textinput0_textedit0_undoCursors = append(textinput0_textedit0_undoCursors, textinput0_textedit0_cursor)
		app.Dirty = true
		if len(textinput0_textedit0_undoValues) > 100 {
			textinput0_textedit0_undoValues = textinput0_textedit0_undoValues[1:]
			app.Dirty = true
			textinput0_textedit0_undoCursors = textinput0_textedit0_undoCursors[1:]
			app.Dirty = true
		}
	}
	textinput0_textedit0_undo := func() {
		if len(textinput0_textedit0_undoValues) == 0 {
			return
		}
		last := len(textinput0_textedit0_undoValues) - 1
		value = textinput0_textedit0_undoValues[last]
		app.Dirty = true
		textinput0_textedit0_cursor = textinput0_textedit0_undoCursors[last]
		app.Dirty = true
		textinput0_textedit0_undoValues = textinput0_textedit0_undoValues[:last]
		app.Dirty = true
		textinput0_textedit0_undoCursors = textinput0_textedit0_undoCursors[:last]
		app.Dirty = true
		textinput0_textedit0_adjustView()
	}
	textinput0_textedit0_deleteSelection := func() {
		textinput0_textedit0_saveUndo()
		value = value[:textinput0_textedit0_selStart()] + value[textinput0_textedit0_selEnd():]
		app.Dirty = true
		textinput0_textedit0_cursor = textinput0_textedit0_selStart()
		app.Dirty = true
		textinput0_textedit0_selAnchor = -1
		app.Dirty = true
		textinput0_textedit0_adjustView()
	}
	textinput0_textedit0_killText := func(text string) {
		textinput0_textedit0_killBuffer = text
		app.Dirty = true
	}
	textinput0_textedit0_transposeChars := func() {
		if textinput0_textedit0_cursor < 2 {
			return
		}
		a := value[textinput0_textedit0_cursor-2]
		b := value[textinput0_textedit0_cursor-1]
		value = value[:textinput0_textedit0_cursor-2] + string(b) + string(a) + value[textinput0_textedit0_cursor:]
		app.Dirty = true
	}
	textinput0_textedit0_handleEvent := func(evt input.Event) {
		if evt.Kind == input.EventFocus {
			textinput0_focused = true
			app.Dirty = true
			stopPropagation()
			return
		}
		if evt.Kind == input.EventBlur {
			textinput0_focused = false
			app.Dirty = true
			textinput0_textedit0_clearSelection()
			if textinput0_strip && !textinput0_readonly {
				for len(value) > 0 && value[0] == ' ' {
					value = value[1:]
					app.Dirty = true
				}
				for len(value) > 0 && value[len(value)-1] == ' ' {
					value = value[:len(value)-1]
					app.Dirty = true
				}
				if textinput0_textedit0_cursor > len(value) {
					textinput0_textedit0_cursor = len(value)
					app.Dirty = true
				}
				textinput0_textedit0_adjustView()
			}
			stopPropagation()
			return
		}
		if evt.Kind == input.EventPaste {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			paste := evt.PasteText
			if textinput0_textedit0_selAnchor >= 0 {
				textinput0_textedit0_deleteSelection()
			}
			if textinput0_maxlength > 0 {
				room := textinput0_maxlength - len(value)
				if room <= 0 {
					stopPropagation()
					return
				}
				if len(paste) > room {
					paste = paste[:room]
				}
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + paste + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + len(paste)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseRelease {
			if textinput0_textedit0_mouseDragging || textinput0_textedit0_wordDragging {
				textinput0_textedit0_mouseDragging = false
				app.Dirty = true
				textinput0_textedit0_wordDragging = false
				app.Dirty = true
				if textinput0_textedit0_selAnchor == textinput0_textedit0_cursor {
					textinput0_textedit0_selAnchor = -1
					app.Dirty = true
				}
				stopPropagation()
				return
			}
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_textedit0_wordDragging {
			pos := textinput0_textedit0_mousePosToTextPos(evt.Mouse.X)
			wl := textinput0_textedit0_wordLeftFrom(pos)
			wr := textinput0_textedit0_wordEndFrom(pos)
			if wl < textinput0_textedit0_wordDragStart {
				textinput0_textedit0_selAnchor = textinput0_textedit0_wordDragEnd
				app.Dirty = true
				textinput0_textedit0_cursor = wl
				app.Dirty = true
			} else {
				textinput0_textedit0_selAnchor = textinput0_textedit0_wordDragStart
				app.Dirty = true
				textinput0_textedit0_cursor = wr
				app.Dirty = true
			}
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_textedit0_mouseDragging {
			pos := textinput0_textedit0_mousePosToTextPos(evt.Mouse.X)
			textinput0_textedit0_cursor = pos
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft {
			clickPos := textinput0_textedit0_mousePosToTextPos(evt.Mouse.X)
			now := time.Now().UnixMilli()
			if evt.Shift {
				if textinput0_textedit0_selAnchor == -1 {
					textinput0_textedit0_selAnchor = textinput0_textedit0_cursor
					app.Dirty = true
				}
				textinput0_textedit0_cursor = clickPos
				app.Dirty = true
				textinput0_textedit0_mouseDragging = true
				app.Dirty = true
				textinput0_textedit0_adjustView()
				stopPropagation()
				return
			}
			if now-textinput0_textedit0_lastClickTime < 500 && evt.Mouse.X == textinput0_textedit0_lastClickX {
				wl := textinput0_textedit0_wordLeftFrom(clickPos)
				wr := textinput0_textedit0_wordEndFrom(clickPos)
				textinput0_textedit0_selAnchor = wl
				app.Dirty = true
				textinput0_textedit0_cursor = wr
				app.Dirty = true
				textinput0_textedit0_wordDragging = true
				app.Dirty = true
				textinput0_textedit0_wordDragStart = wl
				app.Dirty = true
				textinput0_textedit0_wordDragEnd = wr
				app.Dirty = true
				textinput0_textedit0_adjustView()
				textinput0_textedit0_lastClickTime = 0
				app.Dirty = true
				stopPropagation()
				return
			}
			textinput0_textedit0_lastClickTime = now
			app.Dirty = true
			textinput0_textedit0_lastClickX = evt.Mouse.X
			app.Dirty = true
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = clickPos
			app.Dirty = true
			textinput0_textedit0_selAnchor = clickPos
			app.Dirty = true
			textinput0_textedit0_mouseDragging = true
			app.Dirty = true
			textinput0_textedit0_adjustView()
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
				stopPropagation()
				return
			}
			if evt.Mouse.Button == input.ScrollDown && textinput0_viewOffset < len(value)-textinput0_textedit0_contentW {
				textinput0_viewOffset = textinput0_viewOffset + 3
				app.Dirty = true
				if textinput0_viewOffset > len(value)-textinput0_textedit0_contentW {
					textinput0_viewOffset = len(value) - textinput0_textedit0_contentW
					app.Dirty = true
				}
				stopPropagation()
				return
			}
		}
		if (evt.Special == input.KeyBackspace || evt.Special == input.KeyDelete) && textinput0_textedit0_selAnchor >= 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_deleteSelection()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyBackspace && evt.Ctrl && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordLeft()
			textinput0_textedit0_killText(value[pos:textinput0_textedit0_cursor])
			value = value[:pos] + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = pos
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyBackspace && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor-1] + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyDelete && evt.Ctrl && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordRight()
			textinput0_textedit0_killText(value[textinput0_textedit0_cursor:pos])
			value = value[:textinput0_textedit0_cursor] + value[pos:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyDelete && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + value[textinput0_textedit0_cursor+1:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Shift && evt.Ctrl {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordLeft()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Shift && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Shift && evt.Ctrl {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordRight()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Shift && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyHome && evt.Shift {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = 0
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyEnd && evt.Shift {
			textinput0_textedit0_startSelection()
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && evt.Ctrl && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordLeft()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyLeft && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && evt.Ctrl && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordRight()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyRight && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyHome {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = 0
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Special == input.KeyEnd {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'a' {
			textinput0_textedit0_selAnchor = 0
			app.Dirty = true
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'e' {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = len(value)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'f' && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'b' && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_cursor - 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'c' && textinput0_textedit0_selAnchor >= 0 {
			selected := value[textinput0_textedit0_selStart():textinput0_textedit0_selEnd()]
			textinput0_textedit0_killText(selected)
			render.CopyToClipboard(os.Stdout, selected)
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'x' && textinput0_textedit0_selAnchor >= 0 {
			if textinput0_readonly {
				selected := value[textinput0_textedit0_selStart():textinput0_textedit0_selEnd()]
				textinput0_textedit0_killText(selected)
				render.CopyToClipboard(os.Stdout, selected)
				stopPropagation()
				return
			}
			selected := value[textinput0_textedit0_selStart():textinput0_textedit0_selEnd()]
			textinput0_textedit0_killText(selected)
			render.CopyToClipboard(os.Stdout, selected)
			textinput0_textedit0_deleteSelection()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'd' && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + value[textinput0_textedit0_cursor+1:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'k' && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			textinput0_textedit0_killText(value[textinput0_textedit0_cursor:])
			value = value[:textinput0_textedit0_cursor]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'u' && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			textinput0_textedit0_killText(value[:textinput0_textedit0_cursor])
			value = value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = 0
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'w' && textinput0_textedit0_cursor > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordLeft()
			textinput0_textedit0_killText(value[pos:textinput0_textedit0_cursor])
			value = value[:pos] + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = pos
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 't' && textinput0_textedit0_cursor >= 2 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			textinput0_textedit0_transposeChars()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'y' && len(textinput0_textedit0_killBuffer) > 0 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			yank := textinput0_textedit0_killBuffer
			if textinput0_maxlength > 0 {
				room := textinput0_maxlength - len(value)
				if room <= 0 {
					stopPropagation()
					return
				}
				if len(yank) > room {
					yank = yank[:room]
				}
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + yank + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + len(yank)
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == '/' {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_undo()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Alt && evt.Rune == 'f' && textinput0_textedit0_cursor < len(value) {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordRight()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Alt && evt.Rune == 'b' && textinput0_textedit0_cursor > 0 {
			textinput0_textedit0_clearSelection()
			textinput0_textedit0_cursor = textinput0_textedit0_wordLeft()
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && evt.Alt && evt.Rune == 'd' && textinput0_textedit0_cursor < len(value) {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			pos := textinput0_textedit0_wordRight()
			textinput0_textedit0_killText(value[textinput0_textedit0_cursor:pos])
			value = value[:textinput0_textedit0_cursor] + value[pos:]
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
		if evt.Kind == input.EventKey && !evt.Ctrl && !evt.Alt && evt.Rune >= 32 {
			if textinput0_readonly {
				stopPropagation()
				return
			}
			if textinput0_textedit0_selAnchor >= 0 {
				textinput0_textedit0_deleteSelection()
			}
			if textinput0_maxlength > 0 && len(value) >= textinput0_maxlength {
				stopPropagation()
				return
			}
			textinput0_textedit0_saveUndo()
			value = value[:textinput0_textedit0_cursor] + string(evt.Rune) + value[textinput0_textedit0_cursor:]
			app.Dirty = true
			textinput0_textedit0_cursor = textinput0_textedit0_cursor + 1
			app.Dirty = true
			textinput0_textedit0_adjustView()
			stopPropagation()
			return
		}
	}
	textinput0_scrollbar0_isVisible := func() bool {
		return textinput0_scrollbarVisible() && len(value) > textinput0_contentW && textinput0_contentW > 0
	}
	textinput0_scrollbar0_trackChar := func() string {
		if textinput0_scrollbar0_direction == "vertical" {
			return "|"
		}
		return "-"
	}
	textinput0_scrollbar0_thumbChar := func() string {
		return "#"
	}
	textinput0_scrollbar0_startArrow := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		if textinput0_scrollbar0_direction == "vertical" {
			return "^"
		}
		return "<"
	}
	textinput0_scrollbar0_endArrow := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		if textinput0_scrollbar0_direction == "vertical" {
			return "v"
		}
		return ">"
	}
	textinput0_scrollbar0_trackSize := func() int {
		if textinput0_scrollbar0_direction == "vertical" {
			return textinput0_scrollbar0_selfH - 2
		}
		return textinput0_scrollbar0_selfW - 2
	}
	textinput0_scrollbar0_thumbSize := func() int {
		ts := textinput0_scrollbar0_trackSize()
		if ts < 1 {
			return 1
		}
		size := ts * textinput0_contentW / len(value)
		if size < 1 {
			size = 1
		}
		return size
	}
	textinput0_scrollbar0_thumbPos := func() int {
		ts := textinput0_scrollbar0_trackSize()
		ths := textinput0_scrollbar0_thumbSize()
		maxOff := len(value) - textinput0_contentW
		if maxOff <= 0 {
			return 0
		}
		pos := (ts - ths) * textinput0_viewOffset / maxOff
		if pos < 0 {
			pos = 0
		}
		if pos+ths > ts {
			pos = ts - ths
		}
		return pos
	}
	textinput0_scrollbar0_trackText := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		ts := textinput0_scrollbar0_trackSize()
		if ts < 1 {
			return ""
		}
		tp := textinput0_scrollbar0_thumbPos()
		ths := textinput0_scrollbar0_thumbSize()
		tc := textinput0_scrollbar0_trackChar()
		thc := textinput0_scrollbar0_thumbChar()
		result := ""
		for i := 0; i < ts; i++ {
			if i >= tp && i < tp+ths {
				result = result + thc
			} else {
				result = result + tc
			}
		}
		return result
	}
	textinput0_scrollbar0_spacer := func() string {
		if !textinput0_scrollbar0_isVisible() {
			return ""
		}
		return " "
	}
	textinput0_scrollbar0_maxOffset := func() int {
		return len(value) - textinput0_contentW
	}
	textinput0_scrollbar0_offsetForTrackPos := func(pos int) int {
		ts := textinput0_scrollbar0_trackSize()
		if ts <= 0 {
			return textinput0_viewOffset
		}
		maxOff := textinput0_scrollbar0_maxOffset()
		if maxOff <= 0 {
			return 0
		}
		target := pos * maxOff / ts
		if target < 0 {
			target = 0
		}
		if target > maxOff {
			target = maxOff
		}
		return target
	}
	textinput0_scrollbar0_trackStart := func() int {
		if textinput0_scrollbar0_direction == "vertical" {
			return textinput0_scrollbar0_selfY + 1
		}
		return textinput0_scrollbar0_selfX + 1 + 1
	}
	textinput0_scrollbar0_mouseToTrackPos := func(mousePos int) int {
		return mousePos - textinput0_scrollbar0_trackStart()
	}
	textinput0_scrollbar0_isOnTrack := func(evt input.Event) bool {
		if textinput0_scrollbar0_direction == "vertical" {
			return evt.Mouse.X == textinput0_scrollbar0_selfX
		}
		return evt.Mouse.Y == textinput0_scrollbar0_selfY
	}
	textinput0_scrollbar0_mouseTrackPos := func(evt input.Event) int {
		if textinput0_scrollbar0_direction == "vertical" {
			return textinput0_scrollbar0_mouseToTrackPos(evt.Mouse.Y)
		}
		return textinput0_scrollbar0_mouseToTrackPos(evt.Mouse.X)
	}
	textinput0_scrollbar0_cancelAnimation := func() {
		textinput0_scrollbar0_animating = false
		app.Dirty = true
		textinput0_scrollbar0_dragging = false
		app.Dirty = true
		textinput0_scrollbar0_held = false
		app.Dirty = true
	}
	textinput0_scrollbar0_animateStep := func() {
		now := time.Now().UnixMilli()
		elapsed := now - textinput0_scrollbar0_animStart
		duration := int64(200)
		if elapsed >= duration {
			textinput0_viewOffset = textinput0_scrollbar0_animTo
			app.Dirty = true
			textinput0_scrollbar0_animating = false
			app.Dirty = true
			if textinput0_scrollbar0_held {
				textinput0_scrollbar0_dragging = true
				app.Dirty = true
			}
			return
		}
		t := float64(elapsed) / float64(duration)
		inv := 1.0 - t
		eased := 1.0 - inv*inv*inv
		textinput0_viewOffset = textinput0_scrollbar0_animFrom + int(eased*float64(textinput0_scrollbar0_animTo-textinput0_scrollbar0_animFrom))
		app.Dirty = true
		app.RequestFrame()
	}
	textinput0_scrollbar0_startAnimation := func(target int) {
		textinput0_scrollbar0_animating = true
		app.Dirty = true
		textinput0_scrollbar0_animStart = time.Now().UnixMilli()
		app.Dirty = true
		textinput0_scrollbar0_animFrom = textinput0_viewOffset
		app.Dirty = true
		textinput0_scrollbar0_animTo = target
		app.Dirty = true
		app.RequestFrame()
	}
	textinput0_scrollbar0_handleEvent := func(evt input.Event) {
		if evt.Kind == input.EventFrame {
			if textinput0_scrollbar0_animating {
				textinput0_scrollbar0_animateStep()
			}
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseRelease {
			if textinput0_scrollbar0_held || textinput0_scrollbar0_dragging {
				textinput0_scrollbar0_held = false
				app.Dirty = true
				textinput0_scrollbar0_dragging = false
				app.Dirty = true
				stopPropagation()
				return
			}
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MouseMotion && textinput0_scrollbar0_dragging {
			pos := textinput0_scrollbar0_mouseTrackPos(evt)
			textinput0_viewOffset = textinput0_scrollbar0_offsetForTrackPos(pos)
			app.Dirty = true
			stopPropagation()
			return
		}
		if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && evt.Mouse.Button == input.ButtonLeft && textinput0_scrollbar0_isOnTrack(evt) {
			pos := textinput0_scrollbar0_mouseTrackPos(evt)
			ts := textinput0_scrollbar0_trackSize()
			if pos >= 0 && pos < ts {
				target := textinput0_scrollbar0_offsetForTrackPos(pos)
				textinput0_scrollbar0_held = true
				app.Dirty = true
				textinput0_scrollbar0_startAnimation(target)
				stopPropagation()
				return
			}
		}
	}

	dispatchToFocusable := func(idx int, evt input.Event) {
		switch idx {
		case 0:
			textinput0_textedit0_handleEvent(evt)
		case 1:
			textinput0_scrollbar0_handleEvent(evt)
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Value: %v", value),
	}
	textinput0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_leftIndicator()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_textedit0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visiblePreSel()),
	}
	textinput0_textedit0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visibleSel()),
		Style: render.Style{
			Inverse: true,
		},
	}
	textinput0_textedit0_node2 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visiblePostSel()),
	}
	textinput0_textedit0_node3 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_textedit0_visiblePlaceholder()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_textedit0_box0 := &layout.Input{
		Kind:      layout.KindBox,
		Focusable: true,
		CursorCol: textinput0_textedit0_cursorX(),
		CursorRow: 0,
		Children: []*layout.Input{
			{
				Kind:      layout.KindBox,
				Direction: "row",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					textinput0_textedit0_node0,
					textinput0_textedit0_node1,
					textinput0_textedit0_node2,
					textinput0_textedit0_node3,
				},
			},
		},
	}
	textinput0_textedit0_box0.SelfW = &textinput0_textedit0_selfW
	textinput0_textedit0_box0.SelfX = &textinput0_textedit0_selfX
	textinput0_textedit0_box0.SelfY = &textinput0_textedit0_selfY
	textinput0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_rightIndicator()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_spacer()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_startArrow()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node2 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_trackText()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node3 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_endArrow()),
		Style: render.Style{
			Dim: true,
		},
	}
	textinput0_scrollbar0_node4 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", textinput0_scrollbar0_spacer()),
		Style: render.Style{
			Dim: true,
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
				CursorCol: -1,
				CursorRow: -1,
				Children: []*layout.Input{
					node0,
					{
						Kind:       layout.KindBox,
						FixedWidth: 24,
						CursorCol:  -1,
						CursorRow:  -1,
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
									textinput0_textedit0_box0,
									textinput0_node1,
									{
										Kind:    layout.KindText,
										Content: "]",
									},
								},
							},
							{
								Kind:      layout.KindBox,
								Direction: "row",
								Focusable: true,
								CursorCol: -1,
								CursorRow: -1,
								Children: []*layout.Input{
									textinput0_scrollbar0_node0,
									textinput0_scrollbar0_node1,
									textinput0_scrollbar0_node2,
									textinput0_scrollbar0_node3,
									textinput0_scrollbar0_node4,
								},
							},
						},
					},
				},
			},
		},
	}
	sync := func() {
		textinput0_contentW = textinput0_selfW - 4
		textinput0_textedit0_contentW = textinput0_textedit0_selfW
		textinput0_focused = focusIndex == 0
		node0.Content = fmt.Sprintf("Value: %v", value)
		textinput0_node0.Content = fmt.Sprintf("%v", textinput0_leftIndicator())
		textinput0_textedit0_node0.Content = fmt.Sprintf("%v", textinput0_textedit0_visiblePreSel())
		textinput0_textedit0_node1.Content = fmt.Sprintf("%v", textinput0_textedit0_visibleSel())
		textinput0_textedit0_node2.Content = fmt.Sprintf("%v", textinput0_textedit0_visiblePostSel())
		textinput0_textedit0_node3.Content = fmt.Sprintf("%v", textinput0_textedit0_visiblePlaceholder())
		textinput0_node1.Content = fmt.Sprintf("%v", textinput0_rightIndicator())
		textinput0_scrollbar0_node0.Content = fmt.Sprintf("%v", textinput0_scrollbar0_spacer())
		textinput0_scrollbar0_node1.Content = fmt.Sprintf("%v", textinput0_scrollbar0_startArrow())
		textinput0_scrollbar0_node2.Content = fmt.Sprintf("%v", textinput0_scrollbar0_trackText())
		textinput0_scrollbar0_node3.Content = fmt.Sprintf("%v", textinput0_scrollbar0_endArrow())
		textinput0_scrollbar0_node4.Content = fmt.Sprintf("%v", textinput0_scrollbar0_spacer())
		if focusIndex == 0 {
			textinput0_textedit0_box0.CursorCol = textinput0_textedit0_cursorX()
		} else {
			textinput0_textedit0_box0.CursorCol = -1
		}
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var prevTextinput0_selfW int
	var prevTextinput0_textedit0_selfW int
	var prevTextinput0_textedit0_selfX int
	var prevTextinput0_textedit0_selfY int
	var prevTextinput0_scrollbar0_selfX int
	var prevTextinput0_scrollbar0_selfY int
	var prevTextinput0_scrollbar0_selfW int
	var prevTextinput0_scrollbar0_selfH int
	doRender := func() {
		sync()
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		tree := layout.Layout(root, termW, termH)
		changes, scrollChanged := layout.DiffTrees(prevTree, tree)
		if prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap || prevTree.HasOverlap {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			if app.TestBuffer != nil {
				app.TestBuffer = buf
			} else {
				render.ClearScreen(os.Stdout)
				buf.RenderTo(os.Stdout)
			}
		} else if app.TestBuffer != nil {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			app.TestBuffer = buf
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
		if textinput0_textedit0_selfW != prevTextinput0_textedit0_selfW {
			prevTextinput0_textedit0_selfW = textinput0_textedit0_selfW
			app.Dirty = true
		}
		if textinput0_textedit0_selfX != prevTextinput0_textedit0_selfX {
			prevTextinput0_textedit0_selfX = textinput0_textedit0_selfX
			app.Dirty = true
		}
		if textinput0_textedit0_selfY != prevTextinput0_textedit0_selfY {
			prevTextinput0_textedit0_selfY = textinput0_textedit0_selfY
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfX != prevTextinput0_scrollbar0_selfX {
			prevTextinput0_scrollbar0_selfX = textinput0_scrollbar0_selfX
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfY != prevTextinput0_scrollbar0_selfY {
			prevTextinput0_scrollbar0_selfY = textinput0_scrollbar0_selfY
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfW != prevTextinput0_scrollbar0_selfW {
			prevTextinput0_scrollbar0_selfW = textinput0_scrollbar0_selfW
			app.Dirty = true
		}
		if textinput0_scrollbar0_selfH != prevTextinput0_scrollbar0_selfH {
			prevTextinput0_scrollbar0_selfH = textinput0_scrollbar0_selfH
			app.Dirty = true
		}
		if app.TestBuffer == nil {
			if cursorBox := layout.FindCursor(tree); cursorBox != nil {
				render.ShowCursor(os.Stdout, cursorBox.Y+cursorBox.CursorRow, cursorBox.X+cursorBox.CursorCol)
			} else {
				render.HideCursor(os.Stdout)
			}
		}
	}

	_ = textinput0_leftIndicator
	_ = textinput0_rightIndicator
	_ = textinput0_scrollbarVisible
	_ = textinput0_textedit0_adjustView
	_ = textinput0_textedit0_clearSelection
	_ = textinput0_textedit0_selStart
	_ = textinput0_textedit0_selEnd
	_ = textinput0_textedit0_startSelection
	_ = textinput0_textedit0_maskedValue
	_ = textinput0_textedit0_displayValue
	_ = textinput0_textedit0_displayText
	_ = textinput0_textedit0_viewStart
	_ = textinput0_textedit0_viewEnd
	_ = textinput0_textedit0_visiblePreSel
	_ = textinput0_textedit0_visibleSel
	_ = textinput0_textedit0_visiblePostSel
	_ = textinput0_textedit0_visiblePlaceholder
	_ = textinput0_textedit0_cursorX
	_ = textinput0_textedit0_wordLeft
	_ = textinput0_textedit0_wordRight
	_ = textinput0_textedit0_wordLeftFrom
	_ = textinput0_textedit0_wordRightFrom
	_ = textinput0_textedit0_wordEndFrom
	_ = textinput0_textedit0_mousePosToTextPos
	_ = textinput0_textedit0_saveUndo
	_ = textinput0_textedit0_undo
	_ = textinput0_textedit0_deleteSelection
	_ = textinput0_textedit0_killText
	_ = textinput0_textedit0_transposeChars
	_ = textinput0_scrollbar0_isVisible
	_ = textinput0_scrollbar0_trackChar
	_ = textinput0_scrollbar0_thumbChar
	_ = textinput0_scrollbar0_startArrow
	_ = textinput0_scrollbar0_endArrow
	_ = textinput0_scrollbar0_trackSize
	_ = textinput0_scrollbar0_thumbSize
	_ = textinput0_scrollbar0_thumbPos
	_ = textinput0_scrollbar0_trackText
	_ = textinput0_scrollbar0_spacer
	_ = textinput0_scrollbar0_maxOffset
	_ = textinput0_scrollbar0_offsetForTrackPos
	_ = textinput0_scrollbar0_trackStart
	_ = textinput0_scrollbar0_mouseToTrackPos
	_ = textinput0_scrollbar0_isOnTrack
	_ = textinput0_scrollbar0_mouseTrackPos
	_ = textinput0_scrollbar0_cancelAnimation
	_ = textinput0_scrollbar0_animateStep
	_ = textinput0_scrollbar0_startAnimation
	_ = stopPropagation
	app = &tui.App{
		HasMouse: true,
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventSpecial {
				if evt.Special == input.KeyTab {
					prev := focusIndex
					focusIndex = (focusIndex+2)%(focusCount+1) - 1
					if prev >= 0 {
						dispatchToFocusable(prev, input.Event{Kind: input.EventBlur})
					}
					if focusIndex >= 0 {
						dispatchToFocusable(focusIndex, input.Event{Kind: input.EventFocus})
					}
					app.Dirty = true
					return
				}
				if evt.Special == input.KeyShiftTab {
					prev := focusIndex
					focusIndex = (focusIndex+focusCount+1)%(focusCount+1) - 1
					if prev >= 0 {
						dispatchToFocusable(prev, input.Event{Kind: input.EventBlur})
					}
					if focusIndex >= 0 {
						dispatchToFocusable(focusIndex, input.Event{Kind: input.EventFocus})
					}
					app.Dirty = true
					return
				}
			}
			propagationStopped = false
			switch focusIndex {
			case 0:
				textinput0_textedit0_handleEvent(evt)
			case 1:
				textinput0_scrollbar0_handleEvent(evt)
			}
			if evt.Kind == input.EventMouse && evt.Mouse.Action == input.MousePress && !propagationStopped && focusIndex >= 0 {
				prev := focusIndex
				focusIndex = -1
				dispatchToFocusable(prev, input.Event{Kind: input.EventBlur})
				app.Dirty = true
			}
			if !propagationStopped {
				handleKey(evt)
			}
		},
	}
	app.TestWidth = w
	app.TestHeight = h
	app.TestBuffer = render.NewBuffer(w, h)
	app.Render()
	return app
}
