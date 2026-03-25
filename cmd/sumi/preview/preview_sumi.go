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
	interactive := false
	focusState := 0
	_ = focusState

	var app *tui.App
	doStep := func(index int) {
		pvStepTo(index)
		matchStatus = pvMatches(index)
		app.Dirty = true
		current = index
		app.Dirty = true
	}
	doCommand := func(cmd string) {
		if cmd == "quit" {
			app.Quit()
			return
		}
		if cmd == "exit" {
			if pvFocus == FocusInteractive {
				pvExitInteractive()
				interactive = false
				app.Dirty = true
				pvStepTo(current)
				matchStatus = pvMatches(current)
				app.Dirty = true
			}
			pvFocus = FocusControls
			focusState = int(FocusControls)
			app.Dirty = true
			return
		}
		if cmd == "next" {
			if current < pvStepCount()-1 {
				doStep(current + 1)
			}
			return
		}
		if cmd == "prev" {
			if current > 0 {
				doStep(current - 1)
			}
			return
		}
		if cmd == "update" {
			pvUpdateSnapshot(current)
			matchStatus = pvMatches(current)
			app.Dirty = true
			return
		}
		if cmd == "interactive" {
			pvEnterInteractive()
			interactive = true
			app.Dirty = true
			pvFocus = FocusInteractive
			focusState = int(FocusInteractive)
			app.Dirty = true
			return
		}
	}
	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if pvPrefixPending {
			pvPrefixPending = false
			cmd := prefixCommand(evt)
			if cmd == "" {
				pvFocus = FocusControls
				focusState = int(FocusControls)
				app.Dirty = true
			} else {
				doCommand(cmd)
			}
			return
		}
		if pvFocus == FocusInteractive {
			if isCtrlBackslash(evt) {
				pvPrefixPending = true
				return
			}
			if (evt.Kind == input.EventSpecial && evt.Special == input.KeyEscape) || evt.Alt {
				pvExitInteractive()
				interactive = false
				app.Dirty = true
				pvFocus = FocusControls
				focusState = int(FocusControls)
				app.Dirty = true
				pvStepTo(current)
				matchStatus = pvMatches(current)
				app.Dirty = true
				return
			}
			pvSendInput(evt)
			matchStatus = 0
			app.Dirty = true
			return
		}
		if pvIsEditorFocused() {
			if isCtrlBackslash(evt) {
				pvPrefixPending = true
				return
			}
			idx := editorIndex(pvFocus)
			if idx >= 0 && pvEditors[idx] != nil {
				pvEditors[idx].SendEvent(evt)
			}
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
		if evt.Kind == input.EventKey && evt.Rune == 'i' {
			pvEnterInteractive()
			interactive = true
			app.Dirty = true
			pvFocus = FocusInteractive
			focusState = int(FocusInteractive)
			app.Dirty = true
			return
		}
		if evt.Kind == input.EventKey && (evt.Rune == '1' || evt.Rune == '2' || evt.Rune == '3') {
			pvFocus = focusForDigit(evt.Rune)
			focusState = int(pvFocus)
			app.Dirty = true
			return
		}
	}

	box0 := &layout.Input{
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
						Kind:      layout.KindBox,
						Direction: "row",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Actual",
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "cyan"},
								},
							},
							{
								Kind:        layout.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Expected",
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "blue"},
								},
							},
						},
					},
					box0,
					{
						Kind:        layout.KindBox,
						Direction:   "row",
						FixedHeight: pvEditorHeight(),
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FlexGrow:    1,
								Border:      "single",
								BorderTitle: pvSourceTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "magenta"},
								},
							},
							{
								Kind:        layout.KindBox,
								FlexGrow:    1,
								Border:      "single",
								BorderTitle: pvSnapshotTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "magenta"},
								},
							},
						},
					},
					{
						Kind:        layout.KindBox,
						FixedHeight: pvEditorHeight(),
						Border:      "single",
						BorderTitle: pvScenarioTitle(),
						CursorCol:   -1,
						CursorRow:   -1,
						Style: render.Style{
							FG: render.Color{Name: "magenta"},
						},
					},
				},
			},
		},
	}
	sync := func() {
		box0.Children = func() []*layout.Input {
			var cs []*layout.Input
			if interactive {
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: "esc",
					Style: render.Style{
						Inverse: true,
					},
				})
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: " Exit  ",
					Style: render.Style{
						Dim: true,
					},
				})
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: " INTERACTIVE ",
					Style: render.Style{
						FG:   render.Color{Name: "green"},
						Bold: true,
					},
				})
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: fmt.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current+1, pvStepCount(), pvStepName(current)),
					Style: render.Style{
						Bold: true,
					},
				})
			} else {
				if pvIsEditorFocused() {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: fmt.Sprintf(" %v ", pvFocusName()),
						Style: render.Style{
							FG:   render.Color{Name: "cyan"},
							Bold: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "C-\\",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Exit  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "C-\\ h",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Prev  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "C-\\ l",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Next  ",
						Style: render.Style{
							Dim: true,
						},
					})
				} else {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "h",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Prev  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "l",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Next  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "u",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Update  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "i",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Interactive  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "1",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Source  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "2",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Snap  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "3",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Scenario  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "q",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Quit  ",
						Style: render.Style{
							Dim: true,
						},
					})
				}
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
			}
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
	_ = doCommand
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
	interactive := false
	focusState := 0
	_ = focusState

	var app *tui.App
	doStep := func(index int) {
		pvStepTo(index)
		matchStatus = pvMatches(index)
		app.Dirty = true
		current = index
		app.Dirty = true
	}
	doCommand := func(cmd string) {
		if cmd == "quit" {
			app.Quit()
			return
		}
		if cmd == "exit" {
			if pvFocus == FocusInteractive {
				pvExitInteractive()
				interactive = false
				app.Dirty = true
				pvStepTo(current)
				matchStatus = pvMatches(current)
				app.Dirty = true
			}
			pvFocus = FocusControls
			focusState = int(FocusControls)
			app.Dirty = true
			return
		}
		if cmd == "next" {
			if current < pvStepCount()-1 {
				doStep(current + 1)
			}
			return
		}
		if cmd == "prev" {
			if current > 0 {
				doStep(current - 1)
			}
			return
		}
		if cmd == "update" {
			pvUpdateSnapshot(current)
			matchStatus = pvMatches(current)
			app.Dirty = true
			return
		}
		if cmd == "interactive" {
			pvEnterInteractive()
			interactive = true
			app.Dirty = true
			pvFocus = FocusInteractive
			focusState = int(FocusInteractive)
			app.Dirty = true
			return
		}
	}
	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			app.Quit()
			return
		}
		if pvPrefixPending {
			pvPrefixPending = false
			cmd := prefixCommand(evt)
			if cmd == "" {
				pvFocus = FocusControls
				focusState = int(FocusControls)
				app.Dirty = true
			} else {
				doCommand(cmd)
			}
			return
		}
		if pvFocus == FocusInteractive {
			if isCtrlBackslash(evt) {
				pvPrefixPending = true
				return
			}
			if (evt.Kind == input.EventSpecial && evt.Special == input.KeyEscape) || evt.Alt {
				pvExitInteractive()
				interactive = false
				app.Dirty = true
				pvFocus = FocusControls
				focusState = int(FocusControls)
				app.Dirty = true
				pvStepTo(current)
				matchStatus = pvMatches(current)
				app.Dirty = true
				return
			}
			pvSendInput(evt)
			matchStatus = 0
			app.Dirty = true
			return
		}
		if pvIsEditorFocused() {
			if isCtrlBackslash(evt) {
				pvPrefixPending = true
				return
			}
			idx := editorIndex(pvFocus)
			if idx >= 0 && pvEditors[idx] != nil {
				pvEditors[idx].SendEvent(evt)
			}
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
		if evt.Kind == input.EventKey && evt.Rune == 'i' {
			pvEnterInteractive()
			interactive = true
			app.Dirty = true
			pvFocus = FocusInteractive
			focusState = int(FocusInteractive)
			app.Dirty = true
			return
		}
		if evt.Kind == input.EventKey && (evt.Rune == '1' || evt.Rune == '2' || evt.Rune == '3') {
			pvFocus = focusForDigit(evt.Rune)
			focusState = int(pvFocus)
			app.Dirty = true
			return
		}
	}

	box0 := &layout.Input{
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
						Kind:      layout.KindBox,
						Direction: "row",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Actual",
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "cyan"},
								},
							},
							{
								Kind:        layout.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     layout.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Expected",
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "blue"},
								},
							},
						},
					},
					box0,
					{
						Kind:        layout.KindBox,
						Direction:   "row",
						FixedHeight: pvEditorHeight(),
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*layout.Input{
							{
								Kind:        layout.KindBox,
								FlexGrow:    1,
								Border:      "single",
								BorderTitle: pvSourceTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "magenta"},
								},
							},
							{
								Kind:        layout.KindBox,
								FlexGrow:    1,
								Border:      "single",
								BorderTitle: pvSnapshotTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
								Style: render.Style{
									FG: render.Color{Name: "magenta"},
								},
							},
						},
					},
					{
						Kind:        layout.KindBox,
						FixedHeight: pvEditorHeight(),
						Border:      "single",
						BorderTitle: pvScenarioTitle(),
						CursorCol:   -1,
						CursorRow:   -1,
						Style: render.Style{
							FG: render.Color{Name: "magenta"},
						},
					},
				},
			},
		},
	}
	sync := func() {
		box0.Children = func() []*layout.Input {
			var cs []*layout.Input
			if interactive {
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: "esc",
					Style: render.Style{
						Inverse: true,
					},
				})
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: " Exit  ",
					Style: render.Style{
						Dim: true,
					},
				})
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: " INTERACTIVE ",
					Style: render.Style{
						FG:   render.Color{Name: "green"},
						Bold: true,
					},
				})
				cs = append(cs, &layout.Input{
					Kind:    layout.KindText,
					Content: fmt.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current+1, pvStepCount(), pvStepName(current)),
					Style: render.Style{
						Bold: true,
					},
				})
			} else {
				if pvIsEditorFocused() {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: fmt.Sprintf(" %v ", pvFocusName()),
						Style: render.Style{
							FG:   render.Color{Name: "cyan"},
							Bold: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "C-\\",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Exit  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "C-\\ h",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Prev  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "C-\\ l",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Next  ",
						Style: render.Style{
							Dim: true,
						},
					})
				} else {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "h",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Prev  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "l",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Next  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "u",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Update  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "i",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Interactive  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "1",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Source  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "2",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Snap  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "3",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Scenario  ",
						Style: render.Style{
							Dim: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: "q",
						Style: render.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " Quit  ",
						Style: render.Style{
							Dim: true,
						},
					})
				}
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
			}
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
	_ = doCommand
	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			handleKey(evt)
		},
	}
	app.TestWidth = w
	app.TestHeight = h
	if w > 0 && h > 0 {
		app.TestBuffer = render.NewBuffer(w, h)
		app.Render()
	}
	return app
}
