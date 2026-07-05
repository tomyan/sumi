package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type CounterProps struct {
}

func NewCounter(props CounterProps) *sumi.Component {
	count := sumi.New(0)

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
			count.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Content: sumi.Sprintf("Count: %v", count.Get()),
		Style: sumi.Style{
			FG:   sumi.Color{Name: "yellow"},
			Bold: true,
		},
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
						Content: "Signal Counter",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "green"},
							Bold: true,
						},
					},
					{
						Kind:    sumi.KindText,
						Content: "Press any key to increment, q to quit",
					},
					node0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Count: %v", count.Get())
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
