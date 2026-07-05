package template

import "testing"

// C1a: HTML element tags parse as elements (additive; box/text still work).

func TestParseHTMLTextLikeElement(t *testing.T) {
	// Given / When
	doc, err := Parse(`<h1 class="title">Hello {name}</h1>`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h1, ok := doc.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("expected TextElement, got %T", doc.Children[0])
	}
	if h1.Tag != "h1" || h1.Attributes["class"] != "title" {
		t.Errorf("h1 = %+v", h1)
	}
	if len(h1.Parts) != 2 {
		t.Errorf("parts = %+v", h1.Parts)
	}
}

func TestParseHTMLContainerElement(t *testing.T) {
	doc, err := Parse(`<div class="panel"><span>inside</span></div>`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	div, ok := doc.Children[0].(*BoxElement)
	if !ok {
		t.Fatalf("expected BoxElement, got %T", doc.Children[0])
	}
	if div.Tag != "div" {
		t.Errorf("div tag = %q", div.Tag)
	}
	span, ok := div.Children[0].(*TextElement)
	if !ok || span.Tag != "span" {
		t.Fatalf("expected span TextElement child, got %+v", div.Children[0])
	}
}

func TestParseHTMLSelfClosing(t *testing.T) {
	doc, err := Parse(`<div><hr/><span>after</span></div>`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	div := doc.Children[0].(*BoxElement)
	hr, ok := div.Children[0].(*BoxElement)
	if !ok || hr.Tag != "hr" {
		t.Fatalf("expected hr element, got %+v", div.Children[0])
	}
}

func TestParseHTMLEmptyDivIsContainer(t *testing.T) {
	doc, err := Parse(`<div></div>`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if _, ok := doc.Children[0].(*BoxElement); !ok {
		t.Fatalf("empty div should be a container, got %T", doc.Children[0])
	}
}

func TestParseHTMLControlFlowBody(t *testing.T) {
	doc, err := Parse(`<ul>{for _, it := range items.Get()}<li>{it}</li>{/for}</ul>`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	ul, ok := doc.Children[0].(*BoxElement)
	if !ok || ul.Tag != "ul" {
		t.Fatalf("ul = %+v", doc.Children[0])
	}
	if _, ok := ul.Children[0].(*ForNode); !ok {
		t.Fatalf("expected ForNode child, got %T", ul.Children[0])
	}
}

func TestUppercaseStillComponent(t *testing.T) {
	doc, err := Parse(`<Counter label="x" />`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if _, ok := doc.Children[0].(*ComponentElement); !ok {
		t.Fatalf("uppercase tag should stay a component, got %T", doc.Children[0])
	}
}

func TestLegacyBoxTextRejected(t *testing.T) {
	if _, err := Parse(`<div><span>hi</span></div>`); err != nil {
		t.Fatalf("div/span must parse: %v", err)
	}
	if _, err := Parse(`<box></box>`); err == nil {
		t.Error("<box> must be rejected with a helpful error")
	}
	if _, err := Parse(`<span>ok</span><text>no</text>`); err == nil {
		t.Error("<text> must be rejected with a helpful error")
	}
}

// C4: button keeps box form so borders and padding work; the label
// becomes an implicit untagged text child.
func TestParseButtonIsContainerWithImplicitLabel(t *testing.T) {
	// Given / When
	doc, err := Parse(`<button onclick={save}>Save {n}</button>`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn, ok := doc.Children[0].(*BoxElement)
	if !ok {
		t.Fatalf("expected BoxElement, got %T", doc.Children[0])
	}
	if btn.Tag != "button" || btn.Attributes["onclick"] != "{save}" {
		t.Errorf("button = %+v", btn)
	}
	label, ok := btn.Children[0].(*TextElement)
	if !ok || label.Tag != "" {
		t.Fatalf("expected untagged implicit text child, got %+v", btn.Children[0])
	}
	if len(label.Parts) != 2 {
		t.Errorf("label parts = %+v", label.Parts)
	}
}
