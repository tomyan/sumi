package preview

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type PreviewProps struct {
}

func NewPreview(props PreviewProps) *tui.Component {
	current := signal.New(0)
	matchStatus := signal.New(0)
	interactive := signal.New(false)
	focusState := signal.New(0)

	doStep := func(index int) {
		pvStepTo(index)
		matchStatus.Set(pvMatches(index))
		current.Set(index)
	}

	doCommand := func(cmd string) {
		if cmd == "quit" {
			tui.Quit()
			return
		}
		if cmd == "exit" {
			if pvFocus == FocusInteractive {
				pvExitInteractive()
				interactive.Set(false)
				pvStepTo(current.Get())
				matchStatus.Set(pvMatches(current.Get()))
			}
			pvFocus = FocusControls
			focusState.Set(int(FocusControls))
			return
		}
		if cmd == "next" {
			if current.Get() < pvStepCount()-1 {
				doStep(current.Get() + 1)
			}
			return
		}
		if cmd == "prev" {
			if current.Get() > 0 {
				doStep(current.Get() - 1)
			}
			return
		}
		if cmd == "update" {
			pvUpdateSnapshot(current.Get())
			matchStatus.Set(pvMatches(current.Get()))
			return
		}
		if cmd == "interactive" {
			pvEnterInteractive()
			interactive.Set(true)
			pvFocus = FocusInteractive
			focusState.Set(int(FocusInteractive))
			return
		}
	}

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if pvPrefixPending {
			pvPrefixPending = false
			cmd := prefixCommand(evt)
			if cmd == "" {
				pvFocus = FocusControls
				focusState.Set(int(FocusControls))
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
				interactive.Set(false)
				pvFocus = FocusControls
				focusState.Set(int(FocusControls))
				pvStepTo(current.Get())
				matchStatus.Set(pvMatches(current.Get()))
				return
			}
			pvSendInput(evt)
			matchStatus.Set(0)
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
			tui.Quit()
			return
		}
		if (evt.Kind == input.EventSpecial && evt.Special == input.KeyRight) || (evt.Kind == input.EventKey && (evt.Rune == 'l' || evt.Rune == '\r')) {
			if current.Get() < pvStepCount()-1 {
				doStep(current.Get() + 1)
			}
			return
		}
		if (evt.Kind == input.EventSpecial && evt.Special == input.KeyLeft) || (evt.Kind == input.EventKey && evt.Rune == 'h') {
			if current.Get() > 0 {
				doStep(current.Get() - 1)
			}
			return
		}
		if evt.Kind == input.EventKey && evt.Rune == 'u' {
			pvUpdateSnapshot(current.Get())
			matchStatus.Set(pvMatches(current.Get()))
			return
		}
		if evt.Kind == input.EventKey && evt.Rune == 'i' {
			pvEnterInteractive()
			interactive.Set(true)
			pvFocus = FocusInteractive
			focusState.Set(int(FocusInteractive))
			return
		}
		if evt.Kind == input.EventKey && (evt.Rune == '1' || evt.Rune == '2' || evt.Rune == '3') {
			pvFocus = focusForDigit(evt.Rune)
			focusState.Set(int(pvFocus))
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
						Kind:      layout.KindBox,
						Direction: "row",
						FlexGrow:  1,
						CursorCol: -1,
						CursorRow: -1,
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
						FlexGrow:    1,
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

	signal.Effect(func() {
		box0.Children = func() []*layout.Input {
			var cs []*layout.Input
			if interactive.Get() {
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
					Content: fmt.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current.Get()+1, pvStepCount(), pvStepName(current.Get())),
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
				if matchStatus.Get() == 1 {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: " MATCH ",
						Style: render.Style{
							FG:   render.Color{Name: "green"},
							Bold: true,
						},
					})
				} else {
					if matchStatus.Get() == 0 {
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
					Content: fmt.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current.Get()+1, pvStepCount(), pvStepName(current.Get())),
					Style: render.Style{
						Bold: true,
					},
				})
			}
			return cs
		}()
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
