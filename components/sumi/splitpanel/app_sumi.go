package splitpanel

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

func Run() {
	var app *tui.App
	splitpanel0_leftTitle := "Actual"
	splitpanel0_rightTitle := "Expected"
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
					{
						Kind:           layout.KindBox,
						Direction:      "row",
						BorderCollapse: true,
						CursorCol:      -1,
						CursorRow:      -1,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FixedHeight: 5,
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_leftTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
							{
								Kind:        layout.KindBox,
								FixedHeight: 5,
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_rightTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
						},
					},
				},
			},
		},
	}
	sync := func() []*layout.Input {
		var changed []*layout.Input
		return changed
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var nodeBoxMap map[*layout.Input]*layout.Box
	doRender := func() {
		changed := sync()
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		if prevTree != nil && len(changed) == 0 && termW == prevW && termH == prevH {
			return
		}
		if app.TestBuffer == nil && prevTree != nil && len(changed) > 0 && termW == prevW && termH == prevH && !prevTree.HasOverlap && nodeBoxMap != nil {
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
	var app *tui.App
	splitpanel0_leftTitle := "Actual"
	splitpanel0_rightTitle := "Expected"
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
					{
						Kind:           layout.KindBox,
						Direction:      "row",
						BorderCollapse: true,
						CursorCol:      -1,
						CursorRow:      -1,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FixedHeight: 5,
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_leftTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
							{
								Kind:        layout.KindBox,
								FixedHeight: 5,
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_rightTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
						},
					},
				},
			},
		},
	}
	sync := func() []*layout.Input {
		var changed []*layout.Input
		return changed
	}

	var prevTree *layout.Box
	var prevW, prevH int
	var nodeBoxMap map[*layout.Input]*layout.Box
	doRender := func() {
		changed := sync()
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = term.GetSize(int(os.Stdin.Fd()))
		}
		if prevTree != nil && len(changed) == 0 && termW == prevW && termH == prevH {
			return
		}
		if app.TestBuffer == nil && prevTree != nil && len(changed) > 0 && termW == prevW && termH == prevH && !prevTree.HasOverlap && nodeBoxMap != nil {
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
	}

	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			handleKey(evt)
		},
	}
	app.TestWidth = w
	app.TestHeight = h
	app.TestBuffer = render.NewBuffer(w, h)
	app.Render()
	return app
}
