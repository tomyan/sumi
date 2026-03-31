package render

import "testing"

func TestNamedColorToRGB(t *testing.T) {
	tests := []struct {
		name    string
		wantR   uint8
		wantG   uint8
		wantB   uint8
	}{
		{"red", 255, 0, 0},
		{"green", 0, 255, 0},
		{"blue", 0, 0, 255},
		{"black", 0, 0, 0},
		{"white", 255, 255, 255},
	}
	for _, tc := range tests {
		c := Color{Name: tc.name}
		got := c.ToRGB()
		if !got.IsRGB {
			t.Errorf("%s.ToRGB().IsRGB = false", tc.name)
			continue
		}
		if got.R != tc.wantR || got.G != tc.wantG || got.B != tc.wantB {
			t.Errorf("%s.ToRGB() = (%d,%d,%d), want (%d,%d,%d)",
				tc.name, got.R, got.G, got.B, tc.wantR, tc.wantG, tc.wantB)
		}
	}
}

func TestRGBPassthrough(t *testing.T) {
	c := Color{IsRGB: true, R: 100, G: 200, B: 50}
	got := c.ToRGB()
	if got != c {
		t.Errorf("RGB.ToRGB() changed: %v → %v", c, got)
	}
}

func TestLerpColorBlackToWhite(t *testing.T) {
	black := Color{IsRGB: true, R: 0, G: 0, B: 0}
	white := Color{IsRGB: true, R: 255, G: 255, B: 255}

	mid := LerpColor(black, white, 0.5)
	// Expect ~128 for each channel.
	if mid.R < 127 || mid.R > 128 {
		t.Errorf("mid.R = %d, want ~128", mid.R)
	}
	if mid.G < 127 || mid.G > 128 {
		t.Errorf("mid.G = %d, want ~128", mid.G)
	}
}

func TestLerpColorBoundaries(t *testing.T) {
	from := Color{IsRGB: true, R: 80, G: 250, B: 123}
	to := Color{IsRGB: true, R: 45, G: 138, B: 78}

	atZero := LerpColor(from, to, 0.0)
	if atZero.R != from.R || atZero.G != from.G || atZero.B != from.B {
		t.Errorf("LerpColor at 0 = (%d,%d,%d), want (%d,%d,%d)",
			atZero.R, atZero.G, atZero.B, from.R, from.G, from.B)
	}

	atOne := LerpColor(from, to, 1.0)
	if atOne.R != to.R || atOne.G != to.G || atOne.B != to.B {
		t.Errorf("LerpColor at 1 = (%d,%d,%d), want (%d,%d,%d)",
			atOne.R, atOne.G, atOne.B, to.R, to.G, to.B)
	}
}

func TestLerpColorNamedColors(t *testing.T) {
	red := Color{Name: "red"}
	blue := Color{Name: "blue"}

	mid := LerpColor(red, blue, 0.5)
	if !mid.IsRGB {
		t.Fatal("expected RGB result from named color lerp")
	}
	// red(255,0,0) → blue(0,0,255) at 0.5 → (128,0,128) approximately.
	if mid.R < 127 || mid.R > 128 {
		t.Errorf("mid.R = %d, want ~128", mid.R)
	}
	if mid.B < 127 || mid.B > 128 {
		t.Errorf("mid.B = %d, want ~128", mid.B)
	}
}

func TestLerpColorZeroColorReturnsOther(t *testing.T) {
	// If from is zero (unset), return to immediately.
	zero := Color{}
	green := Color{IsRGB: true, R: 0, G: 255, B: 0}

	got := LerpColor(zero, green, 0.5)
	if got != green {
		t.Errorf("LerpColor(zero, green, 0.5) = %v, want %v", got, green)
	}

	got2 := LerpColor(green, zero, 0.5)
	if got2 != green {
		t.Errorf("LerpColor(green, zero, 0.5) = %v, want %v", got2, green)
	}
}
