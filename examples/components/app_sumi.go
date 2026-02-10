package main

import (
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

func Run() {
	counter0 := NewCounterComponent("Clicks")
	counter1 := NewCounterComponent("Score")

	dirty := true

	var prevBuf *render.Buffer
	doRender := func() {
		termW, termH := term.GetSize(int(os.Stdin.Fd()))
		root := &layout.Input{
			Kind:      layout.KindBox,
			Direction: "column",
			Children: []*layout.Input{
				{
					Kind:    layout.KindBox,
					Padding: layout.ParsePadding("1 2"),
					Border:  "single",
					Children: []*layout.Input{
						{
							Kind:    layout.KindText,
							Content: "Sumi Components Demo",
							Style: render.Style{
								FG:   render.Color{Name: "green"},
								Bold: true,
							},
						},
						{
							Kind:    layout.KindText,
							Content: "Press any key to increment, q to quit",
							Style: render.Style{
								FG:  render.Color{Name: "cyan"},
								Dim: true,
							},
						},
						counter0.Layout(),
						counter1.Layout(),
					},
				},
			},
		}
		tree := layout.Layout(root, termW, termH)
		buf := render.NewBuffer(termW, termH)
		renderTree(buf, tree)
		if prevBuf != nil {
			buf.RenderTo(os.Stdout)
		} else {
			buf.RenderTo(os.Stdout)
		}
		prevBuf = buf
		dirty = false
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(os.Stdout)
	defer render.ExitAlternateScreen(os.Stdout)

	keyCh := make(chan rune)
	go func() {
		for {
			key, err := input.ReadKey(os.Stdin)
			if err != nil {
				close(keyCh)
				return
			}
			keyCh <- key
		}
	}()

	resizeCh, stopResize := term.WatchResize()
	defer stopResize()

	doRender()

	for {
		select {
		case key, ok := <-keyCh:
			if !ok || key == 'q' {
				return
			}
			counter0.HandleKey(key)
			counter1.HandleKey(key)
		case <-resizeCh:
			dirty = true
		}
		if dirty || counter0.Dirty() || counter1.Dirty() {
			doRender()
		}
	}
}

func renderTree(buf *render.Buffer, box *layout.Box) {
	if box.Border != "" && box.Border != "none" {
		buf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)
	}
	if box.Content != "" {
		buf.WriteStyledText(box.Y, box.X, box.Content, box.Style)
	}
	for _, child := range box.Children {
		renderTree(buf, child)
	}
}
