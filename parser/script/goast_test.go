package script

import "testing"

func TestParseGoASTIdentifiesSignals(t *testing.T) {
	input := `count := signal.New(0)
doubled := signal.From(func() int { return count.Get() * 2 })
visible := true`

	// When
	info, err := ParseGoAST(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Signals["count"] {
		t.Error("expected count to be identified as signal")
	}
	if !info.Signals["doubled"] {
		t.Error("expected doubled to be identified as signal")
	}
	if info.Signals["visible"] {
		t.Error("visible should not be a signal")
	}
}

func TestParseGoASTIdentifiesProps(t *testing.T) {
	input := `var value *signal.Signal[string]
var label string = "Count"
var readonly bool`

	// When
	info, err := ParseGoAST(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.Props) != 3 {
		t.Fatalf("len(Props) = %d, want 3", len(info.Props))
	}

	// Check prop names and types
	propMap := make(map[string]PropInfo)
	for _, p := range info.Props {
		propMap[p.Name] = p
	}

	if propMap["value"].TypeStr != "*signal.Signal[string]" {
		t.Errorf("value type = %q, want *signal.Signal[string]", propMap["value"].TypeStr)
	}
	if propMap["label"].TypeStr != "string" {
		t.Errorf("label type = %q, want string", propMap["label"].TypeStr)
	}
	if propMap["label"].Default != `"Count"` {
		t.Errorf("label default = %q, want %q", propMap["label"].Default, `"Count"`)
	}
	if propMap["readonly"].TypeStr != "bool" {
		t.Errorf("readonly type = %q, want bool", propMap["readonly"].TypeStr)
	}
}

func TestParseGoASTIdentifiesFunctions(t *testing.T) {
	input := `count := signal.New(0)

func increment() {
    count.Update(func(n int) int { return n + 1 })
}

func handleKey(evt input.Event) {
    if evt.Rune == 'q' { app.Quit() }
}`

	// When
	info, err := ParseGoAST(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.Funcs) != 2 {
		t.Fatalf("len(Funcs) = %d, want 2", len(info.Funcs))
	}
	if info.Funcs[0].Name != "increment" {
		t.Errorf("Funcs[0].Name = %q, want increment", info.Funcs[0].Name)
	}
	if info.Funcs[1].Name != "handleKey" {
		t.Errorf("Funcs[1].Name = %q, want handleKey", info.Funcs[1].Name)
	}
}
