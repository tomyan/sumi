package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type CounterProps struct {
}

func NewCounter(props CounterProps) *sumi.Component {
	count := sumi.New(0)
	doubled := sumi.From(func() int { return count.Get() * 2 })

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
		Tag:     "text",
		Classes: []string{"count"},
		Attrs:   map[string]string{"class": "count"},
		Content: sumi.Sprintf("Count: %v (doubled: %v)", count.Get(), doubled.Get()),
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
				Tag:       "box",
				Classes:   []string{"container"},
				Attrs:     map[string]string{"class": "container", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"title"},
						Attrs:   map[string]string{"class": "title"},
						Content: "Sumi Counter",
					},
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press any key to increment, q to quit",
					},
					node0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Count: %v (doubled: %v)", count.Get(), doubled.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".container {\n\tborder: single;\n\tpadding: 1 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.count {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n"),
	}
}
