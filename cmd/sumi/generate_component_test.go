package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
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

func TestGenerateDirWithTwoComponents(t *testing.T) {
	// Given a directory with counter.sumi (child) and app.sumi (parent)
	dir := t.TempDir()
	counterSrc := `<script>
label := $prop("Count")
count := $state(0)
func increment() {
    count = count + 1
}
</script>
<box onkey="increment">
    <text>{label}: {count}</text>
</box>`
	appSrc := `<box direction="column">
    <counter label="Clicks" />
</box>`
	writeTestFile(t, dir, "counter.sumi", counterSrc)
	writeTestFile(t, dir, "app.sumi", appSrc)

	// When generating the directory
	err := generateDir(dir)

	// Then only the root component generates a Go file (child is inlined)
	if err != nil {
		t.Fatalf("generateDir: %v", err)
	}
	assertValidGoFile(t, filepath.Join(dir, "app_sumi.go"))

	// Child component should NOT get its own file (it's inlined)
	counterFile := filepath.Join(dir, "counter_sumi.go")
	if _, err := os.Stat(counterFile); err == nil {
		t.Errorf("child component should not generate a separate Go file: %s", counterFile)
	}
}

func TestGenerateDirChildTemplateInlined(t *testing.T) {
	// Given a directory with a child component that has $prop
	dir := t.TempDir()
	counterSrc := `<script>
label := $prop("Count")
</script>
<box>
    <text>{label}: static text</text>
</box>`
	appSrc := `<box direction="column">
    <counter label="Clicks" />
</box>`
	writeTestFile(t, dir, "counter.sumi", counterSrc)
	writeTestFile(t, dir, "app.sumi", appSrc)

	// When generating the directory
	err := generateDir(dir)

	// Then the parent's generated code inlines the child template with resolved props
	if err != nil {
		t.Fatalf("generateDir: %v", err)
	}
	src := readTestFile(t, filepath.Join(dir, "app_sumi.go"))

	// Prop should be resolved to literal
	if !strings.Contains(src, `Content: "Clicks: static text"`) {
		t.Errorf("expected inlined prop resolved to literal:\n%s", src)
	}

	// Should NOT have component struct references
	if strings.Contains(src, "NewCounterComponent") {
		t.Errorf("should not have component constructor:\n%s", src)
	}
}

func TestGenerateDirUnknownComponentReturnsError(t *testing.T) {
	// Given a directory with a parent referencing an unknown component
	dir := t.TempDir()
	appSrc := `<box direction="column">
    <unknown label="X" />
</box>`
	writeTestFile(t, dir, "app.sumi", appSrc)

	// When generating the directory
	err := generateDir(dir)

	// Then an error is returned about the unknown component
	if err == nil {
		t.Fatal("expected error for unknown component reference, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected error mentioning 'unknown', got: %v", err)
	}
}

func TestGenerateDirWithScrollbarComponent(t *testing.T) {
	// Given a root app using the embedded scrollbar component
	dir := t.TempDir()
	appSrc := `<script>
offset := $state(0)
func handleEvent(evt input.Event) {
}
</script>
<box focusable="true" onkey="handleEvent">
    <scrollbar
        contentSize={100}
        viewSize={20}
        bind:offset={offset}
        direction="horizontal"
        visible={true}
    />
</box>`
	writeTestFile(t, dir, "app.sumi", appSrc)

	// When generating the directory
	err := generateDir(dir)

	// Then the app compiles with the scrollbar inlined
	if err != nil {
		t.Fatalf("generateDir: %v", err)
	}
	assertValidGoFile(t, filepath.Join(dir, "app_sumi.go"))
}

func TestGenerateDirWithTexteditComponent(t *testing.T) {
	// Given a root app using the embedded textedit component
	dir := t.TempDir()
	appSrc := `<script>
name := $state("")
func handleEvent(evt input.Event) {
}
</script>
<box focusable="true" onkey="handleEvent">
    <textedit bind:value={name} placeholder="Enter name" />
</box>`
	writeTestFile(t, dir, "app.sumi", appSrc)

	// When generating the directory
	err := generateDir(dir)

	// Then the app compiles with textedit inlined
	if err != nil {
		t.Fatalf("generateDir: %v", err)
	}
	src := readTestFile(t, filepath.Join(dir, "app_sumi.go"))
	assertValidGoFile(t, filepath.Join(dir, "app_sumi.go"))
	// The textedit's value should bind to parent's name variable
	if !strings.Contains(src, "name") {
		t.Errorf("expected parent variable 'name' in output:\n%s", src)
	}
}

func TestGenerateDirWithEmbeddedComponent(t *testing.T) {
	// Given a directory with a root component referencing an embedded fundamental component
	dir := t.TempDir()
	appSrc := `<box>
    <placeholder label="world" />
</box>`
	writeTestFile(t, dir, "app.sumi", appSrc)

	// When generating the directory
	err := generateDir(dir)

	// Then the root generates with the embedded component inlined
	if err != nil {
		t.Fatalf("generateDir: %v", err)
	}
	src := readTestFile(t, filepath.Join(dir, "app_sumi.go"))
	assertValidGoFile(t, filepath.Join(dir, "app_sumi.go"))
	// The placeholder component should be inlined — its text content resolved
	if !strings.Contains(src, `"world"`) {
		t.Errorf("expected embedded placeholder prop resolved in output:\n%s", src)
	}
}

// writeTestFile writes a file to the test directory.
func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// readTestFile reads a file and returns its content as a string.
func readTestFile(t *testing.T, path string) string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file %s: %v", path, err)
	}
	return string(src)
}

// assertValidGoFile reads a Go file and asserts it parses without errors.
func assertValidGoFile(t *testing.T, path string) {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, path, src, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code at %s is not valid Go:\n%s\n\nerror: %v", path, string(src), parseErr)
	}
}
