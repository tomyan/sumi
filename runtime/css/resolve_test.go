package css

import (
	"testing"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/render"
)

func TestResolveClassMatch(t *testing.T) {
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red"}},
		},
	}
	props := Resolve(ss, "text", []string{"title"})
	if got := props["color"]; got != "red" {
		t.Errorf("color = %q, want %q", got, "red")
	}
}

func TestResolveElementMatch(t *testing.T) {
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "text", Properties: map[string]string{"bold": "true"}},
		},
	}
	props := Resolve(ss, "text", nil)
	if got := props["bold"]; got != "true" {
		t.Errorf("bold = %q, want %q", got, "true")
	}
}

func TestResolveNoMatch(t *testing.T) {
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red"}},
		},
	}
	props := Resolve(ss, "text", []string{"subtitle"})
	if got := props["color"]; got != "" {
		t.Errorf("color = %q, want empty", got)
	}
}

func TestResolveMultipleRulesMerge(t *testing.T) {
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red", "bold": "true"}},
			{Selector: "text", Properties: map[string]string{"color": "blue"}},
		},
	}
	props := Resolve(ss, "text", []string{"title"})
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
	ss := &style.Stylesheet{Rules: nil}
	props := Resolve(ss, "text", []string{"anything"})
	if len(props) != 0 {
		t.Errorf("got %d properties, want 0", len(props))
	}
}

func TestResolveMultipleClassesMatchMultipleRules(t *testing.T) {
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".primary", Properties: map[string]string{"color": "blue"}},
			{Selector: ".bold", Properties: map[string]string{"bold": "true"}},
		},
	}
	props := Resolve(ss, "text", []string{"primary", "bold"})
	if got := props["color"]; got != "blue" {
		t.Errorf("color = %q, want %q", got, "blue")
	}
	if got := props["bold"]; got != "true" {
		t.Errorf("bold = %q, want %q", got, "true")
	}
}

func TestResolveElementSelectorDoesNotMatchWrongTag(t *testing.T) {
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "box", Properties: map[string]string{"color": "red"}},
		},
	}
	props := Resolve(ss, "text", nil)
	if got := props["color"]; got != "" {
		t.Errorf("color = %q, want empty", got)
	}
}

// --- ToRenderStyle tests ---

func TestToRenderStyleColor(t *testing.T) {
	s := ToRenderStyle(map[string]string{"color": "red"})
	if s.FG.Name != "red" {
		t.Errorf("FG.Name = %q, want %q", s.FG.Name, "red")
	}
}

func TestToRenderStyleBackground(t *testing.T) {
	s := ToRenderStyle(map[string]string{"background": "blue"})
	if s.BG.Name != "blue" {
		t.Errorf("BG.Name = %q, want %q", s.BG.Name, "blue")
	}
}

func TestToRenderStyleBold(t *testing.T) {
	s := ToRenderStyle(map[string]string{"bold": "true"})
	if !s.Bold {
		t.Error("Bold = false, want true")
	}
}

func TestToRenderStyleItalic(t *testing.T) {
	s := ToRenderStyle(map[string]string{"italic": "true"})
	if !s.Italic {
		t.Error("Italic = false, want true")
	}
}

func TestToRenderStyleUnderline(t *testing.T) {
	s := ToRenderStyle(map[string]string{"underline": "true"})
	if !s.Underline {
		t.Error("Underline = false, want true")
	}
}

func TestToRenderStyleDim(t *testing.T) {
	s := ToRenderStyle(map[string]string{"dim": "true"})
	if !s.Dim {
		t.Error("Dim = false, want true")
	}
}

func TestToRenderStyleStrikethrough(t *testing.T) {
	s := ToRenderStyle(map[string]string{"strikethrough": "true"})
	if !s.Strikethrough {
		t.Error("Strikethrough = false, want true")
	}
}

func TestToRenderStyleInverse(t *testing.T) {
	s := ToRenderStyle(map[string]string{"inverse": "true"})
	if !s.Inverse {
		t.Error("Inverse = false, want true")
	}
}

func TestToRenderStyleMultipleProperties(t *testing.T) {
	s := ToRenderStyle(map[string]string{
		"color":      "cyan",
		"background": "black",
		"bold":       "true",
		"underline":  "true",
	})
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
	s := ToRenderStyle(map[string]string{})
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

// Ensure render.Style is used (compile-time check)
var _ render.Style
