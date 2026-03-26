package main

import (
	"testing"
)

func TestReadEmbeddedComponentPlaceholder(t *testing.T) {
	// Given a fundamental component name that exists in the embedded FS

	// When reading the embedded source
	src, err := readEmbeddedComponent("placeholder")

	// Then it returns the source content
	if err != nil {
		t.Fatalf("readEmbeddedComponent: %v", err)
	}
	if src == "" {
		t.Fatal("expected non-empty source")
	}
	if !contains(src, "var label") {
		t.Errorf("expected 'var label' in placeholder source, got: %s", src)
	}
}

func TestReadEmbeddedComponentStdlib(t *testing.T) {
	// Given a stdlib component name (sumi: prefix)
	// The file should be looked up in components/sumi/ subdirectory

	// When reading a non-existent stdlib component
	_, err := readEmbeddedComponent("sumi:NonExistent")

	// Then it returns an error
	if err == nil {
		t.Fatal("expected error for non-existent stdlib component")
	}
}

func TestReadEmbeddedComponentNotFound(t *testing.T) {
	// Given a component name that doesn't exist

	// When reading it
	_, err := readEmbeddedComponent("doesnotexist")

	// Then it returns an error
	if err == nil {
		t.Fatal("expected error for non-existent component")
	}
}

func TestListEmbeddedComponents(t *testing.T) {
	// When listing all embedded components
	names := listEmbeddedComponents()

	// Then it includes the placeholder
	found := false
	for _, name := range names {
		if name == "placeholder" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'placeholder' in embedded components, got: %v", names)
	}
}
