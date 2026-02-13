package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// counterComponentInfo builds a full ComponentInfo for inlining (with Doc/Script).
func counterComponentInfo() *ComponentInfo {
	return &ComponentInfo{
		Name:         "counter",
		ExportedName: "Counter",
		Props:        []string{"label"},
		HasState:     true,
		Doc: &template.Document{
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
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "label", DefaultExpr: `"Count"`},
			},
			StateDecls: []script.StateDecl{
				{Name: "count", InitExpr: "0"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "increment", Params: "", Body: "\n\tcount = count + 1\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "count", Line: "count = count + 1"},
					},
				},
			},
		},
	}
}

func TestInlineStatefulChildState(t *testing.T) {
	// Given a parent with a stateful counter component
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
				Children: []template.Node{
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Clicks"},
					},
				},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterComponentInfo()},
	})

	// Then child state is emitted as a namespaced local
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0_count := 0") {
		t.Errorf("expected namespaced child state counter0_count:\n%s", src)
	}

	// No component struct constructor
	if strings.Contains(src, "NewCounterComponent") {
		t.Errorf("should not have component constructor:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineStatefulChildHandler(t *testing.T) {
	// Given a parent with a stateful counter component
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterComponentInfo()},
	})

	// Then child handler is emitted as a namespaced closure
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0_increment := func() {") {
		t.Errorf("expected namespaced closure counter0_increment:\n%s", src)
	}

	// Handler body should reference namespaced state
	if !strings.Contains(src, "counter0_count = counter0_count + 1") {
		t.Errorf("expected namespaced state in handler body:\n%s", src)
	}

	// Handler should set dirty flag via app
	if !strings.Contains(src, "app.Dirty = true") {
		t.Errorf("expected app.Dirty = true in handler:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineStatefulChildOnkeyInEventLoop(t *testing.T) {
	// Given a parent with a stateful counter component that has onkey
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterComponentInfo()},
	})

	// Then the event loop should call the namespaced handler directly
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0_increment()") {
		t.Errorf("expected inlined onkey call counter0_increment():\n%s", src)
	}

	// Should NOT dispatch via HandleKey
	if strings.Contains(src, ".HandleKey(") {
		t.Errorf("should not have HandleKey dispatch:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineStatefulNoDirtyPolling(t *testing.T) {
	// Given a parent with a stateful counter component
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterComponentInfo()},
	})

	// Then no Dirty() polling should exist — app.Dirty used instead
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, ".Dirty()") {
		t.Errorf("should not have Dirty() polling:\n%s", src)
	}

	// Dirty flag should use app.Dirty
	if !strings.Contains(src, "app.Dirty = true") {
		t.Errorf("expected app.Dirty = true in output:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineStatefulExpressionNodesNamespaced(t *testing.T) {
	// Given a parent with a stateful counter component
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterComponentInfo()},
	})

	// Then expression nodes should be namespaced and reference namespaced vars
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Prop should be resolved, state var should be namespaced
	// "{label}: {count}" with label="Clicks" → fmt.Sprintf("Clicks: %v", counter0_count)
	if !strings.Contains(src, "counter0_count") {
		t.Errorf("expected namespaced counter0_count in expression:\n%s", src)
	}
	// Extracted node should have namespaced name
	if !strings.Contains(src, "counter0_node0") {
		t.Errorf("expected namespaced counter0_node0 extracted variable:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineStatefulMultipleInstances(t *testing.T) {
	// Given a parent with two counter components
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
				Children: []template.Node{
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Clicks"},
					},
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Score"},
					},
				},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterComponentInfo()},
	})

	// Then each instance gets unique namespaced state and handlers
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// First instance
	if !strings.Contains(src, "counter0_count := 0") {
		t.Errorf("expected counter0_count state:\n%s", src)
	}
	if !strings.Contains(src, "counter0_increment := func() {") {
		t.Errorf("expected counter0_increment closure:\n%s", src)
	}

	// Second instance
	if !strings.Contains(src, "counter1_count := 0") {
		t.Errorf("expected counter1_count state:\n%s", src)
	}
	if !strings.Contains(src, "counter1_increment := func() {") {
		t.Errorf("expected counter1_increment closure:\n%s", src)
	}

	// Both onkey handlers called in event loop
	if !strings.Contains(src, "counter0_increment()") {
		t.Errorf("expected counter0_increment() call:\n%s", src)
	}
	if !strings.Contains(src, "counter1_increment()") {
		t.Errorf("expected counter1_increment() call:\n%s", src)
	}

	// Props resolved differently
	if !strings.Contains(src, "counter0_node0") {
		t.Errorf("expected counter0_node0:\n%s", src)
	}
	if !strings.Contains(src, "counter1_node0") {
		t.Errorf("expected counter1_node0:\n%s", src)
	}

	assertValidGo(t, out)
}

// eventAwareCounterComponentInfo builds a ComponentInfo with an event-aware handler.
func eventAwareCounterComponentInfo() *ComponentInfo {
	return &ComponentInfo{
		Name:         "counter",
		ExportedName: "Counter",
		Props:        []string{"label"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{"onkey": "handleKey"},
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
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "label", DefaultExpr: `"Count"`},
			},
			StateDecls: []script.StateDecl{
				{Name: "count", InitExpr: "0"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "handleKey", Params: "evt input.Event", Body: "\n\tcount = count + 1\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "count", Line: "count = count + 1"},
					},
				},
			},
		},
	}
}

