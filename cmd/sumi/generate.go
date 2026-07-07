package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomyan/sumi/gen"
	"github.com/tomyan/sumi/parser/template"
)

// generateFile compiles a single .sumi file to a _sumi.go file in the same directory.
func generateFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	out, err := gen.Generate(path, string(src))
	if err != nil {
		return err
	}
	outPath := gen.OutputPath(path)
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return fmt.Errorf("%s: %w", outPath, err)
	}
	return nil
}

// generateDir parses and generates all .sumi files in a directory.
func generateDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sumi") {
			continue
		}
		if err := generateFile(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// componentName derives a component name from a file path.
// Strips the extension and removes hyphens: "my-widget.sumi" -> "mywidget".
func componentName(path string) string {
	return template.ComponentName(path)
}

// exportedName capitalizes the first letter of a component name.
func exportedName(name string) string {
	return template.ExportedComponentName(name)
}
