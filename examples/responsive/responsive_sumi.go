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
		Classes: []string{"dims"},
		Attrs:   map[string]string{"class": "dims"},
		Content: sumi.Sprintf("Terminal: %vx%v", width.Get(), height.Get()),
		Style: sumi.Style{
			FG:   sumi.Color{Name: "yellow"},
			Bold: true,
		},
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "root",
		Direction: "column",
		Overflow:  "auto",
		MinWidth:  48,
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:      sumi.KindBox,
				Tag:       "box",
				Classes:   []string{"header"},
				Attrs:     map[string]string{"class": "header", "onkey": "handleKey"},
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Style: sumi.Style{
					FG: sumi.Color{Name: "cyan"},
				},
				Children: []*sumi.Input{
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"title"},
						Attrs:   map[string]string{"class": "title"},
						Content: "Sumi Responsive Demo",
						Style: sumi.Style{
							FG:   sumi.Color{Name: "green"},
							Bold: true,
						},
					},
					node0,
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Resize your terminal to see this update! Press q to quit.",
						Style: sumi.Style{
							FG:  sumi.Color{Name: "cyan"},
							Dim: true,
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
