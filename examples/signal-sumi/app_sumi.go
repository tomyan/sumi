package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	count := sumi.New(0)

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			sumi.Quit()
			return
		}
		if evt.Rune == 'q' {
			sumi.Quit()
			return
		}
		if evt.Rune == '+' || evt.Rune == '=' {
			count.Update(func(n int) int { return n + 1 })
		}
		if evt.Rune == '-' {
			count.Update(func(n int) int { return n - 1 })
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Content: sumi.Sprintf("Count: %v", count.Get()),
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
