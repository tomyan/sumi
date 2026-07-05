package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	name := sumi.New("")

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			sumi.Quit()
			return
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Content: sumi.Sprintf("You typed: %v", name.Get()),
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Content: "Text Input Demo",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Type to enter your name",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Name:",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "yellow"},
							Bold: true,
						},
					},
					node0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("You typed: %v", name.Get())
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
