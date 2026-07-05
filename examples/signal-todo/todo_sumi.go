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
		Padding:   sumi.ParsePadding("1 2"),
		Border:    "single",
		CursorCol: -1,
		CursorRow: -1,
		Style: sumi.Style{
			FG: sumi.Color{Name: "cyan"},
		},
	}
	root := &sumi.Input{
		Kind:      sumi.KindBox,
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
				Content: "Todo List",
				Style: sumi.Style{
					FG:   sumi.Color{Name: "green"},
					Bold: true,
				},
			})
			cs = append(cs, &sumi.Input{
				Kind:    sumi.KindText,
				Content: "↑↓ navigate, d delete, q quit",
				Style: sumi.Style{
					Dim: true,
				},
			})
			for i, item := range items.Get() {
				if i == selected.Get() {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: sumi.Sprintf("▶ %v", item),
						Style: sumi.Style{
							FG:   sumi.Color{Name: "cyan"},
							Bold: true,
						},
					})
				} else {
					cs = append(cs, &sumi.Input{
						Kind:    sumi.KindText,
						Content: sumi.Sprintf("  %v", item),
					})
				}
			}
			return cs
		}()
	})

	return &sumi.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
