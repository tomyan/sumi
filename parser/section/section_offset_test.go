package section

import "testing"

func TestScriptStartOffsetInOriginal(t *testing.T) {
	// Given
	input := "<script>\nx := 1\n</script>\n<span>Hi</span>"

	// When
	got, _ := Parse(input)

	// Then: content starts right after "<script>" (8 bytes)
	if got.ScriptStart != 8 {
		t.Errorf("ScriptStart = %d, want 8", got.ScriptStart)
	}
	if input[got.ScriptStart:got.ScriptStart+len(got.Script)] != got.Script {
		t.Errorf("ScriptStart does not point at Script content in original input")
	}
}

func TestStyleStartOffsetAccountsForRemovedScript(t *testing.T) {
	// Given
	input := "<script>\nx := 1\n</script>\n<style>\n.a{}\n</style>\n<span>Hi</span>"

	// When
	got, _ := Parse(input)

	// Then: offset must point at the style content within the original input
	if got.StyleStart == -1 {
		t.Fatalf("StyleStart = -1, want a real offset")
	}
	if input[got.StyleStart:got.StyleStart+len(got.Style)] != got.Style {
		t.Errorf("StyleStart %d does not point at Style content in original input", got.StyleStart)
	}
}

func TestTemplateStartOffsetInOriginal(t *testing.T) {
	// Given
	input := "<script>\nx := 1\n</script>\n\n<span>Hi</span>\n"

	// When
	got, _ := Parse(input)

	// Then: offset must point at the trimmed template within the original input
	if got.TemplateStart == -1 {
		t.Fatalf("TemplateStart = -1, want a real offset")
	}
	if input[got.TemplateStart:got.TemplateStart+len(got.Template)] != got.Template {
		t.Errorf("TemplateStart %d does not point at Template content in original input", got.TemplateStart)
	}
}

func TestAllSectionOffsetsPointAtContent(t *testing.T) {
	// Given
	input := "<sumi:imports>\n\"x\"\n</sumi:imports>\n<script>\ny := 2\n</script>\n<style>\n.b{}\n</style>\n<div>Body</div>"

	// When
	got, _ := Parse(input)

	// Then
	if input[got.ScriptStart:got.ScriptStart+len(got.Script)] != got.Script {
		t.Errorf("ScriptStart %d wrong", got.ScriptStart)
	}
	if input[got.StyleStart:got.StyleStart+len(got.Style)] != got.Style {
		t.Errorf("StyleStart %d wrong", got.StyleStart)
	}
	if input[got.TemplateStart:got.TemplateStart+len(got.Template)] != got.Template {
		t.Errorf("TemplateStart %d wrong", got.TemplateStart)
	}
}

func TestAbsentSectionsReportMinusOne(t *testing.T) {
	// Given
	input := "<span>Only a template</span>"

	// When
	got, _ := Parse(input)

	// Then
	if got.ScriptStart != -1 {
		t.Errorf("ScriptStart = %d, want -1", got.ScriptStart)
	}
	if got.StyleStart != -1 {
		t.Errorf("StyleStart = %d, want -1", got.StyleStart)
	}
	if got.TemplateStart != 0 {
		t.Errorf("TemplateStart = %d, want 0", got.TemplateStart)
	}
}

func TestEmptyInputTemplateStartMinusOne(t *testing.T) {
	// When
	got, _ := Parse("")

	// Then
	if got.TemplateStart != -1 {
		t.Errorf("TemplateStart = %d, want -1", got.TemplateStart)
	}
}
