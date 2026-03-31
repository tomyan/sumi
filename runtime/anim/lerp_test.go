package anim

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestLerpInt(t *testing.T) {
	tests := []struct {
		from, to int
		t        float64
		want     int
	}{
		{0, 100, 0.0, 0},
		{0, 100, 0.5, 50},
		{0, 100, 1.0, 100},
		{10, 20, 0.25, 13},
		{-10, 10, 0.5, 0},
	}
	for _, tc := range tests {
		got := LerpInt(tc.from, tc.to, tc.t)
		if got != tc.want {
			t.Errorf("LerpInt(%d, %d, %v) = %d, want %d", tc.from, tc.to, tc.t, got, tc.want)
		}
	}
}

func TestLerpStyleColors(t *testing.T) {
	from := render.Style{
		FG: render.Color{IsRGB: true, R: 0, G: 0, B: 0},
		BG: render.Color{IsRGB: true, R: 255, G: 255, B: 255},
	}
	to := render.Style{
		FG: render.Color{IsRGB: true, R: 255, G: 255, B: 255},
		BG: render.Color{IsRGB: true, R: 0, G: 0, B: 0},
	}

	mid := LerpStyle(from, to, 0.5)

	if mid.FG.R < 127 || mid.FG.R > 128 {
		t.Errorf("mid FG.R = %d, want ~128", mid.FG.R)
	}
	if mid.BG.R < 127 || mid.BG.R > 128 {
		t.Errorf("mid BG.R = %d, want ~128", mid.BG.R)
	}
}

func TestLerpStyleBooleanSnapsAtHalf(t *testing.T) {
	from := render.Style{Bold: true, Dim: false}
	to := render.Style{Bold: false, Dim: true}

	// Before midpoint: use from's booleans.
	before := LerpStyle(from, to, 0.4)
	if !before.Bold {
		t.Error("expected Bold=true at t=0.4")
	}
	if before.Dim {
		t.Error("expected Dim=false at t=0.4")
	}

	// After midpoint: use to's booleans.
	after := LerpStyle(from, to, 0.6)
	if after.Bold {
		t.Error("expected Bold=false at t=0.6")
	}
	if !after.Dim {
		t.Error("expected Dim=true at t=0.6")
	}
}

func TestLerpStyleBoundaries(t *testing.T) {
	from := render.Style{FG: render.Color{IsRGB: true, R: 80, G: 250, B: 123}, Bold: true}
	to := render.Style{FG: render.Color{IsRGB: true, R: 45, G: 138, B: 78}, Bold: false}

	atZero := LerpStyle(from, to, 0.0)
	if atZero.FG.R != 80 || atZero.Bold != true {
		t.Errorf("at 0: FG.R=%d Bold=%v", atZero.FG.R, atZero.Bold)
	}

	atOne := LerpStyle(from, to, 1.0)
	if atOne.FG.R != 45 || atOne.Bold != false {
		t.Errorf("at 1: FG.R=%d Bold=%v", atOne.FG.R, atOne.Bold)
	}
}
