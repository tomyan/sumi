package codegen

// instanceTracker is a stub for backward compatibility.
// The old component inlining system used this to track component instances.
// It's always nil in the new signal component model.
type instanceTracker struct{}

func newInstanceTracker(_ interface{}) *instanceTracker { return nil }

// componentInstance is a stub type for backward compatibility.
type componentInstance struct {
	Info    *ComponentInfo
	VarName string
	Attrs   map[string]string
}

// ComponentInfo describes a child component (legacy, used by old tests).
type ComponentInfo struct {
	Name         string
	ExportedName string
	Props        []string
	HasState     bool
	Doc          interface{}
	Script       interface{}
	Stylesheet   interface{}
}
