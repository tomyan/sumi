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
		if evt.Kind == sumi.EventKey {
			n := len(items.Get())
			if n > 0 {
				selected.Set((selected.Get() + 1) % n)
			}
		}
	}

	box0 := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "box",
		Classes:   []string{"container"},
		Attrs:     map[string]string{"class": "container", "onkey": "handleKey"},
		Padding:   sumi.ParsePadding("1 2"),
		Border:    "single",
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
				Tag:     "text",
				Classes: []string{"title"},
				Attrs:   map[string]string{"class": "title"},
				Content: "Todo List",
				Style: sumi.Style{
					FG:   sumi.Color{Name: "green"},
					Bold: true,
				},
			})
			cs = append(cs, &sumi.Input{
				Kind:    sumi.KindText,
				Tag:     "text",
				Classes: []string{"hint"},
				Attrs:   map[string]string{"class": "hint"},
				Content: "Press any key to cycle, q to quit",
				Style: sumi.Style{
					FG:  sumi.Color{Name: "cyan"},
					Dim: true,
				},
			})
			for i, item := range items.Get() {
				if i == selected.Get() {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Classes: []string{"selected"},
						Attrs:   map[string]string{"class": "selected"},
						Content: sumi.Sprintf("> %v", item),
						Style: sumi.Style{
							FG:   sumi.Color{Name: "yellow"},
							Bold: true,
						},
					})
				} else {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Tag:     "text",
						Content: sumi.Sprintf("  %v", item),
					})
				}
				cs[len(cs)-1].Key = sumi.Sprint(item)
			}
			return cs
		}()
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".container {\n\tborder: single;\n\tpadding: 1 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.selected {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n"),
	}
}
