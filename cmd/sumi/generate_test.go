package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateFileCreatesGoFile(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello, Sumi!</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	err := generateFile(sumiFile)
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}

	goFile := filepath.Join(dir, "hello_sumi.go")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Fatalf("expected %s to exist", goFile)
	}
}

func TestGenerateFileProducesValidGo(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello, Sumi!</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(sumiFile); err != nil {
		t.Fatalf("generateFile: %v", err)
	}

	goFile := filepath.Join(dir, "hello_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, goFile, src, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(src), parseErr)
	}
}

func TestGenerateFileUsesDirectoryAsPackageName(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory with a known name
	subdir := filepath.Join(dir, "myapp")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	sumiFile := filepath.Join(subdir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(sumiFile); err != nil {
		t.Fatalf("generateFile: %v", err)
	}

	goFile := filepath.Join(subdir, "hello_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	if got := string(src); !contains(got, "package myapp") {
		t.Errorf("expected 'package myapp' in output:\n%s", got)
	}
}

func TestGenerateDirectoryProcessesAllSumiFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"foo.sumi", "bar.sumi"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(`<text>Hello</text>`), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Also create a non-.sumi file that should be ignored
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not sumi"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateDir(dir); err != nil {
		t.Fatalf("generateDir: %v", err)
	}

	for _, name := range []string{"foo_sumi.go", "bar_sumi.go"} {
		goFile := filepath.Join(dir, name)
		if _, err := os.Stat(goFile); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", goFile)
		}
	}
}

func TestGenerateDirectoryDefaultsToCurrentDir(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "app.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore working directory
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)
	os.Chdir(dir)

	if err := generateDir("."); err != nil {
		t.Fatalf("generateDir: %v", err)
	}

	goFile := filepath.Join(dir, "app_sumi.go")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", goFile)
	}
}

func TestGenerateFileReportsParseError(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "bad.sumi")
	// Missing closing tag — template parser should error
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello`), 0644); err != nil {
		t.Fatal(err)
	}

	err := generateFile(sumiFile)
	if err == nil {
		t.Fatal("expected error for malformed .sumi file, got nil")
	}
}

func TestGenerateDirectoryWithNoSumiFilesIsNoOp(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("nothing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := generateDir(dir)
	if err != nil {
		t.Fatalf("expected no error for dir with no .sumi files, got: %v", err)
	}
}

func TestGenerateBoxWithBorderAndPadding(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "boxed.sumi")
	input := `<box border="single" padding="1">
  <text>Hello, Sumi!</text>
  <text>Box layout works!</text>
</box>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(sumiFile); err != nil {
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
	if !contains(code, "renderTree(") {
		t.Errorf("expected renderTree call in generated code:\n%s", code)
	}
}

func TestGenerateNestedBoxes(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "nested.sumi")
	input := `<box><box border="single"><text>Nested</text></box></box>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(sumiFile); err != nil {
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
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "sized.sumi")
	input := `<box direction="column" width="40" height="10">
  <text>Sized box</text>
</box>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(sumiFile); err != nil {
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

	if err := generateFile(sumiFile); err != nil {
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

func TestGenerateFileWithStyleBlock(t *testing.T) {
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "styled.sumi")
	input := `<style>
.title {
  color: red;
  bold: true;
}
</style>
<text class="title">Hello</text>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(sumiFile); err != nil {
		t.Fatalf("generateFile: %v", err)
	}

	goFile := filepath.Join(dir, "styled_sumi.go")
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
	if !contains(code, "render.Style{") {
		t.Errorf("expected render.Style literal in generated code:\n%s", code)
	}
	if !contains(code, "Bold: true") {
		t.Errorf("expected Bold: true in generated code:\n%s", code)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
