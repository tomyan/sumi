package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	confirming := sumi.New(false)
	status := sumi.New("untouched")

	openDialog := func() {
		confirming.Set(true)
	}

	confirmYes := func() {
		status.Set("deleted")
		confirming.Set(false)
	}

	confirmNo := func() {
		status.Set("kept")
		confirming.Set(false)
	}

	dialogClosed := func(evt *sumi.DOMEvent) {
		confirming.Set(false)
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
		Content: sumi.Sprintf("Status: %v", status.Get()),
	}
	box0 := &sumi.Input{
		Kind:  sumi.KindBox,
		Tag:   "dialog",
		Attrs: map[string]string{"onclose": "{dialogClosed}", "open": "{confirming.Get()}"},
		On: map[string]func(*sumi.DOMEvent){
			"close": dialogClosed,
		},
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			{
				Kind:    sumi.KindText,
				Tag:     "span",
				Content: "Really delete everything?",
			},
			{
				Kind:  sumi.KindBox,
				Tag:   "button",
				Attrs: map[string]string{"onclick": "{confirmYes}"},
				On: map[string]func(*sumi.DOMEvent){
					"click": func(evt *sumi.DOMEvent) {
						if h := (confirmYes); h != nil {
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
						Content: "Yes",
					},
				},
			},
			{
				Kind:  sumi.KindBox,
				Tag:   "button",
				Attrs: map[string]string{"onclick": "{confirmNo}"},
				On: map[string]func(*sumi.DOMEvent){
					"click": func(evt *sumi.DOMEvent) {
						if h := (confirmNo); h != nil {
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
						Content: "No",
					},
				},
			},
		},
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
						Content: "Enter opens the dialog; q quits",
					},
					{
						Kind:  sumi.KindBox,
						Tag:   "button",
						Attrs: map[string]string{"onclick": "{openDialog}"},
						On: map[string]func(*sumi.DOMEvent){
							"click": func(evt *sumi.DOMEvent) {
								if h := (openDialog); h != nil {
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
								Content: "Delete all",
							},
						},
					},
					node0,
					box0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Status: %v", status.Get())
		box0.Attrs["open"] = sumi.AttrString(confirming.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet("dialog {\n\tborder: single;\n\tleft: 4;\n\tpadding: 0 1;\n\tposition: absolute;\n\ttop: 2;\n\tz-index: 10;\n}\nbutton {\n\twidth: 14;\n}\nbutton:focus {\n\tcolor: yellow;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
