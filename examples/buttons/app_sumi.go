package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	saves := sumi.New(0)
	cancels := sumi.New(0)

	save := func() {
		saves.Update(func(n int) int { return n + 1 })
	}

	cancel := func() {
		cancels.Update(func(n int) int { return n + 1 })
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
		if evt.Rune == 'q' {
			sumi.Quit()
			return
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("Saved %v times, cancelled %v times", saves.Get(), cancels.Get()),
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
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Tab to move, Enter or click to press; q quits",
					},
					{
						Kind:  sumi.KindBox,
						Tag:   "button",
						Attrs: map[string]string{"onclick": "{save}"},
						On: map[string]func(*sumi.DOMEvent){
							"click": func(evt *sumi.DOMEvent) {
								if h := (save); h != nil {
									h()
								}
							},
						},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "Save",
							},
						},
					},
					{
						Kind:  sumi.KindBox,
						Tag:   "button",
						Attrs: map[string]string{"onclick": "{cancel}"},
						On: map[string]func(*sumi.DOMEvent){
							"click": func(evt *sumi.DOMEvent) {
								if h := (cancel); h != nil {
									h()
								}
							},
						},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "Cancel",
							},
						},
					},
					node0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Saved %v times, cancelled %v times", saves.Get(), cancels.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet("button {\n\tborder: single;\n\twidth: 14;\n}\nbutton:focus {\n\tborder-color: cyan;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
