package template

import "testing"

// B4a: mixed content — loose text interleaved with element children.
// Whitespace-only gaps follow the JSX newline rule: gaps containing a
// newline are dropped (source formatting); single-line gaps become one
// space text node.

func textParts(t *testing.T, node Node) []Part {
	t.Helper()
	text, ok := node.(*TextElement)
	if !ok {
		t.Fatalf("expected TextElement, got %T", node)
	}
	if text.Tag != "" {
		t.Fatalf("loose text should be tagless, got %q", text.Tag)
	}
	return text.Parts
}

func stringValue(t *testing.T, p Part) string {
	t.Helper()
	s, ok := p.(*StringPart)
	if !ok {
		t.Fatalf("expected StringPart, got %T", p)
	}
	return s.Value
}

func TestParseMixedTextAndInlineElement(t *testing.T) {
	// Given / When
	doc, err := Parse(`<p>hello <strong>bold</strong> tail</p>`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	p := doc.Children[0].(*BoxElement)
	if len(p.Children) != 3 {
		t.Fatalf("children = %d, want 3: %+v", len(p.Children), p.Children)
	}
	if got := stringValue(t, textParts(t, p.Children[0])[0]); got != "hello " {
		t.Errorf("leading text = %q, want %q", got, "hello ")
	}
	strong, ok := p.Children[1].(*TextElement)
	if !ok || strong.Tag != "strong" {
		t.Fatalf("expected strong element, got %+v", p.Children[1])
	}
	if got := stringValue(t, textParts(t, p.Children[2])[0]); got != " tail" {
		t.Errorf("trailing text = %q, want %q", got, " tail")
	}
}

func TestParseMixedTextWithExpressions(t *testing.T) {
	// Given / When
	doc, err := Parse(`<p>count: {count} of {total} <em>left</em></p>`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	p := doc.Children[0].(*BoxElement)
	if len(p.Children) != 2 {
		t.Fatalf("children = %d, want 2: %+v", len(p.Children), p.Children)
	}
	parts := textParts(t, p.Children[0])
	if len(parts) != 5 {
		t.Fatalf("parts = %d, want 5: %+v", len(parts), parts)
	}
	if expr, ok := parts[1].(*ExprPart); !ok || expr.Expr != "count" {
		t.Errorf("parts[1] = %+v, want expr count", parts[1])
	}
	if expr, ok := parts[3].(*ExprPart); !ok || expr.Expr != "total" {
		t.Errorf("parts[3] = %+v, want expr total", parts[3])
	}
	em, ok := p.Children[1].(*TextElement)
	if !ok || em.Tag != "em" {
		t.Fatalf("expected em element, got %+v", p.Children[1])
	}
}

func TestParseNewlineGapsDropped(t *testing.T) {
	// Given: elements separated by formatting whitespace with newlines.
	doc, err := Parse("<div>\n\t<div>x</div>\n\t<div>y</div>\n</div>")

	// Then: no whitespace children — parses as before B4a.
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	div := doc.Children[0].(*BoxElement)
	if len(div.Children) != 2 {
		t.Fatalf("children = %d, want 2: %+v", len(div.Children), div.Children)
	}
}

func TestParseSingleLineGapBecomesSpace(t *testing.T) {
	// Given: two inline elements separated by a same-line space.
	doc, err := Parse(`<p><strong>a</strong> <em>b</em></p>`)

	// Then: the gap is kept as a space text node.
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	p := doc.Children[0].(*BoxElement)
	if len(p.Children) != 3 {
		t.Fatalf("children = %d, want 3: %+v", len(p.Children), p.Children)
	}
	if got := stringValue(t, textParts(t, p.Children[1])[0]); got != " " {
		t.Errorf("gap = %q, want single space", got)
	}
}

func TestParseMixedTextBeforeControlFlow(t *testing.T) {
	// Given / When: loose text as a sibling of an {if} block.
	doc, err := Parse(`<div>status: {if ok}<span>good</span>{/if}</div>`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	div := doc.Children[0].(*BoxElement)
	if len(div.Children) != 2 {
		t.Fatalf("children = %d, want 2: %+v", len(div.Children), div.Children)
	}
	if got := stringValue(t, textParts(t, div.Children[0])[0]); got != "status: " {
		t.Errorf("leading text = %q, want %q", got, "status: ")
	}
	if _, ok := div.Children[1].(*IfNode); !ok {
		t.Fatalf("expected IfNode, got %T", div.Children[1])
	}
}

func TestParseMixedTextInsideControlFlowBody(t *testing.T) {
	// Given / When: mixed text + element inside an {if} body.
	doc, err := Parse(`<div>{if ok}all <strong>good</strong>{/if}</div>`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	div := doc.Children[0].(*BoxElement)
	ifNode, ok := div.Children[0].(*IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", div.Children[0])
	}
	if len(ifNode.Then) != 2 {
		t.Fatalf("then = %d nodes, want 2: %+v", len(ifNode.Then), ifNode.Then)
	}
	if got := stringValue(t, textParts(t, ifNode.Then[0])[0]); got != "all " {
		t.Errorf("text = %q, want %q", got, "all ")
	}
	strong, ok := ifNode.Then[1].(*TextElement)
	if !ok || strong.Tag != "strong" {
		t.Fatalf("expected strong, got %+v", ifNode.Then[1])
	}
}
