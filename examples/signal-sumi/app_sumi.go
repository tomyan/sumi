package main

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type AppProps struct {
}

func NewApp(props AppProps) *tui.Component {
	count := signal.New(0)

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			tui.Quit()
			return
		}
		if evt.Rune == 'q' {
			tui.Quit()
			return
		}
		if evt.Rune == '+' || evt.Rune == '=' {
			count.Update(func(n int) int { return n + 1 })
		}
		if evt.Rune == '-' {
			count.Update(func(n int) int { return n - 1 })
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Count: %v", count.Get()),
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
					node0,
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("Count: %v", count.Get())
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
