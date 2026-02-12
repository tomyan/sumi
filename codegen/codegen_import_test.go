package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestTimeImportWhenFuncUsesTime(t *testing.T) {
	// Given — a script with a function body that uses time.Now()
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{{
			Name:       "tick",
			Body:       "start := time.Now().UnixMilli()\n",
			ReturnType: "",
		}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `"time"`) {
		t.Errorf("expected \"time\" import in output:\n%s", src)
	}
}

func TestNoTimeImportWhenNotUsed(t *testing.T) {
	// Given — a script with no time usage
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{{
			Name: "increment",
			Body: "count = count + 1\n",
			StateAssignments: []script.StateAssignment{
				{VarName: "count", Line: "count = count + 1"},
			},
		}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, `"time"`) {
		t.Errorf("should not have \"time\" import in output:\n%s", src)
	}
}
