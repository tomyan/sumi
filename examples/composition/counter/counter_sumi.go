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
		Tag:     "text",
		Content: sumi.Sprintf("%v:", label),
	}
	node1 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "text",
		Content: sumi.Sprintf("%v", count.Get()),
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
				Tag:       "div",
				Attrs:     map[string]string{"onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"label"},
						Attrs:     map[string]string{"class": "label"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node0,
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"count"},
						Attrs:     map[string]string{"class": "count"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node1,
						},
					},
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("%v:", label)
		node1.Content = sumi.Sprintf("%v", count.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".label {\n\tcolor: cyan;\n\tfont-weight: bold;\n}\n.count {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n"),
	}
}
