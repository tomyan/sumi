package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type AppProps struct {
}

func NewApp(props AppProps) *sumi.Component {

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
				Attrs:     map[string]string{"border": "single", "padding": "1 2"},
				Padding:   sumi.ParsePadding("1 2"),
				Border:    "single",
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "Welcome to Sumi!",
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "text",
								Content: "A declarative TTY framework for Go.",
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Attrs:     map[string]string{"border": "single", "padding": "0 1"},
						Padding:   sumi.ParsePadding("0 1"),
						Border:    "single",
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								Tag:       "div",
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "text",
										Content: "Press Enter to exit.",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &sumi.Component{
		Tree: root,
	}
}
