package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type PanelsProps struct {
}

func NewPanels(props PanelsProps) *sumi.Component {

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
				Classes:   []string{"layout"},
				Attrs:     map[string]string{"class": "layout"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"left-col"},
						Attrs:     map[string]string{"class": "left-col"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:        sumi.KindBox,
								Tag:         "div",
								Classes:     []string{"panel"},
								Attrs:       map[string]string{"border-title": "Panel 1", "class": "panel"},
								BorderTitle: "Panel 1",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Classes: []string{"title"},
										Attrs:   map[string]string{"class": "title"},
										Content: "Top-left panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Content: "Content goes here",
									},
								},
							},
							{
								Kind:        sumi.KindBox,
								Tag:         "div",
								Classes:     []string{"panel"},
								Attrs:       map[string]string{"border-title": "Panel 2", "class": "panel"},
								BorderTitle: "Panel 2",
								CursorCol:   -1,
								CursorRow:   -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Classes: []string{"title"},
										Attrs:   map[string]string{"class": "title"},
										Content: "Bottom-left panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Content: "More content here",
									},
								},
							},
						},
					},
					{
						Kind:        sumi.KindBox,
						Tag:         "div",
						Classes:     []string{"panel"},
						Attrs:       map[string]string{"border-title": "Panel 3", "class": "panel"},
						BorderTitle: "Panel 3",
						CursorCol:   -1,
						CursorRow:   -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
								Content: "Right panel",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Content: "This panel spans the full height",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Classes: []string{"hint"},
								Attrs:   map[string]string{"class": "hint"},
								Content: "Press q to quit",
							},
						},
					},
				},
			},
		},
	}

	return &sumi.Component{
		Tree:       root,
		Stylesheet: sumi.MustParseStylesheet(".layout {\n\tborder-collapse: collapse;\n\tdisplay: flex;\n\tflex-direction: row;\n}\n.left-col {\n\tborder: single;\n\tborder-collapse: collapse;\n\tflex-grow: 1;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n"),
	}
}
