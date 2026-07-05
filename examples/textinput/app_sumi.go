package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	name := sumi.New("")

	nameChanged := func(evt *sumi.DOMEvent) {
		name.Set(evt.Data["value"].(string))
	}

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
		Tag:     "text",
		Content: sumi.Sprintf("Hello, %v!", name.Get()),
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
				Classes:   []string{"form"},
				Attrs:     map[string]string{"class": "form", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"hint"},
						Attrs:     map[string]string{"class": "hint"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "Type your name; Ctrl+C quits",
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "Name:",
							},
						},
					},
					{
						Kind:  sumi.KindBox,
						Tag:   "input",
						Attrs: map[string]string{"oninput": "{nameChanged}", "value": ""},
						On: map[string]func(*sumi.DOMEvent){
							"input": nameChanged,
						},
						CursorCol: -1,
						CursorRow: -1,
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node0,
						},
					},
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Hello, %v!", name.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".form {\n\tpadding: 1 2;\n}\ninput {\n\tbackground: #333333;\n}\ninput:focus {\n\tbackground: #444444;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
