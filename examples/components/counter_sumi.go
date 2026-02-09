package main

import (
	"fmt"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
)

type CounterComponent struct {
	label string
	count int
	dirty bool
}

func NewCounterComponent(label string) *CounterComponent {
	return &CounterComponent{
		label: label,
		count: 0,
		dirty: true,
	}
}

func (c *CounterComponent) Layout() *layout.Input {
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		Children: []*layout.Input{
			{
				Kind: layout.KindBox,
				Children: []*layout.Input{
					{
						Kind:    layout.KindText,
						Content: fmt.Sprintf("%v:", c.label),
						Style: render.Style{
							FG:   render.Color{Name: "cyan"},
							Bold: true,
						},
					},
					{
						Kind:    layout.KindText,
						Content: fmt.Sprintf("%v", c.count),
						Style: render.Style{
							FG:   render.Color{Name: "yellow"},
							Bold: true,
						},
					},
				},
			},
		},
	}
	return root
}

func (c *CounterComponent) HandleKey(key rune) {
	c.increment()
}

func (c *CounterComponent) Dirty() bool {
	d := c.dirty
	c.dirty = false
	return d
}

func (c *CounterComponent) increment() {
	c.count = c.count + 1
	c.dirty = true
}
