package css

import (
	"testing"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/render"
)

func TestResolveClassMatch(t *testing.T) {
	// Given
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red"}},
		},
	}

	// When
	props := Resolve(ss, "text", []string{"title"})

	// Then
	if got := props["color"]; got != "red" {
		t.Errorf("color = %q, want %q", got, "red")
	}
}

func TestResolveElementMatch(t *testing.T) {
	// Given
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "text", Properties: map[string]string{"bold": "true"}},
		},
	}

	// When
	props := Resolve(ss, "text", nil)

	// Then
	if got := props["bold"]; got != "true" {
		t.Errorf("bold = %q, want %q", got, "true")
	}
}

func TestResolveNoMatch(t *testing.T) {
	// Given
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red"}},
		},
	}

	// When
	props := Resolve(ss, "text", []string{"subtitle"})

	// Then
	if got := props["color"]; got != "" {
		t.Errorf("color = %q, want empty", got)
	}
}

func TestResolveMultipleRulesMerge(t *testing.T) {
	// Given
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red", "bold": "true"}},
			{Selector: "text", Properties: map[string]string{"color": "blue"}},
		},
	}

	// When
	props := Resolve(ss, "text", []string{"title"})

	// Then
	// Later rule wins for "color"
	if got := props["color"]; got != "blue" {
		t.Errorf("color = %q, want %q", got, "blue")
	}
	// "bold" from first rule survives
	if got := props["bold"]; got != "true" {
		t.Errorf("bold = %q, want %q", got, "true")
	}
}

func TestResolveNoRulesEmptyProperties(t *testing.T) {
	// Given
	ss := &style.Stylesheet{Rules: nil}

	// When
	props := Resolve(ss, "text", []string{"anything"})

	// Then
	if len(props) != 0 {
		t.Errorf("got %d properties, want 0", len(props))
	}
}

func TestResolveMultipleClassesMatchMultipleRules(t *testing.T) {
	// Given
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".primary", Properties: map[string]string{"color": "blue"}},
			{Selector: ".bold", Properties: map[string]string{"bold": "true"}},
		},
	}

	// When
	props := Resolve(ss, "text", []string{"primary", "bold"})

	// Then
	if got := props["color"]; got != "blue" {
		t.Errorf("color = %q, want %q", got, "blue")
	}
	if got := props["bold"]; got != "true" {
		t.Errorf("bold = %q, want %q", got, "true")
	}
}

func TestResolveElementSelectorDoesNotMatchWrongTag(t *testing.T) {
	// Given
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "box", Properties: map[string]string{"color": "red"}},
		},
	}

	// When
	props := Resolve(ss, "text", nil)

	// Then
	if got := props["color"]; got != "" {
		t.Errorf("color = %q, want empty", got)
	}
}

// --- ToRenderStyle tests ---

func TestToRenderStyleColor(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"color": "red"})

	// Then
	if s.FG.Name != "red" {
		t.Errorf("FG.Name = %q, want %q", s.FG.Name, "red")
	}
}

func TestToRenderStyleBackground(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"background": "blue"})

	// Then
	if s.BG.Name != "blue" {
		t.Errorf("BG.Name = %q, want %q", s.BG.Name, "blue")
	}
}

func TestToRenderStyleBold(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"bold": "true"})

	// Then
	if !s.Bold {
		t.Error("Bold = false, want true")
	}
}

func TestToRenderStyleItalic(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"italic": "true"})

	// Then
	if !s.Italic {
		t.Error("Italic = false, want true")
	}
}

func TestToRenderStyleUnderline(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"underline": "true"})

	// Then
	if !s.Underline {
		t.Error("Underline = false, want true")
	}
}

func TestToRenderStyleDim(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"dim": "true"})

	// Then
	if !s.Dim {
		t.Error("Dim = false, want true")
	}
}

func TestToRenderStyleStrikethrough(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"strikethrough": "true"})

	// Then
	if !s.Strikethrough {
		t.Error("Strikethrough = false, want true")
	}
}

func TestToRenderStyleInverse(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"inverse": "true"})

	// Then
	if !s.Inverse {
		t.Error("Inverse = false, want true")
	}
}

func TestToRenderStyleMultipleProperties(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{
		"color":      "cyan",
		"background": "black",
		"bold":       "true",
		"underline":  "true",
	})

	// Then
	if s.FG.Name != "cyan" {
		t.Errorf("FG.Name = %q, want %q", s.FG.Name, "cyan")
	}
	if s.BG.Name != "black" {
		t.Errorf("BG.Name = %q, want %q", s.BG.Name, "black")
	}
	if !s.Bold {
		t.Error("Bold = false, want true")
	}
	if !s.Underline {
		t.Error("Underline = false, want true")
	}
}

func TestToRenderStyleEmptyProperties(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{})

	// Then
	if s.FG.Name != "" {
		t.Errorf("FG.Name = %q, want empty", s.FG.Name)
	}
	if s.BG.Name != "" {
		t.Errorf("BG.Name = %q, want empty", s.BG.Name)
	}
	if s.Bold || s.Italic || s.Underline || s.Dim || s.Strikethrough || s.Inverse {
		t.Error("expected all bool fields to be false")
	}
}

func TestToRenderStyleBorderColor(t *testing.T) {
	// When
	s := ToRenderStyle(map[string]string{"border-color": "cyan"})

	// Then
	if s.FG.Name != "cyan" {
		t.Errorf("FG.Name = %q, want %q", s.FG.Name, "cyan")
	}
}

func TestToRenderStyleBorderColorOverridesColor(t *testing.T) {
	// Given — both color and border-color are set
	// When
	s := ToRenderStyle(map[string]string{
		"color":        "red",
		"border-color": "cyan",
	})

	// Then — border-color wins for FG
	if s.FG.Name != "cyan" {
		t.Errorf("FG.Name = %q, want %q (border-color should override color)", s.FG.Name, "cyan")
	}
}

// Ensure render.Style is used (compile-time check)
var _ render.Style
