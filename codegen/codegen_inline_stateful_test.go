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

// assertValidGo checks that the output is valid Go source code.
func assertValidGo(t *testing.T, out []byte) {
	t.Helper()
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}
