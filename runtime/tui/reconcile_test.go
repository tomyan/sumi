package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestReconcilePreservesExistingInstances(t *testing.T) {
	// Given — a reconciler with initial items
	r := tui.NewReconciler[string]()
	createCount := 0
	create := func(key string) *tui.Component {
		createCount++
		return &tui.Component{
			Tree: &layout.Input{Kind: layout.KindText, Content: key},
		}
	}

	// When — first reconcile
	trees := r.Reconcile([]string{"a", "b", "c"}, create)
	if createCount != 3 {
		t.Fatalf("createCount = %d, want 3", createCount)
	}

	// When — reconcile with same keys in different order
	createCount = 0
	trees = r.Reconcile([]string{"c", "a", "b"}, create)

	// Then — no new creates (all reused)
	if createCount != 0 {
		t.Errorf("createCount = %d, want 0 (all reused)", createCount)
	}
	if len(trees) != 3 {
		t.Fatalf("len(trees) = %d, want 3", len(trees))
	}
	// Order should match new list
	if trees[0].Content != "c" {
		t.Errorf("trees[0] = %q, want c", trees[0].Content)
	}
}

func TestReconcileCreatesNewInstances(t *testing.T) {
	// Given
	r := tui.NewReconciler[string]()
	create := func(key string) *tui.Component {
		return &tui.Component{
			Tree: &layout.Input{Kind: layout.KindText, Content: key},
		}
	}
	r.Reconcile([]string{"a", "b"}, create)

	// When — add a new item
	trees := r.Reconcile([]string{"a", "b", "c"}, create)

	// Then
	if len(trees) != 3 {
		t.Fatalf("len(trees) = %d, want 3", len(trees))
	}
	if trees[2].Content != "c" {
		t.Errorf("trees[2] = %q, want c", trees[2].Content)
	}
}

func TestReconcileDisposesRemovedInstances(t *testing.T) {
	// Given
	r := tui.NewReconciler[string]()
	disposed := map[string]bool{}
	create := func(key string) *tui.Component {
		k := key // capture
		return &tui.Component{
			Tree:    &layout.Input{Kind: layout.KindText, Content: key},
			Dispose: func() { disposed[k] = true },
		}
	}
	r.Reconcile([]string{"a", "b", "c"}, create)

	// When — remove "b"
	r.Reconcile([]string{"a", "c"}, create)

	// Then
	if !disposed["b"] {
		t.Error("b should have been disposed")
	}
	if disposed["a"] || disposed["c"] {
		t.Error("a and c should not have been disposed")
	}
}

func TestReconcileEmptyList(t *testing.T) {
	// Given
	r := tui.NewReconciler[string]()
	disposed := map[string]bool{}
	create := func(key string) *tui.Component {
		k := key
		return &tui.Component{
			Tree:    &layout.Input{Kind: layout.KindText, Content: key},
			Dispose: func() { disposed[k] = true },
		}
	}
	r.Reconcile([]string{"a", "b"}, create)

	// When — empty list
	trees := r.Reconcile([]string{}, create)

	// Then
	if len(trees) != 0 {
		t.Errorf("len(trees) = %d, want 0", len(trees))
	}
	if !disposed["a"] || !disposed["b"] {
		t.Error("all should have been disposed")
	}
}
