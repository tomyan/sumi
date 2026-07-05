package counter

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type CounterProps struct {
	Label string
}

func NewCounter(props CounterProps) *sumi.Component {
	label := props.Label
	if label == "" {
		label = "Count"
	}

	count := sumi.New(0)

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventKey {
			count.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Content: sumi.Sprintf("%v:", label),
		Style: sumi.Style{
			FG:   sumi.Color{Name: "cyan"},
			Bold: true,
		},
	}
	node1 := &sumi.Input{
		Kind:    sumi.KindText,
		Content: sumi.Sprintf("%v", count.Get()),
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
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					node0,
					node1,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("%v:", label)
		node1.Content = sumi.Sprintf("%v", count.Get())
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
