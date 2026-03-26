package main

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

type TodoProps struct {
}

func NewTodo(props TodoProps) *tui.Component {
	items := signal.New([]string{"Buy groceries", "Write tests", "Review PR"})
	selected := signal.New(0)

	handleKey := func(evt input.Event) {
		if evt.Kind == input.EventSignal {
			tui.Quit()
			return
		}
		if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') {
			tui.Quit()
			return
		}
		if evt.Kind == input.EventSpecial && evt.Special == input.KeyDown {
			n := len(items.Get())
			if n > 0 {
				selected.Set((selected.Get() + 1) % n)
			}
		}
		if evt.Kind == input.EventSpecial && evt.Special == input.KeyUp {
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

	box0 := &layout.Input{
		Kind:      layout.KindBox,
		Padding:   layout.ParsePadding("1 2"),
		Border:    "single",
		CursorCol: -1,
		CursorRow: -1,
		Style: render.Style{
			FG: render.Color{Name: "cyan"},
		},
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children: []*layout.Input{
			box0,
		},
	}

	signal.Effect(func() {
		box0.Children = func() []*layout.Input {
			var cs []*layout.Input
			cs = append(cs, &layout.Input{
				Kind:    layout.KindText,
				Content: "Todo List",
				Style: render.Style{
					FG:   render.Color{Name: "green"},
					Bold: true,
				},
			})
			cs = append(cs, &layout.Input{
				Kind:    layout.KindText,
				Content: "↑↓ navigate, d delete, q quit",
				Style: render.Style{
					Dim: true,
				},
			})
			for i, item := range items.Get() {
				if i == selected.Get() {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: fmt.Sprintf("▶ %v", item),
						Style: render.Style{
							FG:   render.Color{Name: "cyan"},
							Bold: true,
						},
					})
				} else {
					cs = append(cs, &layout.Input{
						Kind:    layout.KindText,
						Content: fmt.Sprintf("  %v", item),
					})
				}
			}
			return cs
		}()
	})

	return &tui.Component{
		Tree:    root,
		OnEvent: handleKey,
	}
}
