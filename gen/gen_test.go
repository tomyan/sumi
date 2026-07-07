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

func TestGenerateStaticTemplateProducesComponentForm(t *testing.T) {
	// Given a script-free template
	src := `<span>Hello, Sumi!</span>`

	// When
	out, err := Generate("hello.sumi", src)

	// Then it compiles to the same constructor form as reactive components —
	// no legacy static func Run() path.
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	s := string(out)
	if !contains(s, "func NewHello(props HelloProps) *sumi.Component") {
		t.Errorf("expected NewHello constructor in output:\n%s", s)
	}
	if !contains(s, "type HelloProps struct") {
		t.Errorf("expected HelloProps struct in output:\n%s", s)
	}
	if contains(s, "func Run()") {
		t.Errorf("static func Run() path must be retired:\n%s", s)
	}
}

func TestGenerateStaticParentMountsChild(t *testing.T) {
	// Given a script-free parent that mounts a local child component
	src := `<div><Widget /></div>`

	// When
	out, err := Generate("parent.sumi", src)

	// Then the parent instantiates the child and retains it in Children, so
	// the child resolves against its own stylesheet at runtime (previously a
	// script-free parent fell to the whole-app path and left the child
	// uninstantiated).
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	s := string(out)
	if !contains(s, "NewWidget(WidgetProps{") {
		t.Errorf("expected child instantiation in output:\n%s", s)
	}
	if !contains(s, "Children: []*sumi.Component{widget0}") {
		t.Errorf("expected child retained in Children:\n%s", s)
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
