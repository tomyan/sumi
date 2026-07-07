package gen

import "testing"

func TestOutputPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"bare file", "counter.sumi", "counter_sumi.go"},
		{"with directory", "examples/counter/counter.sumi", "examples/counter/counter_sumi.go"},
		{"hyphenated name preserved", "split-panel.sumi", "split-panel_sumi.go"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := OutputPath(tc.in); got != tc.want {
				t.Errorf("OutputPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsSignalScript(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"signal New", "count := sumi.New(0)", true},
		{"derived From", "d := sumi.From(func() int { return 1 })", true},
		{"effect", "sumi.Effect(func() {})", true},
		{"env", "w := sumi.Env(\"width\")", true},
		{"signal package", "s := signal.New(0)", true},
		{"var declaration", "func f() {}\nvar x = 1", true},
		{"leading var", "var x = 1", true},
		{"plain handler only", "func handle() { doThing() }", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSignalScript(tc.src); got != tc.want {
				t.Errorf("isSignalScript(%q) = %v, want %v", tc.src, got, tc.want)
			}
		})
	}
}

func TestGenerateProducesGoForStaticTemplate(t *testing.T) {
	// Given a minimal static template
	src := `<span>Hello, Sumi!</span>`

	// When
	out, err := Generate("hello.sumi", src)

	// Then it emits a Go file for the derived package
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Generate returned no code")
	}
	if !contains(string(out), "package") {
		t.Errorf("expected a package clause in output:\n%s", out)
	}
}

func TestGenerateReportsParseError(t *testing.T) {
	// Given malformed template markup
	_, err := Generate("bad.sumi", `<span>Hello`)

	// Then the error is surfaced with the source path
	if err == nil {
		t.Fatal("expected error for malformed .sumi source, got nil")
	}
	if !contains(err.Error(), "bad.sumi") {
		t.Errorf("expected error to name the source path, got: %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return len(substr) == 0
}
