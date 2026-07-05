package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type ModalProps struct {
}

func NewModal(props ModalProps) *sumi.Component {
	showModal := sumi.New(false)

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
			showModal.Set(!showModal.Get())
		}
	}

	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "root",
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
	}

	sumi.Effect(func() {
		root.Children = func() []*sumi.Input {
			var cs []*sumi.Input
			cs = append(cs, &sumi.Input{
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
						Content: "Modal Demo",
					},
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press any key to toggle modal, q to quit",
					},
					{
						Kind:    sumi.KindText,
						Tag:     "text",
						Content: "Background content here",
					},
				},
			})
			if showModal.Get() {
				cs = append(cs, &sumi.Input{
					Kind:        sumi.KindBox,
					Tag:         "box",
					Classes:     []string{"modal"},
					Attrs:       map[string]string{"class": "modal", "height": "8", "left": "10", "position": "fixed", "top": "5", "width": "40", "z-index": "2"},
					FixedWidth:  40,
					FixedHeight: 8,
					Position:    "fixed",
					Top:         5,
					Left:        10,
					ZIndex:      2,
					CursorCol:   -1,
					CursorRow:   -1,
					Children: []*sumi.Input{
						{
							Kind:    sumi.KindText,
							Tag:     "text",
							Classes: []string{"modal-title"},
							Attrs:   map[string]string{"class": "modal-title"},
							Content: "Modal Dialog",
						},
						{
							Kind:    sumi.KindText,
							Tag:     "text",
							Content: "This is a fixed-position modal overlay.",
						},
						{
							Kind:    sumi.KindText,
							Tag:     "text",
							Classes: []string{"hint"},
							Attrs:   map[string]string{"class": "hint"},
							Content: "Press any key to close",
						},
					},
				})
			}
			return cs
		}()
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".container {\n\tborder: single;\n\tpadding: 1 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.modal {\n\tbackground: black;\n\tborder: single;\n\tborder-color: yellow;\n\tpadding: 1 2;\n}\n.modal-title {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n"),
	}
}
