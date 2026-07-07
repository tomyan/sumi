package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

// bindDoc wraps a single native control (with a bind attribute) in a root div.
func bindDoc(tag string, attrs map[string]string, children ...template.Node) *template.Document {
	return &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children: []template.Node{
					&template.BoxElement{Tag: tag, Attributes: attrs, Children: children},
				},
			},
		},
	}
}

const bindScript = `name := sumi.New("")

func handleKey(evt sumi.Event) {
    if evt.Kind == sumi.EventSignal { app.Quit(); return }
}`

func generateBind(t *testing.T, doc *template.Document) string {
	t.Helper()
	out, err := GenerateComponent(doc, bindScript, nil, ComponentOptions{
		PackageName:   "main",
		ComponentName: "App",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	return string(out)
}

func TestBindValueOnTextInput(t *testing.T) {
	// Given — a text input with bind:value
	doc := bindDoc("input", map[string]string{"type": "text", "bind:value": "{name}"})

	// When
	src := generateBind(t, doc)

	// Then — update half wires the input event to Set with the string value.
	if !strings.Contains(src, `"input": func(evt *sumi.DOMEvent) {`) {
		t.Errorf("expected input handler:\n%s", src)
	}
	if !strings.Contains(src, `name.Set(evt.Data["value"].(string))`) {
		t.Errorf("expected Set from value data:\n%s", src)
	}
	// And the display half re-projects the signal in the sync effect.
	if !strings.Contains(src, "sumi.BindInputValue(box0, name.Get())") {
		t.Errorf("expected BindInputValue sync:\n%s", src)
	}
	// And the bind attribute never leaks into the Attrs map.
	if strings.Contains(src, "bind:value") {
		t.Errorf("bind attr must not appear in generated code:\n%s", src)
	}
}

func TestBindCheckedOnCheckbox(t *testing.T) {
	// Given — a checkbox with bind:checked
	doc := bindDoc("input", map[string]string{"type": "checkbox", "bind:checked": "{name}"})

	// When
	src := generateBind(t, doc)

	// Then — the change event carries the bool checked value.
	if !strings.Contains(src, `"change": func(evt *sumi.DOMEvent) {`) {
		t.Errorf("expected change handler:\n%s", src)
	}
	if !strings.Contains(src, `name.Set(evt.Data["checked"].(bool))`) {
		t.Errorf("expected Set from checked data:\n%s", src)
	}
	if !strings.Contains(src, "sumi.BindChecked(box0, name.Get())") {
		t.Errorf("expected BindChecked sync:\n%s", src)
	}
}

func TestBindValueOnSelect(t *testing.T) {
	// Given — a select with bind:value
	doc := bindDoc("select", map[string]string{"bind:value": "{name}"},
		&template.TextElement{Tag: "option", Attributes: map[string]string{"value": "a"},
			Parts: []template.Part{&template.StringPart{Value: "A"}}},
	)

	// When
	src := generateBind(t, doc)

	// Then — select uses the change event, not input.
	if !strings.Contains(src, `"change": func(evt *sumi.DOMEvent) {`) {
		t.Errorf("expected change handler for select:\n%s", src)
	}
	if !strings.Contains(src, `name.Set(evt.Data["value"].(string))`) {
		t.Errorf("expected Set from value data:\n%s", src)
	}
	if !strings.Contains(src, "sumi.BindSelectValue(box0, name.Get())") {
		t.Errorf("expected BindSelectValue sync:\n%s", src)
	}
}

func TestBindValueConflictingHandlerIsError(t *testing.T) {
	// Given — an input with both bind:value and an oninput handler
	doc := bindDoc("input", map[string]string{
		"bind:value": "{name}",
		"oninput":    "{handleKey}",
	})

	// When
	_, err := GenerateComponent(doc, bindScript, nil, ComponentOptions{
		PackageName: "app", ComponentName: "app",
	})

	// Then — the clash is a generation error, not a silent drop.
	if err == nil {
		t.Fatal("expected error for oninput alongside bind:value")
	}
	if !strings.Contains(err.Error(), "bind:value") || !strings.Contains(err.Error(), "oninput") {
		t.Errorf("error should name both sides: %v", err)
	}
}
