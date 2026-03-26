package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestComponentNameFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"counter.sumi", "counter"},
		{"app.sumi", "app"},
		{"/some/dir/counter.sumi", "counter"},
		{"my-widget.sumi", "mywidget"},
	}
	for _, tt := range tests {
		// Given a file path
		// When extracting the component name
		got := componentName(tt.path)

		// Then it returns the expected lowercase name
		if got != tt.want {
			t.Errorf("componentName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestExportedName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"counter", "Counter"},
		{"app", "App"},
		{"mywidget", "Mywidget"},
	}
	for _, tt := range tests {
		// Given a component name
		// When converting to exported form
		got := exportedName(tt.name)

		// Then the first letter is capitalized
		if got != tt.want {
			t.Errorf("exportedName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestGenerateSingleFileStillWorks(t *testing.T) {
	// Given a single .sumi file with no component references
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	// When generating the single file directly (backward compat)
	err := generateFile(sumiFile)

	// Then it succeeds and produces valid Go
	if err != nil {
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