func TestInlineStatefulEventAwareClosure(t *testing.T) {
	// Given a parent with an event-aware inlined component
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": eventAwareCounterComponentInfo()},
	})

	// Then child handler should be event-aware
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Namespaced closure with event parameter
	if !strings.Contains(src, "counter0_handleKey := func(evt input.Event)") {
		t.Errorf("expected parameterized inlined closure:\n%s", src)
	}

	// Event loop should call with evt, not as zero-arg
	if !strings.Contains(src, "counter0_handleKey(evt)") {
		t.Errorf("expected counter0_handleKey(evt) call:\n%s", src)
	}

	// No auto-quit since handler is event-aware
	if strings.Contains(src, "evt.Ctrl && evt.Rune == 'c'") {
		t.Errorf("event-aware inlined handler should not have auto Ctrl+C quit:\n%s", src)
	}

	assertValidGo(t, out)
}

// inputComponentWithProp builds a ComponentInfo with a prop used as a bare variable in a function body.
func inputComponentWithProp() *ComponentInfo {
	return &ComponentInfo{
		Name:         "textinput",
		ExportedName: "Textinput",
		Props:        []string{"placeholder"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{"onkey": "handleEvent"},
					Children: []template.Node{
						&template.TextElement{
							Parts: []template.Part{
								&template.ExprPart{Expr: "displayLine()"},
							},
						},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "placeholder", DefaultExpr: `""`},
			},
			StateDecls: []script.StateDecl{
				{Name: "value", InitExpr: `""`},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "displayLine", ReturnType: "string",
					Body: "\n\ttext := value\n\tif len(text) == 0 && len(placeholder) > 0 {\n\t\ttext = placeholder\n\t}\n\treturn text\n",
				},
				{
					Name: "handleEvent", Params: "evt input.Event",
					Body: "\n\tvalue = value + string(evt.Rune)\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "value", Line: "value = value + string(evt.Rune)"},
					},
				},
			},
		},
	}
}

func TestInlineLiteralPropDeclared(t *testing.T) {
	// Given a parent using a component with a literal prop value
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "textinput",
				Attributes: map[string]string{"placeholder": "Enter name..."},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"textinput": inputComponentWithProp()},
	})

	// Then the literal prop should be declared as a namespaced variable
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `textinput0_placeholder := "Enter name..."`) {
		t.Errorf("expected namespaced prop declaration:\n%s", src)
	}

	// Function body should reference namespaced prop variable
	if !strings.Contains(src, "textinput0_placeholder") {
		t.Errorf("expected namespaced placeholder reference in function body:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineLiteralPropDefault(t *testing.T) {
	// Given a parent using a component WITHOUT passing the prop (uses default)
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "textinput",
				Attributes: map[string]string{},
			},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"textinput": inputComponentWithProp()},
	})

	// Then the prop should be declared with its default value
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `textinput0_placeholder := ""`) {
		t.Errorf("expected default prop declaration:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestInlineExpressionPropMappedDirectly(t *testing.T) {
	// Given a parent using a component with an expression prop
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "textinput",
				Attributes: map[string]string{"placeholder": "{hint}"},
			},
		},
	}
	parentScript := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "hint", InitExpr: `"Type here"`},
		},
	}

	// When generating with inlinable component info
	out, err := Generate(doc, parentScript, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"textinput": inputComponentWithProp()},
	})

	// Then the expression prop should map to the parent variable (no separate declaration)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Function body should reference parent's variable directly
	if !strings.Contains(src, "len(hint)") {
		t.Errorf("expected parent variable 'hint' in function body:\n%s", src)
	}

	// Should NOT declare textinput0_placeholder for expression props
	if strings.Contains(src, "textinput0_placeholder") {
		t.Errorf("expression prop should not declare namespaced variable:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestRecursiveComponentInlining(t *testing.T) {
	// Given:
	// - Component "inner" has a prop and state
	// - Component "outer" contains <inner /> and has its own state
	// - Root doc uses <outer />
	innerInfo := &ComponentInfo{
		Name:         "inner",
		ExportedName: "Inner",
		Props:        []string{"label"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.TextElement{
					Parts: []template.Part{
						&template.ExprPart{Expr: "label"},
						&template.StringPart{Value: ": "},
						&template.ExprPart{Expr: "val"},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls:  []script.PropDecl{{Name: "label", DefaultExpr: `"X"`}},
			StateDecls: []script.StateDecl{{Name: "val", InitExpr: `0`}},
		},
	}

	outerInfo := &ComponentInfo{
		Name:         "outer",
		ExportedName: "Outer",
		Props:        []string{"title"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{},
					Children: []template.Node{
						&template.TextElement{
							Parts: []template.Part{
								&template.ExprPart{Expr: "title"},
							},
						},
						&template.ComponentElement{
							Name:       "inner",
							Attributes: map[string]string{"label": "nested"},
						},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls:  []script.PropDecl{{Name: "title", DefaultExpr: `"T"`}},
			StateDecls: []script.StateDecl{{Name: "count", InitExpr: `0`}},
		},
	}

	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "outer",
				Attributes: map[string]string{"title": "Main"},
			},
		},
	}

	// When generating with nested inlinable components
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components: map[string]*ComponentInfo{
			"inner": innerInfo,
			"outer": outerInfo,
		},
	})

	// Then both layers of state are namespaced
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Outer state: outer0_count
	if !strings.Contains(src, "outer0_count := 0") {
		t.Errorf("expected outer0_count state:\n%s", src)
	}

	// Inner state: outer0_inner0_val (nested under outer0)
	if !strings.Contains(src, "outer0_inner0_val := 0") {
		t.Errorf("expected outer0_inner0_val nested state:\n%s", src)
	}

	assertValidGo(t, out)
}

// assertValidGo checks that the output is valid Go source code.
func assertValidGo(t *testing.T, out []byte) {
	t.Helper()
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}
