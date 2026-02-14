package preview

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestParseStyledLinePlainText(t *testing.T) {
	// Given
	line := "hello world"

	// When
	segments := parseStyledLine(line)

	// Then
	if len(segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(segments))
	}
	if segments[0].text != "hello world" {
		t.Errorf("text = %q, want %q", segments[0].text, "hello world")
	}
}

func TestParseStyledLineBold(t *testing.T) {
	// Given
	line := "hello <<bold>>world<</>>"

	// When
	segments := parseStyledLine(line)

	// Then
	if len(segments) != 2 {
		t.Fatalf("got %d segments, want 2", len(segments))
	}
	if segments[0].text != "hello " {
		t.Errorf("segment[0].text = %q, want %q", segments[0].text, "hello ")
	}
	if segments[1].text != "world" {
		t.Errorf("segment[1].text = %q, want %q", segments[1].text, "world")
	}
	if !segments[1].style.Bold {
		t.Error("expected bold style on segment 1")
	}
}

func TestParseStyledLineColor(t *testing.T) {
	// Given
	line := "<<red>>error<</>>"

	// When
	segments := parseStyledLine(line)

	// Then
	if len(segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(segments))
	}
	if segments[0].style.FG != (render.Color{Name: "red"}) {
		t.Errorf("FG = %v, want red", segments[0].style.FG)
	}
}

func TestParseStyledLineBackground(t *testing.T) {
	// Given
	line := "<<bg:blue,white>>alert<</>>"

	// When
	segments := parseStyledLine(line)

	// Then
	if len(segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(segments))
	}
	if segments[0].style.BG != (render.Color{Name: "blue"}) {
		t.Errorf("BG = %v, want blue", segments[0].style.BG)
	}
	if segments[0].style.FG != (render.Color{Name: "white"}) {
		t.Errorf("FG = %v, want white", segments[0].style.FG)
	}
}

func TestParseStyledLineResetMidline(t *testing.T) {
	// Given
	line := "<<bold>>bold<</>> normal"

	// When
	segments := parseStyledLine(line)

	// Then
	if len(segments) != 2 {
		t.Fatalf("got %d segments, want 2", len(segments))
	}
	if !segments[0].style.Bold {
		t.Error("expected bold on segment 0")
	}
	if segments[1].style.Bold {
		t.Error("expected no bold on segment 1 (after reset)")
	}
}
