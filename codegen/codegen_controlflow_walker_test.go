package codegen

import (
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestDocHasExprsInsideIfNode(t *testing.T) {
	// Given - expression inside an IfNode
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "x",
				Then: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.ExprPart{Expr: "count"},
						},
					},
				},
			},
		},
	}

	// When
	result := docHasExprs(doc)

	// Then
	if !result {
		t.Error("expected docHasExprs to return true for expression inside IfNode")
	}
}

func TestWalkNodeFindsComponentInsideIf(t *testing.T) {
	// Given
	components := map[string]*ComponentInfo{
		"counter": {ExportedName: "Counter", Props: nil},
	}
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "visible",
				Then: []template.Node{
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{},
					},
				},
			},
		},
	}

	// When
	instances := collectComponentInstances(doc, components)

	// Then
	if len(instances) != 1 {
		t.Fatalf("got %d instances, want 1", len(instances))
	}
	if instances[0].VarName != "counter0" {
		t.Errorf("VarName = %q, want %q", instances[0].VarName, "counter0")
	}
}
