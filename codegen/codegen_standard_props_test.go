package codegen

import (
	"regexp"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

// containsField reports whether the generated source assigns value to field,
// tolerating go/format struct-literal alignment padding.
func containsField(src, field, value string) bool {
	re := regexp.MustCompile(regexp.QuoteMeta(field) + `:\s+` + regexp.QuoteMeta(value))
	return re.MatchString(src)
}

// A1: layout attributes/properties use standard CSS names.

func generateBox(t *testing.T, attrs map[string]string) string {
	t.Helper()
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: attrs,
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, nil, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return string(out)
}

func TestGenerateFlexDirectionAttribute(t *testing.T) {
	// Given / When
	src := generateBox(t, map[string]string{"flex-direction": "row"})

	// Then
	if !containsField(src, "Direction", `"row"`) {
		t.Errorf("flex-direction should emit Direction:\n%s", src)
	}
}

func TestGenerateJustifyContentAttribute(t *testing.T) {
	src := generateBox(t, map[string]string{"justify-content": "space-between"})
	if !containsField(src, "Justify", `"space-between"`) {
		t.Errorf("justify-content should emit Justify:\n%s", src)
	}
}

func TestGenerateJustifyContentFlexStartNormalized(t *testing.T) {
	src := generateBox(t, map[string]string{"justify-content": "flex-end"})
	if !containsField(src, "Justify", `"end"`) {
		t.Errorf("flex-end should normalize to end:\n%s", src)
	}
}

func TestGenerateAlignItemsAttribute(t *testing.T) {
	src := generateBox(t, map[string]string{"align-items": "center"})
	if !containsField(src, "Align", `"center"`) {
		t.Errorf("align-items should emit Align:\n%s", src)
	}
}

func TestGenerateAlignItemsFlexStartNormalized(t *testing.T) {
	src := generateBox(t, map[string]string{"align-items": "flex-start"})
	if !containsField(src, "Align", `"start"`) {
		t.Errorf("flex-start should normalize to start:\n%s", src)
	}
}

// A1 clean break: legacy names are no longer consumed.
func TestLegacyLayoutAttributeNamesDropped(t *testing.T) {
	src := generateBox(t, map[string]string{
		"direction": "row",
		"justify":   "center",
		"align":     "center",
	})
	for _, field := range []string{"Direction", "Justify", "Align"} {
		if containsField(src, field, `"row"`) || containsField(src, field, `"center"`) {
			t.Errorf("legacy attribute must be ignored, found %s in:\n%s", field, src)
		}
	}
}

// A2: non-integer (pixel-derived) values on int attributes drop silently.
func TestPixelValuesOnIntAttributesDropSilently(t *testing.T) {
	src := generateBox(t, map[string]string{"width": "20px", "height": "1.5em"})
	if strings.Contains(src, "FixedWidth") || strings.Contains(src, "FixedHeight") {
		t.Errorf("pixel-derived lengths must be dropped:\n%s", src)
	}
}
