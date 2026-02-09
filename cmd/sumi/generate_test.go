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
