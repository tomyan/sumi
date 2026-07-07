package splitpanel

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {

	handleKey := func(evt sumi.Event) {
		if evt.Kind == sumi.EventSignal {
			sumi.Quit()
			return
		}
		if evt.Ctrl && evt.Rune == 'c' {
			sumi.Quit()
			return
		}
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
				Classes:   []string{"root"},
				Attrs:     map[string]string{"class": "root", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:        sumi.KindBox,
						Tag:         "div",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Actual", "class": "panel", "height": "5"},
						FixedHeight: 5,
						BorderTitle: "Actual",
						CursorCol:   -1,
						CursorRow:   -1,
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "div",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Expected", "class": "panel", "height": "5"},
						FixedHeight: 5,
						BorderTitle: "Expected",
						CursorCol:   -1,
						CursorRow:   -1,
					},
				},
			},
		},
	}

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".root {\n\tborder-collapse: collapse;\n\tdisplay: flex;\n\tflex-direction: row;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n"),
	}
}
