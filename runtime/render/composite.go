package render

// compositeStyle blends a style's translucent colours (A != 0) with the
// cell currently at (row, col). Backgrounds blend over the existing
// background; foregrounds blend over the effective backdrop (the style's
// own background if set, else the existing one). Non-RGB backdrops can't
// blend — the source paints opaque. Results are opaque (A = 0).
func (b *Buffer) compositeStyle(row, col int, style Style) Style {
	if style.BG.A == 0 && style.FG.A == 0 {
		return style
	}
	dst := b.cells[row][col].Style
	if style.BG.A != 0 {
		style.BG = blendColor(style.BG, dst.BG)
	}
	if style.FG.A != 0 {
		backdrop := style.BG
		if !backdrop.IsRGB && backdrop.Name == "" {
			backdrop = dst.BG
		}
		style.FG = blendColor(style.FG, backdrop)
	}
	return style
}

// blendColor composites src over dst by src's alpha. Sources or
// backdrops outside RGB space fall back to painting src opaque.
func blendColor(src, dst Color) Color {
	alpha := src.A
	src.A = 0
	if !src.IsRGB || !dst.IsRGB {
		return src
	}
	a := float64(alpha) / 255
	return Color{
		IsRGB: true,
		R:     uint8(float64(src.R)*a + float64(dst.R)*(1-a) + 0.5),
		G:     uint8(float64(src.G)*a + float64(dst.G)*(1-a) + 0.5),
		B:     uint8(float64(src.B)*a + float64(dst.B)*(1-a) + 0.5),
	}
}
