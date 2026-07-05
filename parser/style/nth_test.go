package style

import "testing"

// A5: An+B parsing for :nth-* pseudo-classes.

func TestParseNthKeywords(t *testing.T) {
	cases := []struct {
		expr string
		a, b int
	}{
		{"odd", 2, 1},
		{"even", 2, 0},
		{"3", 0, 3},
		{"n", 1, 0},
		{"2n", 2, 0},
		{"2n+1", 2, 1},
		{"-n+3", -1, 3},
		{"3n-1", 3, -1},
		{"+2n+2", 2, 2},
	}
	for _, c := range cases {
		nth, err := ParseNth(c.expr)
		if err != nil {
			t.Errorf("ParseNth(%q) error: %v", c.expr, err)
			continue
		}
		if nth.A != c.a || nth.B != c.b {
			t.Errorf("ParseNth(%q) = %+v, want A=%d B=%d", c.expr, nth, c.a, c.b)
		}
	}
}

func TestNthMatches(t *testing.T) {
	// nth-child indices are 1-based.
	cases := []struct {
		expr    string
		index   int
		matches bool
	}{
		{"odd", 1, true},
		{"odd", 2, false},
		{"even", 2, true},
		{"3", 3, true},
		{"3", 4, false},
		{"2n+1", 5, true},
		{"-n+3", 3, true},
		{"-n+3", 4, false},
		{"n", 7, true},
	}
	for _, c := range cases {
		nth, err := ParseNth(c.expr)
		if err != nil {
			t.Fatalf("ParseNth(%q): %v", c.expr, err)
		}
		if got := nth.Matches(c.index); got != c.matches {
			t.Errorf("%q.Matches(%d) = %v, want %v", c.expr, c.index, got, c.matches)
		}
	}
}

func TestParseNthInvalid(t *testing.T) {
	for _, expr := range []string{"", "x", "n+n", "2m+1"} {
		if _, err := ParseNth(expr); err == nil {
			t.Errorf("ParseNth(%q) should error", expr)
		}
	}
}
