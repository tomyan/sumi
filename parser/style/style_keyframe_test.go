package style

import "testing"

func TestParseKeyframesBasic(t *testing.T) {
	input := `
@keyframes pulse {
  0% { color: #50fa7b; }
  50% { color: #2d8a4e; }
  100% { color: #50fa7b; }
}
`
	s, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Keyframes) != 1 {
		t.Fatalf("got %d keyframes, want 1", len(s.Keyframes))
	}
	kf := s.Keyframes[0]
	if kf.Name != "pulse" {
		t.Errorf("name = %q, want %q", kf.Name, "pulse")
	}
	if len(kf.Stops) != 3 {
		t.Fatalf("got %d stops, want 3", len(kf.Stops))
	}
	if kf.Stops[0].Percent != 0 {
		t.Errorf("stop 0 percent = %v, want 0", kf.Stops[0].Percent)
	}
	if kf.Stops[1].Percent != 0.5 {
		t.Errorf("stop 1 percent = %v, want 0.5", kf.Stops[1].Percent)
	}
	if kf.Stops[2].Percent != 1 {
		t.Errorf("stop 2 percent = %v, want 1", kf.Stops[2].Percent)
	}
	if kf.Stops[0].Properties["color"] != "#50fa7b" {
		t.Errorf("stop 0 color = %q", kf.Stops[0].Properties["color"])
	}
}

func TestParseKeyframesFromTo(t *testing.T) {
	input := `
@keyframes fade {
  from { color: red; }
  to { color: blue; }
}
`
	s, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Keyframes) != 1 {
		t.Fatalf("got %d keyframes, want 1", len(s.Keyframes))
	}
	kf := s.Keyframes[0]
	if len(kf.Stops) != 2 {
		t.Fatalf("got %d stops, want 2", len(kf.Stops))
	}
	if kf.Stops[0].Percent != 0 {
		t.Errorf("from percent = %v, want 0", kf.Stops[0].Percent)
	}
	if kf.Stops[1].Percent != 1 {
		t.Errorf("to percent = %v, want 1", kf.Stops[1].Percent)
	}
}

func TestParseKeyframesWithRules(t *testing.T) {
	input := `
.dot { color: green; }
@keyframes pulse {
  0% { color: red; }
  100% { color: blue; }
}
.other { bold: true; }
`
	s, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Rules) != 2 {
		t.Errorf("got %d rules, want 2", len(s.Rules))
	}
	if len(s.Keyframes) != 1 {
		t.Errorf("got %d keyframes, want 1", len(s.Keyframes))
	}
}

func TestParseMultipleKeyframes(t *testing.T) {
	input := `
@keyframes a { from { color: red; } to { color: blue; } }
@keyframes b { 0% { bold: true; } 100% { bold: false; } }
`
	s, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Keyframes) != 2 {
		t.Fatalf("got %d keyframes, want 2", len(s.Keyframes))
	}
	if s.Keyframes[0].Name != "a" || s.Keyframes[1].Name != "b" {
		t.Errorf("names = %q, %q", s.Keyframes[0].Name, s.Keyframes[1].Name)
	}
}
