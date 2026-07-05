package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	firstFocused := sumi.New(false)
	secondFocused := sumi.New(false)

	firstStatus := func() string {
		if firstFocused.Get() {
			return "focused"
		}
		return "blurred"
	}

	secondStatus := func() string {
		if secondFocused.Get() {
			return "focused"
		}
		return "blurred"
	}

	handleFirst := func(evt sumi.Event) {
		if evt.Kind == sumi.EventFocus {
			firstFocused.Set(true)
			return
		}
		if evt.Kind == sumi.EventBlur {
			firstFocused.Set(false)
			return
		}
	}

	handleSecond := func(evt sumi.Event) {
		if evt.Kind == sumi.EventFocus {
			secondFocused.Set(true)
			return
		}
		if evt.Kind == sumi.EventBlur {
			secondFocused.Set(false)
			return
		}
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
		Content: sumi.Sprintf("First field: %v", firstStatus()),
	}
	node1 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("Second field: %v", secondStatus()),
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
						Content: "Tab / Shift+Tab moves focus; q quits",
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"field"},
						Attrs:     map[string]string{"class": "field", "focusable": "true", "onkey": "handleFirst"},
						Focusable: true,
						OnKey:     handleFirst,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node0,
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"field"},
						Attrs:     map[string]string{"class": "field", "focusable": "true", "onkey": "handleSecond"},
						Focusable: true,
						OnKey:     handleSecond,
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node1,
						},
					},
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("First field: %v", firstStatus())
		node1.Content = sumi.Sprintf("Second field: %v", secondStatus())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".field {\n\tborder: single;\n\tpadding: 0 1;\n}\n.field:focus {\n\tborder-color: cyan;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
