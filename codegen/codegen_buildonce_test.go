package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestBuildOnceExtractsExpressionNode(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Count: "},
					&template.ExprPart{Expr: "count"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Expression text node should be extracted as a named variable
	if !strings.Contains(src, "node0 := &layout.Input{") {
		t.Errorf("expected extracted node0 variable:\n%s", src)
	}

	// Tree should be at function scope (root := before doRender :=)
	rootIdx := strings.Index(src, "root := &layout.Input{")
	doRenderIdx := strings.Index(src, "doRender := func()")
	if rootIdx < 0 || doRenderIdx < 0 || rootIdx >= doRenderIdx {
		t.Errorf("expected root tree before doRender closure:\n%s", src)
	}

	// Sync function should exist and patch Content (compare-before-assign for static docs)
	if !strings.Contains(src, "sync := func()") {
		t.Errorf("expected sync function:\n%s", src)
	}
	if !strings.Contains(src, `node0.Content = v`) {
		t.Errorf("expected sync to patch node0.Content:\n%s", src)
	}

	// doRender should call sync
	if !strings.Contains(src, "sync()") {
		t.Errorf("expected doRender to call sync():\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceStaticTextNotExtracted(t *testing.T) {
	// Given — static text (no expressions) should NOT be extracted
	doc := &template.Document{
		Children: []template.Node{
			textNode("Hello"),
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Count: "},
					&template.ExprPart{Expr: "count"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Only one node should be extracted (the expression one)
	if strings.Contains(src, "node1 :=") {
		t.Errorf("should not extract static text nodes:\n%s", src)
	}
	// Static text should remain inline in the tree
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected static text inline in tree:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceMultipleExpressionNodes(t *testing.T) {
	// Given — two expression text nodes should each get their own variable
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "X: "},
					&template.ExprPart{Expr: "x"},
				},
			},
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Y: "},
					&template.ExprPart{Expr: "y"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
			{Name: "y", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "node0 := &layout.Input{") {
		t.Errorf("expected node0:\n%s", src)
	}
	if !strings.Contains(src, "node1 := &layout.Input{") {
		t.Errorf("expected node1:\n%s", src)
	}

	// Both should have sync lines (compare-before-assign for static docs)
	if !strings.Contains(src, `node0.Content = v`) {
		t.Errorf("expected sync for node0:\n%s", src)
	}
	if !strings.Contains(src, `node1.Content = v`) {
		t.Errorf("expected sync for node1:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceDynamicChildrenInBoxSync(t *testing.T) {
	// Given — box with dynamic children ({if}) should have Children rebuilt in sync
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children: []template.Node{
					textNode("Title"),
					&template.IfNode{
						Condition: "showModal",
						Then:      []template.Node{textNode("Modal content")},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "showModal", InitExpr: "false"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Box with dynamic children should be extracted as a named variable
	if !strings.Contains(src, "box0 := &layout.Input{") {
		t.Errorf("expected extracted box0 variable:\n%s", src)
	}

	// Sync should rebuild box0.Children via IIFE
	if !strings.Contains(src, "box0.Children = func() []*layout.Input {") {
		t.Errorf("expected sync to rebuild box0.Children:\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceDynamicChildrenAtRoot(t *testing.T) {
	// Given — root has dynamic children ({if} at root level)
	doc := &template.Document{
		Children: []template.Node{
			textNode("Background"),
			&template.IfNode{
				Condition: "showModal",
				Then:      []template.Node{textNode("Modal")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "showModal", InitExpr: "false"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Sync should rebuild root.Children
	if !strings.Contains(src, "root.Children = func() []*layout.Input {") {
		t.Errorf("expected sync to rebuild root.Children:\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceForLoopInSync(t *testing.T) {
	// Given — {for} loop should also be rebuilt in sync
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Children: []template.Node{
					textNode("Title"),
					&template.ForNode{
						Clause:   "i, item := range items",
						Key:      "item",
						Children: []template.Node{textNode("item")},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "items", InitExpr: `[]string{"a", "b"}`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Box with for loop should be extracted
	if !strings.Contains(src, "box0 := &layout.Input{") {
		t.Errorf("expected extracted box0:\n%s", src)
	}

	// Sync should rebuild Children with for loop
	if !strings.Contains(src, "box0.Children = func() []*layout.Input {") {
		t.Errorf("expected sync to rebuild box0.Children:\n%s", src)
	}
	if !strings.Contains(src, "for i, item := range items") {
		t.Errorf("expected for loop in sync:\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceInlinesStatelessComponent(t *testing.T) {
	// Given — a stateless component (props only, no state)
	headerDoc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.ExprPart{Expr: "title"},
						},
					},
					&template.TextElement{
						Parts: []template.Part{
							&template.ExprPart{Expr: "subtitle"},
						},
					},
				},
			},
		},
	}
	headerScript := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "title", DefaultExpr: `"Default"`},
			{Name: "subtitle", DefaultExpr: `""`},
		},
	}

	// Root doc references the component
	rootDoc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "header",
				Attributes: map[string]string{"title": "My App", "subtitle": "v1.0"},
			},
		},
	}
	rootScript := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	opts := Options{
		PackageName: "main",
		Components: map[string]*ComponentInfo{
			"header": {
				Name:         "header",
				ExportedName: "Header",
				Props:        []string{"title", "subtitle"},
				Doc:          headerDoc,
				Script:       headerScript,
			},
		},
	}

	// When
	out, err := Generate(rootDoc, rootScript, nil, opts)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Props should be resolved to literal strings
	if !strings.Contains(src, `Content: "My App"`) {
		t.Errorf("expected title prop resolved to literal:\n%s", src)
	}
	if !strings.Contains(src, `Content: "v1.0"`) {
		t.Errorf("expected subtitle prop resolved to literal:\n%s", src)
	}

	// Should NOT have component struct references
	if strings.Contains(src, ".Layout()") {
		t.Errorf("should not have component.Layout() call:\n%s", src)
	}
	if strings.Contains(src, "NewHeaderComponent") {
		t.Errorf("should not have component constructor:\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceInlinedComponentWithMixedPropExpr(t *testing.T) {
	// Given — component text has "{label}: clicks" (prop + static mixed)
	childDoc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.ExprPart{Expr: "label"},
					&template.StringPart{Value: ": clicks"},
				},
			},
		},
	}
	childScript := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Count"`},
		},
	}

	rootDoc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "badge",
				Attributes: map[string]string{"label": "Score"},
			},
		},
	}
	rootScript := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}

	opts := Options{
		PackageName: "main",
		Components: map[string]*ComponentInfo{
			"badge": {
				Name:         "badge",
				ExportedName: "Badge",
				Props:        []string{"label"},
				Doc:          childDoc,
				Script:       childScript,
			},
		},
	}

	// When
	out, err := Generate(rootDoc, rootScript, nil, opts)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Mixed prop expression should resolve: "{label}: clicks" → "Score: clicks"
	if !strings.Contains(src, `Content: "Score: clicks"`) {
		t.Errorf("expected resolved mixed prop expression:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceTreeNotInsideDoRender(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then — tree should NOT be rebuilt inside doRender
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	doRenderIdx := strings.Index(src, "doRender := func()")
	if doRenderIdx < 0 {
		t.Fatalf("expected doRender closure:\n%s", src)
	}
	// Scope to just the Run() function (before CreateApp)
	createAppIdx := strings.Index(src, "func CreateApp(")
	doRenderBody := src[doRenderIdx:]
	if createAppIdx > doRenderIdx {
		doRenderBody = src[doRenderIdx:createAppIdx]
	}

	// The tree construction should NOT appear inside doRender
	if strings.Contains(doRenderBody, "root := &layout.Input{") {
		t.Errorf("tree should be built outside doRender, not inside:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}
