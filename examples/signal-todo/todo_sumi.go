package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type TodoProps struct {
}

func NewTodo(props TodoProps) *sumi.Component {
	items := sumi.New([]string{"Buy groceries", "Write tests", "Review PR"})
	selected := sumi.New(0)

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			sumi.Quit()
			return
		}
		if evt.Kind == sumi.EventSpecial && evt.Special == sumi.KeyDown {
			n := len(items.Get())
			if n > 0 {
				selected.Set((selected.Get() + 1) % n)
			}
		}
		if evt.Kind == sumi.EventSpecial && evt.Special == sumi.KeyUp {
			n := len(items.Get())
			if n > 0 {
				selected.Set((selected.Get() - 1 + n) % n)
			}
		}
		if evt.Rune == 'd' {
			idx := selected.Get()
			its := items.Get()
			if idx < len(its) {
				items.Set(append(its[:idx], its[idx+1:]...))
				if selected.Get() >= len(items.Get()) && selected.Get() > 0 {
					selected.Set(selected.Get() - 1)
				}
			}
		}
	}

	box0 := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "div",
		Classes:   []string{"container"},
		Attrs:     map[string]string{"class": "container", "onkey": "handleKey"},
		CursorCol: -1,
		CursorRow: -1,
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "root",
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			box0,
		},
	}

	sumi.Effect(func() {
		box0.Children = func() []*sumi.Input {
			var cs []*sumi.Input
			cs = append(cs, &sumi.Input{
				Kind:    sumi.KindText,
				Tag:     "span",
				Classes: []string{"title"},
				Attrs:   map[string]string{"class": "title"},
				Content: "Todo List",
			})
			cs = append(cs, &sumi.Input{
				Kind:    sumi.KindText,
				Tag:     "span",
				Classes: []string{"hint"},
				Attrs:   map[string]string{"class": "hint"},
				Content: "↑↓ navigate, d delete, q quit",
			})
			for i, item := range items.Get() {
				if i == selected.Get() {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"selected"},
						Attrs:   map[string]string{"class": "selected"},
						Content: sumi.Sprintf("▶ %v", item),
					})
				} else {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"item"},
						Attrs:   map[string]string{"class": "item"},
						Content: sumi.Sprintf("  %v", item),
					})
				}
			}
			return cs
		}()
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".container {\n\tborder: single;\n\tborder-color: cyan;\n\tpadding: 1 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\topacity: dim;\n}\n.selected {\n\tcolor: cyan;\n\tfont-weight: bold;\n}\n.item {\n}\n"),
	}
}
