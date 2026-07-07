package greeting

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type GreetingProps struct {
	Name string
}

func NewGreeting(props GreetingProps) *sumi.Component {
	name := props.Name
	if name == "" {
		name = "World"
	}

	count := sumi.New(0)

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Rune == 'q' {
			sumi.Quit()
			return
		}
		if evt.Kind == sumi.EventKey {
			count.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "text",
		Content: sumi.Sprintf("Hello, %v!", name),
	}
	node1 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "text",
		Content: sumi.Sprintf("Keys pressed: %v", count.Get()),
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
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node0,
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
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
		node0.Content = sumi.Sprintf("Hello, %v!", name)
		node1.Content = sumi.Sprintf("Keys pressed: %v", count.Get())
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
