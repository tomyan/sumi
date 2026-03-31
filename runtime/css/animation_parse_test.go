package css

import (
	"testing"

	"github.com/tomyan/sumi/runtime/anim"
)

func TestParseAnimationShorthand(t *testing.T) {
	props := map[string]string{
		"animation": "pulse 2s infinite ease-in-out",
	}
	spec := ParseAnimation(props)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.Name != "pulse" {
		t.Errorf("Name = %q, want %q", spec.Name, "pulse")
	}
	if spec.DurationMs != 2000 {
		t.Errorf("DurationMs = %d, want 2000", spec.DurationMs)
	}
	if spec.IterationCount != -1 {
		t.Errorf("IterationCount = %d, want -1 (infinite)", spec.IterationCount)
	}
	if spec.TimingFunction != anim.EaseInOut {
		t.Errorf("TimingFunction = %v, want EaseInOut", spec.TimingFunction)
	}
}

func TestParseAnimationMinimal(t *testing.T) {
	props := map[string]string{
		"animation": "fade 500ms",
	}
	spec := ParseAnimation(props)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.Name != "fade" {
		t.Errorf("Name = %q, want %q", spec.Name, "fade")
	}
	if spec.DurationMs != 500 {
		t.Errorf("DurationMs = %d, want 500", spec.DurationMs)
	}
	if spec.IterationCount != 1 {
		t.Errorf("IterationCount = %d, want 1", spec.IterationCount)
	}
}

func TestParseAnimationWithDirection(t *testing.T) {
	props := map[string]string{
		"animation": "bounce 1s alternate 3",
	}
	spec := ParseAnimation(props)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.Direction != "alternate" {
		t.Errorf("Direction = %q, want %q", spec.Direction, "alternate")
	}
	if spec.IterationCount != 3 {
		t.Errorf("IterationCount = %d, want 3", spec.IterationCount)
	}
}

func TestParseAnimationEmpty(t *testing.T) {
	spec := ParseAnimation(map[string]string{})
	if spec != nil {
		t.Error("expected nil for empty props")
	}
}

func TestParseAnimationLonghand(t *testing.T) {
	props := map[string]string{
		"animation-name":            "pulse",
		"animation-duration":        "2s",
		"animation-timing-function": "ease-in",
		"animation-iteration-count": "infinite",
		"animation-direction":       "alternate",
		"animation-fill-mode":       "forwards",
	}
	spec := ParseAnimation(props)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.Name != "pulse" || spec.DurationMs != 2000 || spec.IterationCount != -1 {
		t.Errorf("spec = %+v", spec)
	}
	if spec.Direction != "alternate" || spec.FillMode != "forwards" {
		t.Errorf("direction=%q fillMode=%q", spec.Direction, spec.FillMode)
	}
}
