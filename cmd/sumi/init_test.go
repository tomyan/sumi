package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// G1: sumi init scaffolds a runnable app.

func TestInitScaffoldsRunnableApp(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a module in a temp dir")
	}

	// Given
	dir := filepath.Join(t.TempDir(), "myapp")
	sumiRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	// When
	if err := runInit(dir, "example.com/myapp", sumiRoot); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Then: the scaffold files exist.
	for _, f := range []string{"app.sumi", "main.go", "go.mod", "app_sumi.go"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
	mod, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(mod), "module example.com/myapp") {
		t.Errorf("go.mod = %s, want module path", mod)
	}
	if !strings.Contains(string(mod), "replace github.com/tomyan/sumi => "+sumiRoot) {
		t.Errorf("go.mod = %s, want replace to local checkout", mod)
	}

	// And: the app builds.
	build := exec.Command("go", "build", "./...")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("scaffold does not build: %v\n%s", err, out)
	}
}

func TestInitRefusesNonEmptyDir(t *testing.T) {
	// Given
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// When
	err := runInit(dir, "example.com/x", "")

	// Then
	if err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Errorf("err = %v, want not-empty refusal", err)
	}
}

func TestFindSumiCheckoutWalksUp(t *testing.T) {
	// Given: cwd inside this repo.
	// When
	root, err := findSumiCheckout(".")

	// Then: the checkout root containing our go.mod.
	if err != nil {
		t.Fatalf("findSumiCheckout: %v", err)
	}
	mod, _ := os.ReadFile(filepath.Join(root, "go.mod"))
	if !strings.Contains(string(mod), "module github.com/tomyan/sumi") {
		t.Errorf("found %s, not the sumi checkout", root)
	}
}
