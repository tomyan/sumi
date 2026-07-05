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
		Tag:       "box",
		Attrs:     map[string]string{"flex-direction": "row"},
		Direction: "row",
		CursorCol: -1,
		CursorRow: -1,
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "root",
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				Tag:       "box",
				Attrs:     map[string]string{"onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"panels"},
						Attrs:     map[string]string{"class": "panels"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"actual"},
								Attrs:       map[string]string{"border-title": "Actual", "class": "actual", "height": "{pvComponentHeight()}"},
								FixedHeight: pvComponentHeight(),
								BorderTitle: "Actual",
								CursorCol:   -1,
								CursorRow:   -1,
							},
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"expected"},
								Attrs:       map[string]string{"border-title": "Expected", "class": "expected", "height": "{pvComponentHeight()}"},
								FixedHeight: pvComponentHeight(),
								BorderTitle: "Expected",
								CursorCol:   -1,
								CursorRow:   -1,
							},
						},
					},
					box0,
					{
						Kind:      sumi.KindBox,
						Tag:       "box",
						Classes:   []string{"editors"},
						Attrs:     map[string]string{"class": "editors"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"editor-left"},
								Attrs:       map[string]string{"border-title": "{pvSourceTitle()}", "class": "editor-left"},
								BorderTitle: pvSourceTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
							},
							{
								Kind:        sumi.KindBox,
								Tag:         "box",
								Classes:     []string{"editor-right"},
								Attrs:       map[string]string{"border-title": "{pvSnapshotTitle()}", "class": "editor-right"},
								BorderTitle: pvSnapshotTitle(),
								CursorCol:   -1,
								CursorRow:   -1,
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "box",
						Classes:     []string{"scenario-editor"},
						Attrs:       map[string]string{"border-title": "{pvScenarioTitle()}", "class": "scenario-editor"},
						BorderTitle: pvScenarioTitle(),
						CursorCol:   -1,
						CursorRow:   -1,
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
					Tag:     "text",
					Classes: []string{"key"},
					Attrs:   map[string]string{"class": "key"},
					Content: "esc",
				})
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Tag:     "text",
					Classes: []string{"label"},
					Attrs:   map[string]string{"class": "label"},
					Content: " Exit  ",
				})
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Tag:     "text",
					Classes: []string{"interactive"},
					Attrs:   map[string]string{"class": "interactive"},
					Content: " INTERACTIVE ",
				})
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Tag:     "text",
					Classes: []string{"info"},
					Attrs:   map[string]string{"class": "info"},
					Content: sumi.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current.Get()+1, pvStepCount(), pvStepName(current.Get())),
				})
			} else {
				if pvIsEditorFocused() {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"focus-indicator"},
						Attrs:   map[string]string{"class": "focus-indicator"},
						Content: sumi.Sprintf(" %v ", pvFocusName()),
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "C-\\",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Exit  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "C-\\ h",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Prev  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "C-\\ l",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Next  ",
					})
				} else {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "h",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Prev  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "l",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Next  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "u",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Update  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "i",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Interactive  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "1",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Source  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "2",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Snap  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "3",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Scenario  ",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"key"},
						Attrs:   map[string]string{"class": "key"},
						Content: "q",
					})
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: " Quit  ",
					})
				}
				if matchStatus.Get() == 1 {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"match"},
						Attrs:   map[string]string{"class": "match"},
						Content: " MATCH ",
					})
				} else {
					if matchStatus.Get() == 0 {
						cs = append(cs, &sumi.Input{
							Kind:    sumi.KindText,
							Tag:     "text",
							Classes: []string{"no-snap"},
							Attrs:   map[string]string{"class": "no-snap"},
							Content: " NO SNAPSHOT ",
						})
					} else {
						cs = append(cs, &sumi.Input{
							Kind:    sumi.KindText,
							Tag:     "text",
							Classes: []string{"diff"},
							Attrs:   map[string]string{"class": "diff"},
							Content: " DIFF ",
						})
					}
				}
				cs = append(cs, &sumi.Input{
					Kind:    sumi.KindText,
					Tag:     "text",
					Classes: []string{"info"},
					Attrs:   map[string]string{"class": "info"},
					Content: sumi.Sprintf("  %v  Frame %v/%v  %v", pvScenarioName(), current.Get()+1, pvStepCount(), pvStepName(current.Get())),
				})
			}
			return cs
		}()
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".panels {\n\tflex-direction: row;\n}\n.actual {\n\tborder: single;\n\tborder-color: cyan;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.expected {\n\tborder: single;\n\tborder-color: blue;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.editors {\n\tflex-direction: row;\n\tflex-grow: 1;\n}\n.editor-left {\n\tborder: single;\n\tborder-color: magenta;\n\tflex-grow: 1;\n}\n.editor-right {\n\tborder: single;\n\tborder-color: magenta;\n\tflex-grow: 1;\n}\n.scenario-editor {\n\tborder: single;\n\tborder-color: magenta;\n\tflex-grow: 1;\n}\n.match {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.diff {\n\tcolor: red;\n\tfont-weight: bold;\n}\n.no-snap {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n.info {\n\tfont-weight: bold;\n}\n.key {\n\tinverse: true;\n}\n.label {\n\topacity: dim;\n}\n.interactive {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.focus-indicator {\n\tcolor: cyan;\n\tfont-weight: bold;\n}\n"),
	}
}
