package preview

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type PreviewProps struct {
}

func NewPreview(props PreviewProps) *sumi.Component {
	current := sumi.New(0)
	matchStatus := sumi.New(0)
	interactive := sumi.New(false)
	focusState := sumi.New(0)

	doStep := func(index int) {
		pvStepTo(index)
		matchStatus.Set(pvMatches(index))
		current.Set(index)
	}

	doCommand := func(cmd string) {
		if cmd == "quit" {
			sumi.Quit()
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

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
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
			if (evt.Kind == sumi.EventSpecial && evt.Special == sumi.KeyEscape) || evt.Alt {
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
			sumi.Quit()
			return
		}
		if (evt.Kind == sumi.EventSpecial && evt.Special == sumi.KeyRight) || (evt.Kind == sumi.EventKey && (evt.Rune == 'l' || evt.Rune == '\r')) {
			if current.Get() < pvStepCount()-1 {
				doStep(current.Get() + 1)
			}
			return
		}
		if (evt.Kind == sumi.EventSpecial && evt.Special == sumi.KeyLeft) || (evt.Kind == sumi.EventKey && evt.Rune == 'h') {
			if current.Get() > 0 {
				doStep(current.Get() - 1)
			}
			return
		}
		if evt.Kind == sumi.EventKey && evt.Rune == 'u' {
			pvUpdateSnapshot(current.Get())
			matchStatus.Set(pvMatches(current.Get()))
			return
		}
		if evt.Kind == sumi.EventKey && evt.Rune == 'i' {
			pvEnterInteractive()
			interactive.Set(true)
			pvFocus = FocusInteractive
			focusState.Set(int(FocusInteractive))
			return
		}
		if evt.Kind == sumi.EventKey && (evt.Rune == '1' || evt.Rune == '2' || evt.Rune == '3') {
			pvFocus = focusForDigit(evt.Rune)
			focusState.Set(int(pvFocus))
			return
		}
	}

	box0 := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "row",
		CursorCol: -1,
		CursorRow: -1,
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Direction: "row",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     sumi.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Actual",
								CursorCol:   -1,
								CursorRow:   -1,
								Style: sumi.Style{
									FG: sumi.Color{Name: "cyan"},
								},
							},
							{
								Kind:        sumi.KindBox,
								FixedHeight: pvComponentHeight(),
								FlexGrow:    1,
								Padding:     sumi.ParsePadding("0 1"),
								Border:      "single",
								BorderTitle: "Expected",
								CursorCol:   -1,
								CursorRow:   -1,
								Style: sumi.Style{
									FG: sumi.Color{Name: "blue"},
								},
							},
						},
					},
					box0,
					{
						Kind:      sumi.KindBox,
						Direction: "row",
						FlexGrow:  1,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								FlexGrow:    1,
								Border:      "single",
								BorderTitle: pvSourceTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
								Style: sumi.Style{
									FG: sumi.Color{Name: "magenta"},
								},
							},
							{
								Kind:        sumi.KindBox,
								FlexGrow:    1,
								Border:      "single",
								BorderTitle: pvSnapshotTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
								Style: sumi.Style{
									FG: sumi.Color{Name: "magenta"},
								},
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						FlexGrow:    1,
						Border:      "single",
						BorderTitle: pvScenarioTitle(),
						CursorCol:   -1,
						CursorRow:   -1,
						Style: sumi.Style{
							FG: sumi.Color{Name: "magenta"},
						},
					},
				},
			},
		},
	}

	sumi.Effect(func() {
		box0.Children = func() []*sumi.Input {
			var cs []*sumi.Input
			if interactive.Get() {
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Content: "esc",
					Style: sumi.Style{
						Inverse: true,
					},
				})
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Content: " Exit  ",
					Style: sumi.Style{
						Dim: true,
					},
				})
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Content: " INTERACTIVE ",
					Style: sumi.Style{
						FG:   sumi.Color{Name: "green"},
						Bold: true,
					},
				})
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Content: sumi.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current.Get()+1, pvStepCount(), pvStepName(current.Get())),
					Style: sumi.Style{
						Bold: true,
					},
				})
			} else {
				if pvIsEditorFocused() {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: sumi.Sprintf(" %v ", pvFocusName()),
						Style: sumi.Style{
							FG:   sumi.Color{Name: "cyan"},
							Bold: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "C-\\",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Exit  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "C-\\ h",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Prev  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "C-\\ l",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Next  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
				} else {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "h",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Prev  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "l",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Next  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "u",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Update  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "i",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Interactive  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "1",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Source  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "2",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Snap  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "3",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Scenario  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: "q",
						Style: sumi.Style{
							Inverse: true,
						},
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " Quit  ",
						Style: sumi.Style{
							Dim: true,
						},
					})
				}
				if matchStatus.Get() == 1 {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: " MATCH ",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "green"},
							Bold: true,
						},
					})
				} else {
					if matchStatus.Get() == 0 {
						cs = append(cs, &sumi.Input{
							Kind:    sumi.KindText,
							Content: " NO SNAPSHOT ",
							Style: sumi.Style{
								FG:   sumi.Color{Name: "yellow"},
								Bold: true,
							},
						})
					} else {
						cs = append(cs, &sumi.Input{
							Kind:    sumi.KindText,
							Content: " DIFF ",
							Style: sumi.Style{
								FG:   sumi.Color{Name: "red"},
								Bold: true,
							},
						})
					}
				}
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Content: sumi.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current.Get()+1, pvStepCount(), pvStepName(current.Get())),
					Style: sumi.Style{
						Bold: true,
					},
				})
			}
			return cs
		}()
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
