package script

import "testing"

func TestSingleScrollDecl(t *testing.T) {
	// Given
	input := `pos := $scroll(main)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.ScrollDecls) != 1 {
		t.Fatalf("expected 1 scroll decl, got %d", len(s.ScrollDecls))
	}
	if s.ScrollDecls[0].Name != "pos" {
		t.Errorf("name: got %q, want %q", s.ScrollDecls[0].Name, "pos")
	}
	if s.ScrollDecls[0].BoxID != "main" {
		t.Errorf("boxID: got %q, want %q", s.ScrollDecls[0].BoxID, "main")
	}
}

func TestMultipleScrollDecls(t *testing.T) {
	// Given
	input := `left := $scroll(leftPanel)
right := $scroll(rightPanel)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.ScrollDecls) != 2 {
		t.Fatalf("expected 2 scroll decls, got %d", len(s.ScrollDecls))
	}
	if s.ScrollDecls[0].BoxID != "leftPanel" {
		t.Errorf("first boxID: got %q, want %q", s.ScrollDecls[0].BoxID, "leftPanel")
	}
	if s.ScrollDecls[1].BoxID != "rightPanel" {
		t.Errorf("second boxID: got %q, want %q", s.ScrollDecls[1].BoxID, "rightPanel")
	}
}

func TestMixedStateAndScroll(t *testing.T) {
	// Given
	input := `count := $state(0)
scrollPos := $scroll(content)

func increment() {
	count = count + 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Errorf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if len(s.ScrollDecls) != 1 {
		t.Errorf("expected 1 scroll decl, got %d", len(s.ScrollDecls))
	}
	if s.ScrollDecls[0].Name != "scrollPos" {
		t.Errorf("scroll name: got %q, want %q", s.ScrollDecls[0].Name, "scrollPos")
	}
}

func TestUnterminatedScroll(t *testing.T) {
	// When
	_, err := Parse(`pos := $scroll(main`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated $scroll, got nil")
	}
}

func TestEmptyScriptNoScrollDecls(t *testing.T) {
	// When
	s, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.ScrollDecls) != 0 {
		t.Errorf("expected 0 scroll decls, got %d", len(s.ScrollDecls))
	}
}
