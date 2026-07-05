package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type ModalProps struct {
}

func NewModal(props ModalProps) *sumi.Component {
	showModal := sumi.New(false)

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			sumi.Quit()
			return
		}
		if evt.Kind == sumi.EventKey {
			showModal.Set(!showModal.Get())
		}
	}

	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
	}

	sumi.Effect(func() {
		root.Children = func() []*sumi.Input {
			var cs []*sumi.Input
			cs = append(cs, &sumi.Input{
				Kind:      sumi.KindBox,
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Content: "Modal Demo",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Press any key to toggle modal, q to quit",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Background content here",
					},
				},
			})
			if showModal.Get() {
				cs = append(cs, &sumi.Input{
					Kind:        sumi.KindBox,
					FixedWidth:  40,
					FixedHeight: 8,
					Padding:     sumi.ParsePadding("1 2"),
					Border:      "single",
					Position:    "fixed",
					Top:         5,
					Left:        10,
					ZIndex:      2,
					CursorCol:   -1,
					CursorRow:   -1,
					Style: sumi.Style{
						FG: sumi.Color{Name: "yellow"},
						BG: sumi.Color{Name: "black"},
					},
					Children: []*sumi.Input{
						{
							Kind:    sumi.KindText,
							Content: "Modal Dialog",
							Style: sumi.Style{
								FG:   sumi.Color{Name: "yellow"},
								Bold: true,
							},
						},
						{
							Kind:    sumi.KindText,
							Content: "This is a fixed-position modal overlay.",
						},
						{
							Kind:    sumi.KindText,
							Content: "Press any key to close",
							Style: sumi.Style{
								FG:  sumi.Color{Name: "cyan"},
								Dim: true,
							},
						},
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
