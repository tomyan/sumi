package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateFileCreatesGoFile(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello, Sumi!</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}
	goFile := filepath.Join(dir, "hello_sumi.go")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Fatalf("expected %s to exist", goFile)
	}
}

func TestGenerateFileProducesValidGo(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello, Sumi!</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
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

func TestGenerateFileUsesDirectoryAsPackageName(t *testing.T) {
	// Given
	dir := t.TempDir()
	subdir := filepath.Join(dir, "myapp")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	sumiFile := filepath.Join(subdir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
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
	// Given
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

	// When
	err := generateDir(dir)

	// Then
	if err != nil {
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
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "app.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello</text>`), 0644); err != nil {
		t.Fatal(err)
	}
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)
	os.Chdir(dir)

	// When
	err = generateDir(".")

	// Then
	if err != nil {
		t.Fatalf("generateDir: %v", err)
	}
	goFile := filepath.Join(dir, "app_sumi.go")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", goFile)
	}
}

func TestGenerateFileReportsParseError(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "bad.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello`), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err == nil {
		t.Fatal("expected error for malformed .sumi file, got nil")
	}
}

func TestGenerateDirectoryWithNoSumiFilesIsNoOp(t *testing.T) {
	// Given
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("nothing"), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateDir(dir)

	// Then
	if err != nil {
		t.Fatalf("expected no error for dir with no .sumi files, got: %v", err)
	}
}

func TestGenerateFileUsesPackageMainWhenMainGoExists(t *testing.T) {
	// Given a directory containing main.go alongside a .sumi file
	dir := t.TempDir()
	subdir := filepath.Join(dir, "mywidget")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(subdir, "main.go")
	if err := os.WriteFile(mainGo, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sumiFile := filepath.Join(subdir, "hello.sumi")
	if err := os.WriteFile(sumiFile, []byte(`<text>Hello</text>`), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then the generated file should use package main, not package mywidget
	if err != nil {
		t.Fatalf("generateFile: %v", err)
	}
	goFile := filepath.Join(subdir, "hello_sumi.go")
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	code := string(src)
	if !contains(code, "package main") {
		t.Errorf("expected 'package main' when main.go exists, got:\n%s", code)
	}
}

func TestGenerateFileWithStyleBlock(t *testing.T) {
	// Given
	dir := t.TempDir()
	sumiFile := filepath.Join(dir, "styled.sumi")
	input := `<style>
.title {
  color: red;
  font-weight: bold;
}
</style>
<text class="title">Hello</text>`
	if err := os.WriteFile(sumiFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	// When
	err := generateFile(sumiFile)

	// Then
	if err != nil {
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
	if !contains(code, "MustParseStylesheet") {
		t.Errorf("expected embedded stylesheet in generated code:\n%s", code)
	}
	if !contains(code, "font-weight: bold") {
		t.Errorf("expected the rule in the embedded stylesheet:\n%s", code)
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
