package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

// C3a: on<type>={expr} attributes emit an On handler map; the legacy
// OnClick field is gone.
func TestGenerateOnClickEmitsHandlerMap(t *testing.T) {
	// Given — a zero-arg handler wired via onclick
	scriptSrc := `count := sumi.New(0)

func increment() {
    count.Update(func(n int) int { return n + 1 })
}`
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Tag:        "div",
				Attributes: map[string]string{"onclick": "{increment}"},
				Children: []template.Node{
					&template.TextElement{Tag: "span", Parts: []template.Part{&template.ExprPart{Expr: "count"}}},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "clicker",
		ComponentName: "Clicker",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, "OnClick") {
		t.Errorf("legacy OnClick field must not be emitted:\n%s", src)
	}
	if !strings.Contains(src, `"click": func(evt *sumi.DOMEvent)`) {
		t.Errorf("expected wrapped click handler in On map:\n%s", src)
	}
	if !strings.Contains(src, "if h := (increment); h != nil") {
		t.Errorf("expected nil-checked zero-arg wrapper:\n%s", src)
	}
}

// C3a: handlers declared with parameters receive the DOM event directly.
func TestGenerateEventAwareClickHandlerEmitsDirectReference(t *testing.T) {
	// Given
	scriptSrc := `count := sumi.New(0)

func handleClick(evt *sumi.DOMEvent) {
    count.Update(func(n int) int { return n + 1 })
    evt.StopPropagation()
}`
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Tag:        "div",
				Attributes: map[string]string{"onclick": "{handleClick}"},
				Children: []template.Node{
					&template.TextElement{Tag: "span", Parts: []template.Part{&template.ExprPart{Expr: "count"}}},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "clicker",
		ComponentName: "Clicker",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `"click": handleClick`) {
		t.Errorf("expected direct handler reference for event-aware func:\n%s", src)
	}
	if strings.Contains(src, "if h := (handleClick)") {
		t.Errorf("event-aware handler must not be wrapped:\n%s", src)
	}
}
