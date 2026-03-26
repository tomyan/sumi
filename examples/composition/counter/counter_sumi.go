package counter

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type CounterProps struct {
	Label string
}

func NewCounter(props CounterProps) *tui.Component {
	label := props.Label
	if label == "" {
		label = "Count"
	}

	count := signal.New(0)

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventKey {
			count.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v:", label),
		Style: render.Style{
			FG:   render.Color{Name: "cyan"},
			Bold: true,
		},
	}
	node1 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("%v", count.Get()),
		Style: render.Style{
			FG:   render.Color{Name: "yellow"},
			Bold: true,
		},
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
		node0.Content = fmt.Sprintf("%v:", label)
		node1.Content = fmt.Sprintf("%v", count.Get())
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
