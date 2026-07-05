package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	name := sumi.New("")

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			sumi.Quit()
			return
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("You typed: %v", name.Get()),
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
				Classes:   []string{"container"},
				Attrs:     map[string]string{"class": "container", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"title"},
						Attrs:   map[string]string{"class": "title"},
						Content: "Text Input Demo",
					},
					{
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Type to enter your name",
					},
					{
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"label"},
						Attrs:   map[string]string{"class": "label"},
						Content: "Name:",
					},
					node0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("You typed: %v", name.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".container {\n\tborder: single;\n\tpadding: 1 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.label {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n"),
	}
}
