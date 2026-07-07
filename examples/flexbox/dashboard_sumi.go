package main

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type DashboardProps struct {
}

func NewDashboard(props DashboardProps) *sumi.Component {

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
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"header"},
						Attrs:     map[string]string{"class": "header"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Classes: []string{"title"},
								Attrs:   map[string]string{"class": "title"},
								Content: "Sumi Flexbox Dashboard",
							},
						},
					},
					{
						Kind:    sumi.KindText,
						Tag:     "span",
						Classes: []string{"hint"},
						Attrs:   map[string]string{"class": "hint"},
						Content: "Press q to quit",
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"panels"},
						Attrs:     map[string]string{"class": "panels"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:      sumi.KindBox,
								Tag:       "div",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Left Panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Content: "This panel uses flex-grow",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Content: "to fill available space.",
									},
								},
							},
							{
								Kind:      sumi.KindBox,
								Tag:       "div",
								Classes:   []string{"panel"},
								Attrs:     map[string]string{"class": "panel"},
								CursorCol: -1,
								CursorRow: -1,
								Children: []*sumi.Input{
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Classes: []string{"panel-title"},
										Attrs:   map[string]string{"class": "panel-title"},
										Content: "Right Panel",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Content: "Both panels share the",
									},
									{
										Kind:    sumi.KindText,
										Tag:     "span",
										Content: "width equally.",
									},
								},
							},
						},
					},
					{
						Kind:      sumi.KindBox,
						Tag:       "div",
						Classes:   []string{"footer"},
						Attrs:     map[string]string{"class": "footer"},
						CursorCol: -1,
						CursorRow: -1,
						Children: []*sumi.Input{
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Classes: []string{"status"},
								Attrs:   map[string]string{"class": "status"},
								Content: "Ready",
							},
							{
								Kind:    sumi.KindText,
								Tag:     "span",
								Classes: []string{"version"},
								Attrs:   map[string]string{"class": "version"},
								Content: "sumi v0.1",
							},
						},
					},
				},
			},
		},
	}

	return &sumi.Component{
		Tree:       root,
		Stylesheet: sumi.MustParseStylesheet(".header {\n\tborder: single;\n\tdisplay: flex;\n\tjustify-content: center;\n\tpadding: 0 2;\n}\n.title {\n\tcolor: green;\n\tfont-weight: bold;\n}\n.hint {\n\tcolor: cyan;\n\topacity: dim;\n}\n.panels {\n\tdisplay: flex;\n\tflex-direction: row;\n\tgap: 1;\n}\n.panel {\n\tborder: single;\n\tflex-grow: 1;\n\tpadding: 0 1;\n}\n.panel-title {\n\tcolor: yellow;\n\tfont-weight: bold;\n}\n.footer {\n\tborder: single;\n\tdisplay: flex;\n\tflex-direction: row;\n\tjustify-content: space-between;\n\tpadding: 0 2;\n}\n.status {\n\tcolor: green;\n}\n.version {\n\tcolor: cyan;\n\topacity: dim;\n}\n"),
	}
}
