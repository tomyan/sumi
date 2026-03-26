package tui

import "github.com/tomyan/sumi/runtime/layout"

// Reconciler caches component instances by key for efficient list rendering.
// When a list re-renders, existing instances are reused (preserving signal state),
// new keys create fresh instances, and removed keys dispose their instances.
type Reconciler[K comparable] struct {
	cache map[K]*Component
}

// NewReconciler creates a new keyed reconciler.
func NewReconciler[K comparable]() *Reconciler[K] {
	return &Reconciler[K]{cache: make(map[K]*Component)}
}

// Reconcile takes a list of keys and a create function.
// Returns layout trees in the order of the keys list.
// Reuses existing instances for known keys, creates new ones for unknown keys,
// and disposes instances for keys no longer in the list.
func (r *Reconciler[K]) Reconcile(keys []K, create func(K) *Component) []*layout.Input {
	newCache := make(map[K]*Component, len(keys))
	trees := make([]*layout.Input, len(keys))

	for i, key := range keys {
		if comp, ok := r.cache[key]; ok {
			// Reuse existing instance.
			newCache[key] = comp
			trees[i] = comp.Tree
		} else {
			// Create new instance.
			comp := create(key)
			newCache[key] = comp
			trees[i] = comp.Tree
		}
	}

	// Dispose removed instances.
	for key, comp := range r.cache {
		if _, kept := newCache[key]; !kept {
			if comp.Dispose != nil {
				comp.Dispose()
			}
		}
	}

	r.cache = newCache
	return trees
}
