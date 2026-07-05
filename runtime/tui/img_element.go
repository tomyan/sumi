package tui

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
)

// syncImgElement renders the src image as half-blocks into Cells: each
// cell shows two vertically stacked pixels via '▀' (fg = top pixel,
// bg = bottom pixel; transparent halves show the terminal background).
// The decode is cached against the src attribute.
func syncImgElement(n *layout.Input) {
	src := n.Attrs["src"]
	if src == "" || len(src) > 0 && src[0] == '{' {
		return
	}
	if n.Attrs["sumi:img-src"] == src {
		return // already rendered for this src
	}
	if n.Attrs == nil {
		n.Attrs = map[string]string{}
	}
	n.Attrs["sumi:img-src"] = src

	img := loadImage(src)
	if img == nil {
		n.Cells = nil
		return
	}
	cols, rows := imgCellSize(n, img)
	if cols <= 0 || rows <= 0 {
		n.Cells = nil
		return
	}
	n.Cells = paintHalfBlocks(img, cols, rows)
	if _, ok := n.Attrs["width"]; !ok {
		n.FixedWidth = cols
	}
	if _, ok := n.Attrs["height"]; !ok {
		n.FixedHeight = rows
	}
}

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return img
}

// imgCellSize picks the cell grid: width/height attributes in cells, or
// the intrinsic pixel size (one column per pixel, two pixel rows per cell).
func imgCellSize(n *layout.Input, img image.Image) (cols, rows int) {
	bounds := img.Bounds()
	cols = bounds.Dx()
	rows = (bounds.Dy() + 1) / 2
	if v, ok := n.Attrs["width"]; ok {
		if w, err := strconv.Atoi(v); err == nil && w > 0 {
			cols = w
		}
	}
	if v, ok := n.Attrs["height"]; ok {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			rows = h
		}
	}
	return cols, rows
}

// paintHalfBlocks samples the image (nearest neighbour) onto a cols×rows
// cell grid of half-blocks.
func paintHalfBlocks(img image.Image, cols, rows int) *render.Buffer {
	buf := render.NewBuffer(cols, rows)
	bounds := img.Bounds()
	pxW, pxH := bounds.Dx(), bounds.Dy()
	sample := func(col, halfRow int) (render.Color, bool) {
		x := bounds.Min.X + col*pxW/cols
		y := bounds.Min.Y + halfRow*pxH/(rows*2)
		r, g, b, a := img.At(x, y).RGBA()
		if a < 0x8000 {
			return render.Color{}, false
		}
		return render.Color{IsRGB: true, R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8)}, true
	}
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			top, topOK := sample(col, row*2)
			bottom, bottomOK := sample(col, row*2+1)
			switch {
			case topOK && bottomOK:
				buf.SetStyledCell(row, col, '▀', render.Style{FG: top, BG: bottom})
			case topOK:
				buf.SetStyledCell(row, col, '▀', render.Style{FG: top})
			case bottomOK:
				buf.SetStyledCell(row, col, '▄', render.Style{FG: bottom})
			}
		}
	}
	return buf
}
