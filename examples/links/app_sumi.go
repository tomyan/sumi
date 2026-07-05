package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	lastOpened := sumi.New("none")

	docsClicked := func() {
		lastOpened.Set("docs")
	}

	blogClicked := func() {
		lastOpened.Set("blog")
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
		Content: sumi.Sprintf("Last activated: %v", lastOpened.Get()),
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
						Content: "Tab to move, Enter to open; q quits",
					},
					{
						Kind:    sumi.KindText,
						Tag:     "a",
						Attrs:   map[string]string{"href": "https://example.com/docs", "onclick": "{docsClicked}"},
						Content: "Documentation",
						On: map[string]func(*sumi.DOMEvent){
							"click": func(evt *sumi.DOMEvent) {
								if h := (docsClicked); h != nil {
									h()
								}
							},
						},
					},
					{
						Kind:    sumi.KindText,
						Tag:     "a",
						Attrs:   map[string]string{"href": "https://example.com/blog", "onclick": "{blogClicked}"},
						Content: "Blog",
						On: map[string]func(*sumi.DOMEvent){
							"click": func(evt *sumi.DOMEvent) {
								if h := (blogClicked); h != nil {
									h()
								}
							},
						},
					},
					node0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("Last activated: %v", lastOpened.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet("a {\n\tdisplay: block;\n}\na:focus {\n\tinverse: true;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
