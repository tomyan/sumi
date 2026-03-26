package greeting

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type GreetingProps struct {
	Name string
}

func NewGreeting(props GreetingProps) *tui.Component {
	name := props.Name
	if name == "" {
		name = "World"
	}

	count := signal.New(0)

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if evt.Rune == 'q' {
			tui.Quit()
			return
		}
		if evt.Kind == input.EventKey {
			count.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Hello, %v!", name),
	}
	node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Keys pressed: %v", count.Get()),
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
					node1,
				},
			},
		},
	}

	signal.Effect(func() {
		node0.Content = fmt.Sprintf("Hello, %v!", name)
		node1.Content = fmt.Sprintf("Keys pressed: %v", count.Get())
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
