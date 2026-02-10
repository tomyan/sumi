package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateBoxWithBorderAndPadding(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "boxed.sumi")
	input := `<box border="single" padding="1">
  <text>Hello, Sumi!</text>
  <text>Box layout works!</text>
</box>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}
	goFile := filepath.Join(dir, "boxed_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, goFile, src, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(src), parseErr)
	}
	code := string(src)
	if !contains(code, "layout.KindBox") {
		t.Errorf("expected layout.KindBox in generated code:\n%s", code)
	}
	if !contains(code, "Border:") || !contains(code, `"single"`) {
		t.Errorf("expected Border with \"single\" in generated code:\n%s", code)
	}
	if !contains(code, "layout.ParsePadding") {
		t.Errorf("expected layout.ParsePadding in generated code:\n%s", code)
	}
	if !contains(code, "layout.Layout(") {
		t.Errorf("expected layout.Layout call in generated code:\n%s", code)
	}
	if !contains(code, "layout.RenderTree(") {
		t.Errorf("expected layout.RenderTree call in generated code:\n%s", code)
	}
}

func TestGenerateNestedBoxes(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "nested.sumi")
	input := `<box><box border="single"><text>Nested</text></box></box>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}
	goFile := filepath.Join(dir, "nested_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, goFile, src, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(src), parseErr)
	}
	code := string(src)
	if !contains(code, `Content: "Nested"`) {
		t.Errorf("expected Content: \"Nested\" in generated code:\n%s", code)
	}
	if !contains(code, "Border:") || !contains(code, `"single"`) {
		t.Errorf("expected Border with \"single\" in generated code:\n%s", code)
	}
}

func TestGenerateBoxWithDirectionAndSize(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "sized.sumi")
	input := `<box direction="column" width="40" height="10">
  <text>Sized box</text>
</box>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}
	goFile := filepath.Join(dir, "sized_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, goFile, src, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(src), parseErr)
	}
	code := string(src)
	if !contains(code, `Direction: "column"`) {
		t.Errorf("expected Direction: \"column\" in generated code:\n%s", code)
	}
	if !contains(code, "FixedWidth") {
		t.Errorf("expected FixedWidth in generated code:\n%s", code)
	}
	if !contains(code, "FixedHeight") {
		t.Errorf("expected FixedHeight in generated code:\n%s", code)
	}
}

func TestGenerateMixedBoxesAndText(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "mixed.sumi")
	input := `<text>Top-level text</text>
<box border="single">
  <text>Inside box</text>
</box>
<text>Bottom text</text>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}
	goFile := filepath.Join(dir, "mixed_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, goFile, src, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(src), parseErr)
	}
	code := string(src)
	if !contains(code, `Content: "Top-level text"`) {
		t.Errorf("expected top-level text content in generated code:\n%s", code)
	}
	if !contains(code, `Content: "Inside box"`) {
		t.Errorf("expected inside box text content in generated code:\n%s", code)
	}
	if !contains(code, `Content: "Bottom text"`) {
		t.Errorf("expected bottom text content in generated code:\n%s", code)
	}
	if !contains(code, "layout.KindBox") {
		t.Errorf("expected layout.KindBox in generated code:\n%s", code)
	}
}
