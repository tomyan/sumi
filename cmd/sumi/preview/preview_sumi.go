package preview

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
	current := 0
	matchStatus := 0

	var app *tui.App
	splitpanel0_leftTitle := "Actual"
	splitpanel0_rightTitle := "Expected"
	doStep := func(index int) {
		pvStepTo(index)
		matchStatus = pvMatches(index)
		app.Dirty = true
		current = index
		app.Dirty = true
	}
	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			app.Quit()
			return
		}
		if (evt.Kind == input.EventSpecial && evt.Special == input.KeyRight) || (evt.Kind == input.EventKey && (evt.Rune == 'l' || evt.Rune == '\r')) {
			if current < pvStepCount()-1 {
				doStep(current + 1)
			}
			return
		}
		if (evt.Kind == input.EventSpecial && evt.Special == input.KeyLeft) || (evt.Kind == input.EventKey && evt.Rune == 'h') {
			if current > 0 {
				doStep(current - 1)
			}
			return
		}
		if evt.Kind == input.EventKey && evt.Rune == 'u' {
			pvUpdateSnapshot(current)
			matchStatus = pvMatches(current)
			app.Dirty = true
			return
		}
	}

	box0 := &layout.Input{
		Kind:        layout.KindBox,
		Border:      "single",
		BorderTitle: pvSourceTitle(),
		Overflow:    "scroll",
		CursorCol:   -1,
		CursorRow:   -1,
	}
	box1 := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "row",
		CursorCol: -1,
		CursorRow: -1,
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
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_leftTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
							{
								Kind:        layout.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_rightTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
						},
					},
					box0,
					box1,
					{
						Kind:      layout.KindBox,
						Direction: "row",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "h",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Prev  ",
								Style: render.Style{
									Dim: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "l",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Next  ",
								Style: render.Style{
									Dim: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "u",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Update  ",
								Style: render.Style{
									Dim: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "q",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Quit",
								Style: render.Style{
									Dim: true,
								},
							},
						},
					},
				},
			},
		},
	}
	sync := func() {
		box0.Children = func() []*layout.Input {
			var cs []*layout.Input
			for i, line := range pvSourceLines {
				cs = append(cs, &layout.Input{
					Kind:      layout.KindBox,
					Direction: "row",
					CursorCol: -1,
					CursorRow: -1,
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: fmt.Sprintf("%v", fmt.Sprintf("%3d ", i+1)),
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: fmt.Sprintf("%v", line),
						},
					},
				})
			}
			return cs
		}()
		box1.Children = func() []*layout.Input {
			var cs []*layout.Input
			if matchStatus == 1 {
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: " MATCH ",
					Style: render.Style{
						FG:   render.Color{Name: "green"},
						Bold: true,
					},
				})
			} else {
				if matchStatus == 0 {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " NO SNAPSHOT ",
						Style: render.Style{
							FG:   render.Color{Name: "yellow"},
							Bold: true,
						},
					})
				} else {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " DIFF ",
						Style: render.Style{
							FG:   render.Color{Name: "red"},
							Bold: true,
						},
					})
				}
			}
			cs = append(cs, &layout.Input{
				Kind:    layout.KindText,
				Content: fmt.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current+1, pvStepCount(), pvStepName(current)),
				Style: render.Style{
					Bold: true,
				},
			})
			return cs
		}()
	}

	var prevTree *layout.Box
	var prevW, prevH int
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
	}

	_ = doStep
	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			handleKey(evt)
		},
	}
	app.Run()
}

func CreateApp(w, h int) *tui.App {
	current := 0
	matchStatus := 0

	var app *tui.App
	splitpanel0_leftTitle := "Actual"
	splitpanel0_rightTitle := "Expected"
	doStep := func(index int) {
		pvStepTo(index)
		matchStatus = pvMatches(index)
		app.Dirty = true
		current = index
		app.Dirty = true
	}
	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			app.Quit()
			return
		}
		if (evt.Kind == input.EventSpecial && evt.Special == input.KeyRight) || (evt.Kind == input.EventKey && (evt.Rune == 'l' || evt.Rune == '\r')) {
			if current < pvStepCount()-1 {
				doStep(current + 1)
			}
			return
		}
		if (evt.Kind == input.EventSpecial && evt.Special == input.KeyLeft) || (evt.Kind == input.EventKey && evt.Rune == 'h') {
			if current > 0 {
				doStep(current - 1)
			}
			return
		}
		if evt.Kind == input.EventKey && evt.Rune == 'u' {
			pvUpdateSnapshot(current)
			matchStatus = pvMatches(current)
			app.Dirty = true
			return
		}
	}

	box0 := &layout.Input{
		Kind:        layout.KindBox,
		Border:      "single",
		BorderTitle: pvSourceTitle(),
		Overflow:    "scroll",
		CursorCol:   -1,
		CursorRow:   -1,
	}
	box1 := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "row",
		CursorCol: -1,
		CursorRow: -1,
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
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_leftTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
							{
								Kind:        layout.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: splitpanel0_rightTitle,
								CursorCol:   -1,
								CursorRow:   -1,
							},
						},
					},
					box0,
					box1,
					{
						Kind:      layout.KindBox,
						Direction: "row",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*layout.Input{
							{
								Kind:    layout.KindText,
								Content: "h",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Prev  ",
								Style: render.Style{
									Dim: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "l",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Next  ",
								Style: render.Style{
									Dim: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "u",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Update  ",
								Style: render.Style{
									Dim: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: "q",
								Style: render.Style{
									Inverse: true,
								},
							},
							{
								Kind:    layout.KindText,
								Content: " Quit",
								Style: render.Style{
									Dim: true,
								},
							},
						},
					},
				},
			},
		},
	}
	sync := func() {
		box0.Children = func() []*layout.Input {
			var cs []*layout.Input
			for i, line := range pvSourceLines {
				cs = append(cs, &layout.Input{
					Kind:      layout.KindBox,
					Direction: "row",
					CursorCol: -1,
					CursorRow: -1,
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: fmt.Sprintf("%v", fmt.Sprintf("%3d ", i+1)),
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: fmt.Sprintf("%v", line),
						},
					},
				})
			}
			return cs
		}()
		box1.Children = func() []*layout.Input {
			var cs []*layout.Input
			if matchStatus == 1 {
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: " MATCH ",
					Style: render.Style{
						FG:   render.Color{Name: "green"},
						Bold: true,
					},
				})
			} else {
				if matchStatus == 0 {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " NO SNAPSHOT ",
						Style: render.Style{
							FG:   render.Color{Name: "yellow"},
							Bold: true,
						},
					})
				} else {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " DIFF ",
						Style: render.Style{
							FG:   render.Color{Name: "red"},
							Bold: true,
						},
					})
				}
			}
			cs = append(cs, &layout.Input{
				Kind:    layout.KindText,
				Content: fmt.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current+1, pvStepCount(), pvStepName(current)),
				Style: render.Style{
					Bold: true,
				},
			})
			return cs
		}()
	}

	var prevTree *layout.Box
	var prevW, prevH int
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
	}

	_ = doStep
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
