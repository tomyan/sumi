package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {
	firstFocused := sumi.New(false)
	secondFocused := sumi.New(false)
	firstText := sumi.New("")
	secondText := sumi.New("")
	rootKeys := sumi.New(0)

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

	firstFocus := func(evt *sumi.DOMEvent) {
		firstFocused.Set(true)
	}

	firstBlur := func(evt *sumi.DOMEvent) {
		firstFocused.Set(false)
	}

	firstKey := func(evt *sumi.DOMEvent) {
		if evt.Key.Kind == sumi.EventKey {
			firstText.Update(func(s string) string { return s + string(evt.Key.Rune) })
			evt.StopPropagation()
		}
	}

	secondFocus := func(evt *sumi.DOMEvent) {
		secondFocused.Set(true)
	}

	secondBlur := func(evt *sumi.DOMEvent) {
		secondFocused.Set(false)
	}

	secondKey := func(evt *sumi.DOMEvent) {
		if evt.Key.Kind == sumi.EventKey {
			secondText.Update(func(s string) string { return s + string(evt.Key.Rune) })
			evt.StopPropagation()
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
		if evt.Kind == sumi.EventKey {
			rootKeys.Update(func(n int) int { return n + 1 })
		}
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("First (%v): %v", firstStatus(), firstText.Get()),
	}
	node1 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("Second (%v): %v", secondStatus(), secondText.Get()),
	}
	node2 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "span",
		Content: sumi.Sprintf("Root saw %v unconsumed keys", rootKeys.Get()),
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
						Content: "Tab / Shift+Tab moves focus; type into the focused field",
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"field"},
						Attrs:     map[string]string{"class": "field", "focusable": "true", "onblur": "{firstBlur}", "onfocus": "{firstFocus}", "onkeydown": "{firstKey}"},
						Focusable: true,
						On: map[string]func(*sumi.DOMEvent){
							"blur":    firstBlur,
							"focus":   firstFocus,
							"keydown": firstKey,
						},
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
						Attrs:     map[string]string{"class": "field", "focusable": "true", "onblur": "{secondBlur}", "onfocus": "{secondFocus}", "onkeydown": "{secondKey}"},
						Focusable: true,
						On: map[string]func(*sumi.DOMEvent){
							"blur":    secondBlur,
							"focus":   secondFocus,
							"keydown": secondKey,
						},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							node1,
						},
					},
					node2,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("First (%v): %v", firstStatus(), firstText.Get())
		node1.Content = sumi.Sprintf("Second (%v): %v", secondStatus(), secondText.Get())
		node2.Content = sumi.Sprintf("Root saw %v unconsumed keys", rootKeys.Get())
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".field {\n\tborder: single;\n\tpadding: 0 1;\n}\n.field:focus {\n\tborder-color: cyan;\n}\n.hint {\n\topacity: dim;\n}\n"),
	}
}
