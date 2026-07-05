package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type ResponsiveProps struct {
}

func NewResponsive(props ResponsiveProps) *sumi.Component {
	width := sumi.Env[int]("width")
	height := sumi.Env[int]("height")

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			sumi.Quit()
			return
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Content: sumi.Sprintf("Terminal: %vx%v", width.Get(), height.Get()),
		Style: sumi.Style{
			FG:   sumi.Color{Name: "yellow"},
			Bold: true,
		},
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Direction: "column",
		Overflow:  "auto",
		MinWidth:  48,
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Style: sumi.Style{
					FG: sumi.Color{Name: "cyan"},
				},
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Content: "Sumi Responsive Demo",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "green"},
							Bold: true,
						},
					},
					node0,
					{
						Kind:    sumi.KindText,
						Content: "Resize your terminal to see this update! Press q to quit.",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
						},
					},
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Terminal: %vx%v", width.Get(), height.Get())
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
