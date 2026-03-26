package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateComponentFunction(t *testing.T) {
	// Given — a counter component with signal state and event handler
	scriptSrc := `count := signal.New(0)

func handleKey(evt input.Event) {
    if evt.Kind == input.EventSignal { app.Quit(); return }
    if evt.Rune == 'q' { app.Quit(); return }
    if evt.Kind == input.EventKey {
        count.Update(func(n int) int { return n + 1 })
    }
}`
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Count: "},
							&template.ExprPart{Expr: "count"},
						},
					},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should have component constructor
	if !strings.Contains(src, "func NewCounter(") {
		t.Errorf("expected NewCounter function:\n%s", src)
	}

	// Should return *tui.Component
	if !strings.Contains(src, "*tui.Component") {
		t.Errorf("expected *tui.Component return:\n%s", src)
	}

	// Should use signal.New
	if !strings.Contains(src, "signal.New(0)") {
		t.Errorf("expected signal.New(0):\n%s", src)
	}

	// Template should auto-unwrap: count → count.Get()
	if !strings.Contains(src, "count.Get()") {
		t.Errorf("expected count.Get() in template expression:\n%s", src)
	}

	// Should have signal.Effect for sync
	if !strings.Contains(src, "signal.Effect") {
		t.Errorf("expected signal.Effect:\n%s", src)
	}

	// Should import signal package
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/signal"`) {
		t.Errorf("expected signal import:\n%s", src)
	}

	// Should be valid Go
	assertValidGo(t, out)
}

func TestGenerateComponentWithChildComponent(t *testing.T) {
	// Given — a parent that uses a child component
	scriptSrc := `func handleKey(evt input.Event) {
    if evt.Kind == input.EventSignal { app.Quit(); return }
    if evt.Rune == 'q' { app.Quit(); return }
}`
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Parent App"},
						},
					},
					// Component usage: <Greeting name="Sumi" />
					&template.ComponentElement{
						Name:       "Greeting",
						Attributes: map[string]string{"name": "Sumi"},
					},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "main",
		ComponentName: "App",
		Components: map[string]ComponentChildInfo{
			"Greeting": {
				ImportPath: "github.com/example/greeting",
				Package:    "greeting",
			},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should instantiate child component
	if !strings.Contains(src, "greeting.NewGreeting") {
		t.Errorf("expected greeting.NewGreeting call:\n%s", src)
	}

	// Should pass props
	if !strings.Contains(src, "greeting.GreetingProps") {
		t.Errorf("expected GreetingProps struct:\n%s", src)
	}

	// Should embed child tree
	if !strings.Contains(src, ".Tree") {
		t.Errorf("expected .Tree reference:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestGenerateComponentWithBindProp(t *testing.T) {
	// Given — a parent that binds a signal to a child's prop
	scriptSrc := `name := signal.New("")

func handleKey(evt input.Event) {
    if evt.Kind == input.EventSignal { app.Quit(); return }
}`
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Name: "},
							&template.ExprPart{Expr: "name"},
						},
					},
					&template.ComponentElement{
						Name: "Field",
						Attributes: map[string]string{
							"bind:value": "{name}",
						},
					},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "main",
		ComponentName: "App",
		Components: map[string]ComponentChildInfo{
			"Field": {
				ImportPath: "github.com/example/field",
				Package:    "field",
			},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// bind:value should pass the signal directly (not .Get())
	if !strings.Contains(src, "Value: name") {
		t.Errorf("expected Value: name (signal reference) in props:\n%s", src)
	}

	// Should NOT wrap with .Get() for bind props
	if strings.Contains(src, "Value: name.Get()") {
		t.Errorf("bind prop should pass signal reference, not .Get():\n%s", src)
	}

	assertValidGo(t, out)
}

func TestGenerateComponentWithSlotPlaceholder(t *testing.T) {
	// Given — a card component with a slot placeholder
	scriptSrc := ``
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "card"},
				Children: []template.Node{
					&template.SlotElement{
						Name: "children",
					},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "card",
		ComponentName: "Card",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Props should have Children field
	if !strings.Contains(src, "Children []*layout.Input") {
		t.Errorf("expected Children prop:\n%s", src)
	}

	// Tree should reference props.Children
	if !strings.Contains(src, "props.Children") {
		t.Errorf("expected props.Children reference:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestGenerateComponentWithProps(t *testing.T) {
	// Given — a component with props
	scriptSrc := `var label string = "Count"

count := signal.New(0)

func increment() {
    count.Update(func(n int) int { return n + 1 })
}`
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "increment"},
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.ExprPart{Expr: "label"},
							&template.StringPart{Value: ": "},
							&template.ExprPart{Expr: "count"},
						},
					},
				},
			},
		},
	}

	// When
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should have props struct
	if !strings.Contains(src, "type CounterProps struct") {
		t.Errorf("expected CounterProps struct:\n%s", src)
	}

	// Should have Label field with default
	if !strings.Contains(src, "Label string") {
		t.Errorf("expected Label field:\n%s", src)
	}

	// Constructor should take props
	if !strings.Contains(src, "func NewCounter(props CounterProps)") {
		t.Errorf("expected props parameter:\n%s", src)
	}

	// label should NOT be unwrapped with .Get() (it's a plain prop, not a signal)
	// count should be unwrapped
	if !strings.Contains(src, "count.Get()") {
		t.Errorf("expected count.Get():\n%s", src)
	}

	assertValidGo(t, out)
}
