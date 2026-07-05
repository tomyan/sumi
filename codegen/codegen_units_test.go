package codegen

import (
	"strings"
	"testing"
)

// A3: cell/ch units (1ch = 1cell) and percentage sizing.

func TestGenerateWidthWithCellUnit(t *testing.T) {
	// Given / When
	src := generateBox(t, map[string]string{"width": "20cell"})

	// Then
	if !containsField(src, "FixedWidth", "20") {
		t.Errorf("width: 20cell should emit FixedWidth: 20:\n%s", src)
	}
}

func TestGenerateWidthWithChUnit(t *testing.T) {
	src := generateBox(t, map[string]string{"width": "20ch"})
	if !containsField(src, "FixedWidth", "20") {
		t.Errorf("width: 20ch should emit FixedWidth: 20:\n%s", src)
	}
}

func TestGenerateGapWithCellUnit(t *testing.T) {
	src := generateBox(t, map[string]string{"gap": "1cell"})
	if !containsField(src, "Gap", "1") {
		t.Errorf("gap: 1cell should emit Gap: 1:\n%s", src)
	}
}

func TestGenerateWidthPercentage(t *testing.T) {
	src := generateBox(t, map[string]string{"width": "50%"})
	if !containsField(src, "WidthPct", "50") {
		t.Errorf("width: 50%% should emit WidthPct: 50:\n%s", src)
	}
	if strings.Contains(src, "FixedWidth") {
		t.Errorf("percentage width must not emit FixedWidth:\n%s", src)
	}
}

func TestGenerateHeightPercentage(t *testing.T) {
	src := generateBox(t, map[string]string{"height": "30%"})
	if !containsField(src, "HeightPct", "30") {
		t.Errorf("height: 30%% should emit HeightPct: 30:\n%s", src)
	}
}

func TestGeneratePercentageOnNonSizeAttrDropped(t *testing.T) {
	// gap has no percentage meaning here yet — must drop, not emit garbage.
	src := generateBox(t, map[string]string{"gap": "50%"})
	if strings.Contains(src, "Gap") {
		t.Errorf("gap: 50%% must be dropped:\n%s", src)
	}
}
