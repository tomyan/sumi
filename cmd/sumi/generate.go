package main

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomyan/sumi/codegen"
	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/section"
	"github.com/tomyan/sumi/parser/template"
)

// generateFile compiles a single .sumi file to a _sumi.go file in the same directory.
func generateFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	sections, err := section.Parse(string(src))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	doc, err := template.Parse(sections.Template)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	var sc *script.Script
	if sections.Script != "" {
		sc, err = script.Parse(sections.Script)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	pkgName := packageName(path)
	out, err := codegen.Generate(doc, sc, codegen.Options{PackageName: pkgName})
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	outPath := outputPath(path)
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

// generateDir finds all .sumi files in a directory and generates each one.
func generateDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sumi") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := generateFile(path); err != nil {
			return err
		}
	}

	return nil
}

// outputPath converts "foo.sumi" to "foo_sumi.go".
func outputPath(sumiPath string) string {
	dir := filepath.Dir(sumiPath)
	base := filepath.Base(sumiPath)
	name := strings.TrimSuffix(base, ".sumi")
	return filepath.Join(dir, name+"_sumi.go")
}

// packageName derives the Go package name from the directory containing the .sumi file.
// Falls back to "main" if the directory name isn't a valid Go identifier.
func packageName(sumiPath string) string {
	absPath, err := filepath.Abs(sumiPath)
	if err != nil {
		return "main"
	}
	dir := filepath.Dir(absPath)
	name := filepath.Base(dir)
	if !token.IsIdentifier(name) {
		return "main"
	}
	return name
}
