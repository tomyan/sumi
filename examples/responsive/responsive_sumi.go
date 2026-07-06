package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type ResponsiveProps struct {
}

func NewResponsive(props ResponsiveProps) *sumi.Component {
	width := sumi.Env[int]("width")
	height := sumi.Env[int]("height")

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			sumi.Quit()
			return
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "text",
		Content: sumi.Sprintf("Terminal: %vx%v", width.Get(), height.Get()),
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
				Classes:   []string{"header"},
				Attrs:     map[string]string{"class": "header", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"title"},
						Attrs:     map[string]string{"class": "title"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "Sumi Responsive Demo",
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"dims"},
						Attrs:     map[string]string{"class": "dims"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node0,
						},
					},
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
								Content: "Resize your terminal to see this update! Press q to quit.",
							},
						},
					},
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Terminal: %vx%v", width.Get(), height.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet("root {\n\tmin-width: 48;\n\toverflow: auto;\n}\n.header {\n\tborder: single;\n\tborder-color: cyan;\n\tpadding: 1 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.dims {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n"),
	}
}
