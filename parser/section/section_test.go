package section

import (
	"testing"
)

func TestParseTemplateOnly(t *testing.T) {
	// Given
	input := `<span>Hello</span>`

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Template != "<span>Hello</span>" {
		t.Errorf("Template = %q, want %q", got.Template, "<span>Hello</span>")
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

<span>Hello</span>`

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
	if got.Template != "<span>Hello</span>" {
		t.Errorf("Template = %q, want %q", got.Template, "<span>Hello</span>")
	}
}

func TestParseScriptAndTemplate(t *testing.T) {
	// Given
	input := `<script>
name := $state("world")
</script>

<span>Hello, {name}</span>`

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
	if got.Template != "<span>Hello, {name}</span>" {
		t.Errorf("Template = %q, want %q", got.Template, "<span>Hello, {name}</span>")
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

func TestParseImportsSection(t *testing.T) {
	// Given
	input := `<sumi:imports>
    "myui"
    alias "github.com/someone/otherui"
</sumi:imports>

<script>
x := $state(0)
</script>

<span>Hello</span>`

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Imports != "\n    \"myui\"\n    alias \"github.com/someone/otherui\"\n" {
		t.Errorf("Imports = %q", got.Imports)
	}
	if got.Script != "\nx := $state(0)\n" {
		t.Errorf("Script = %q, want %q", got.Script, "\nx := $state(0)\n")
	}
	if got.Template != "<span>Hello</span>" {
		t.Errorf("Template = %q, want %q", got.Template, "<span>Hello</span>")
	}
}

func TestParseNoImportsSection(t *testing.T) {
	// Given
	input := `<script>x := 1</script>
<span>Hello</span>`

	// When
	got, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Imports != "" {
		t.Errorf("Imports = %q, want empty", got.Imports)
	}
}

func TestParseWhitespaceBetweenSections(t *testing.T) {
	// Given
	input := `<script>
x := 1
</script>


<span>Hello</span>
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
	if got.Template != "<span>Hello</span>" {
		t.Errorf("Template = %q, want %q", got.Template, "<span>Hello</span>")
	}
}
