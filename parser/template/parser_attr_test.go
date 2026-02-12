package template

import "testing"

func TestParseExpressionAttribute(t *testing.T) {
	// Given — attribute with curly braces for expression
	input := `<box onkey={handleKey}><text>hello</text></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if got := box.Attributes["onkey"]; got != "{handleKey}" {
		t.Errorf("Attributes[\"onkey\"] = %q, want %q", got, "{handleKey}")
	}
}

func TestParseExpressionAttributeWithSpaces(t *testing.T) {
	// Given — expression with spaces inside curlies
	input := `<box width={w - 2}><text>hello</text></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if got := box.Attributes["width"]; got != "{w - 2}" {
		t.Errorf("Attributes[\"width\"] = %q, want %q", got, "{w - 2}")
	}
}

func TestParseShorthandAttribute(t *testing.T) {
	// Given — shorthand {name} is equivalent to name={name}
	input := `<counter {count} />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := doc.Children[0].(*ComponentElement)
	if got := comp.Attributes["count"]; got != "{count}" {
		t.Errorf("Attributes[\"count\"] = %q, want %q", got, "{count}")
	}
}

func TestParseMixedAttributeSyntax(t *testing.T) {
	// Given — mix of quoted, expression, and shorthand attributes
	input := `<counter label="Clicks" count={myCount} {onReset} />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := doc.Children[0].(*ComponentElement)
	if got := comp.Attributes["label"]; got != "Clicks" {
		t.Errorf("Attributes[\"label\"] = %q, want %q", got, "Clicks")
	}
	if got := comp.Attributes["count"]; got != "{myCount}" {
		t.Errorf("Attributes[\"count\"] = %q, want %q", got, "{myCount}")
	}
	if got := comp.Attributes["onReset"]; got != "{onReset}" {
		t.Errorf("Attributes[\"onReset\"] = %q, want %q", got, "{onReset}")
	}
}

func TestParseBoxExpressionAttribute(t *testing.T) {
	// Given — box element with expression attribute
	input := `<box cursor-x={cursor}><text>hi</text></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if got := box.Attributes["cursor-x"]; got != "{cursor}" {
		t.Errorf("Attributes[\"cursor-x\"] = %q, want %q", got, "{cursor}")
	}
}

func TestParseQuotedAttributeUnchanged(t *testing.T) {
	// Given — existing quoted syntax still works
	input := `<box direction="row"><text>hi</text></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if got := box.Attributes["direction"]; got != "row" {
		t.Errorf("Attributes[\"direction\"] = %q, want %q", got, "row")
	}
}
