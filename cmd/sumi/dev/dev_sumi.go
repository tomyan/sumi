package dev

import (
	sumi "github.com/tomyan/sumi/runtime/prelude"
)

type DevProps struct {
}

func NewDev(props DevProps) *sumi.Component {
	status := sumi.New("sumi dev — starting")
	barClass := sumi.New("bar ok")
	devBind(status, barClass)

	handleKey := func(evt sumi.Event) {
		devForward(evt)
	}

	regionResize := func(evt *sumi.DOMEvent) {
		devRegionResize(evt)
	}

	node0 := &sumi.Input{
		Kind:    sumi.KindText,
		Tag:     "text",
		Content: sumi.Sprintf("%v", status.Get()),
	}
	box0 := &sumi.Input{
		Kind:      sumi.KindBox,
		Tag:       "div",
		Classes:   []string{"{barClass.Get()}"},
		Attrs:     map[string]string{"class": "{barClass.Get()}"},
		CursorCol: -1,
		CursorRow: -1,
		Children: []*sumi.Input{
			node0,
		},
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
				Classes:   []string{"dev"},
				Attrs:     map[string]string{"class": "dev", "onkey": "handleKey"},
				CursorCol: -1,
				CursorRow: -1,
				Children: []*sumi.Input{
					{
						Kind:  sumi.KindBox,
						Tag:   "region",
						Attrs: map[string]string{"onresize": "{regionResize}"},
						On: map[string]func(*sumi.DOMEvent){
							"resize": regionResize,
						},
						CursorCol: -1,
						CursorRow: -1,
					},
					box0,
				},
			},
		},
	}

	sumi.Effect(func() {
		node0.Content = sumi.Sprintf("%v", status.Get())
		box0.Classes = sumi.SplitClasses(barClass.Get())
		box0.Attrs["class"] = barClass.Get()
	})

	return &sumi.Component{
		Tree:       root,
		OnEvent:    handleKey,
		Stylesheet: sumi.MustParseStylesheet(".dev {\n\tdisplay: flex;\n\tflex-direction: column;\n\theight: 100%;\n}\nregion {\n\tflex-grow: 1;\n}\n.bar {\n\tcolor: white;\n}\n.bar.ok {\n\tbackground: #1c3a2a;\n}\n.bar.err {\n\tbackground: #6b1a1a;\n}\n"),
	}
}
