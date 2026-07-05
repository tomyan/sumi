package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	notify := sumi.New(false)
	size := sumi.New("small")

	notifyChanged := func(evt *sumi.DOMEvent) {
		notify.Set(evt.Data["checked"].(bool))
	}

	sizeChanged := func(evt *sumi.DOMEvent) {
		size.Set(evt.Data["value"].(string))
	}

	notifyLabel := func() string {
		if notify.Get() {
			return "on"
		}
		return "off"
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
		Content: sumi.Sprintf("Notifications (%v)", notifyLabel()),
	}
	node1 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("Size: %v", size.Get()),
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
						Content: "Tab to move, Space to toggle; q quits",
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"row"},
						Attrs:     map[string]string{"class": "row"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:  sumi.KindBox,
								Tag:   "input",
								Attrs: map[string]string{"onchange": "{notifyChanged}", "type": "checkbox"},
								On: map[string]func(*sumi.DOMEvent){
									"change": notifyChanged,
								},
								CursorCol: -1,
								CursorRow: -1,
							},
							node0,
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"row"},
						Attrs:     map[string]string{"class": "row"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:  sumi.KindBox,
								Tag:   "input",
								Attrs: map[string]string{"checked": "true", "name": "size", "onchange": "{sizeChanged}", "type": "radio", "value": "small"},
								On: map[string]func(*sumi.DOMEvent){
									"change": sizeChanged,
								},
								CursorCol: -1,
								CursorRow: -1,
							},
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Content: "Small",
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"row"},
						Attrs:     map[string]string{"class": "row"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:  sumi.KindBox,
								Tag:   "input",
								Attrs: map[string]string{"name": "size", "onchange": "{sizeChanged}", "type": "radio", "value": "large"},
								On: map[string]func(*sumi.DOMEvent){
									"change": sizeChanged,
								},
								CursorCol: -1,
								CursorRow: -1,
							},
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Content: "Large",
							},
						},
					},
					node1,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Notifications (%v)", notifyLabel())
		node1.Content = sumi.Sprintf("Size: %v", size.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".row {\n\tflex-direction: row;\n\tgap: 1;\n}\ninput:focus {\n\tcolor: yellow;\n}\ninput:checked {\n\tcolor: green;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
