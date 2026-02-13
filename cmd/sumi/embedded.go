package main

import (
	"fmt"
	"strings"

	"github.com/tomyan/sumi/components"
	"github.com/tomyan/sumi/parser/template"
)

// readEmbeddedComponent reads a built-in .sumi component source by tag name.
// Fundamental components (e.g., "placeholder") are read from components/<name>.sumi.
// Stdlib components (e.g., "sumi:TextInput") are read from components/sumi/<file>.sumi.
func readEmbeddedComponent(tagName string) (string, error) {
	filename := template.TagComponentFile(tagName)
	var path string
	if prefix, _, ok := template.SplitPrefix(tagName); ok {
		path = prefix + "/" + filename
	} else {
		path = filename
	}
	data, err := components.FS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("embedded component %q not found: %w", tagName, err)
	}
	return string(data), nil
}

// listEmbeddedComponents returns the tag names of all embedded components.
func listEmbeddedComponents() []string {
	var names []string
	names = append(names, listComponentsInDir(".", "")...)
	names = append(names, listComponentsInDir("sumi", "sumi")...)
	return names
}

// listComponentsInDir lists .sumi files in an embedded directory with an optional prefix.
func listComponentsInDir(dir, prefix string) []string {
	entries, err := components.FS.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sumi") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".sumi")
		name = strings.ReplaceAll(name, "-", "")
		if prefix != "" {
			names = append(names, prefix+":"+name)
		} else {
			names = append(names, name)
		}
	}
	return names
}
