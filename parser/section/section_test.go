package section

import (
	"testing"
)

func TestParseTemplateOnly(t *testing.T) {
	// Given
	input := `<text>Hello</text>`

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Template != "<text>Hello</text>" {
		t.Errorf("Template = %q, want %q", got.Template, "<text>Hello</text>")
	}
	if got.Script != "" {
		t.Errorf("Script = %q, want empty", got.Script)
	}
	if got.Style != "" {
		t.Errorf("Style = %q, want empty", got.Style)
	}
}

func TestParseAllThreeSections(t *testing.T) {
	// Given
	input := `<script>
count := $state(0)
</script>

<style>
.title { color: green; }
</style>

<text>Hello</text>`

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Script != "\ncount := $state(0)\n" {
		t.Errorf("Script = %q, want %q", got.Script, "\ncount := $state(0)\n")
	}
	if got.Style != "\n.title { color: green; }\n" {
		t.Errorf("Style = %q, want %q", got.Style, "\n.title { color: green; }\n")
	}
	if got.Template != "<text>Hello</text>" {
		t.Errorf("Template = %q, want %q", got.Template, "<text>Hello</text>")
	}
}

func TestParseScriptAndTemplate(t *testing.T) {
	// Given
	input := `<script>
name := $state("world")
</script>

<text>Hello, {name}</text>`

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Script != "\nname := $state(\"world\")\n" {
		t.Errorf("Script = %q, want %q", got.Script, "\nname := $state(\"world\")\n")
	}
	if got.Style != "" {
		t.Errorf("Style = %q, want empty", got.Style)
	}
	if got.Template != "<text>Hello, {name}</text>" {
		t.Errorf("Template = %q, want %q", got.Template, "<text>Hello, {name}</text>")
	}
}

func TestParseEmptyInput(t *testing.T) {
	// When
	got, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Template != "" {
		t.Errorf("Template = %q, want empty", got.Template)
	}
	if got.Script != "" {
		t.Errorf("Script = %q, want empty", got.Script)
	}
	if got.Style != "" {
		t.Errorf("Style = %q, want empty", got.Style)
	}
}

func TestParseWhitespaceBetweenSections(t *testing.T) {
	// Given
	input := `<script>
x := 1
</script>


<text>Hello</text>
  `

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Script != "\nx := 1\n" {
		t.Errorf("Script = %q, want %q", got.Script, "\nx := 1\n")
	}
	if got.Template != "<text>Hello</text>" {
		t.Errorf("Template = %q, want %q", got.Template, "<text>Hello</text>")
	}
}
