package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/template"
)

// tagVarName returns a Go-safe variable name base for a tag name.
// "counter" → "counter", "sumi:TextInput" → "textinput"
func tagVarName(tagName string) string {
	if _, local, ok := template.SplitPrefix(tagName); ok {
		return strings.ToLower(local)
	}
	return tagName
}

// componentInstance tracks a single component reference in the template.
type componentInstance struct {
	VarName string         // e.g. "counter0"
	Info    *ComponentInfo // component type info
	Attrs   map[string]string
}

// collectComponentInstances walks the document and assigns unique variable names.
// Recursively descends into inlined component templates to collect nested references.
// Nested component attrs are namespaced through parent stateMaps during collection
// so that all attrs reference root-scope variables by the time code generation begins.
func collectComponentInstances(doc *template.Document, components map[string]*ComponentInfo) []componentInstance {
	if len(components) == 0 {
		return nil
	}
	var instances []componentInstance
	counts := map[string]int{}
	walkNodes(doc.Children, components, counts, &instances, "", nil)
	return instances
}

// walkNodes recursively collects ComponentElement instances from a node list.
// parentStateMap is the enclosing component's stateMap (nil for root-level nodes).
func walkNodes(nodes []template.Node, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance, prefix string, parentStateMap map[string]string) {
	for _, node := range nodes {
		walkNode(node, components, counts, instances, prefix, parentStateMap)
	}
}

// walkNode collects ComponentElement instances from a single node.
func walkNode(node template.Node, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance, prefix string, parentStateMap map[string]string) {
	switch n := node.(type) {
	case *template.ComponentElement:
		addComponentInstance(n, components, counts, instances, prefix, parentStateMap)
	case *template.BoxElement:
		walkNodes(n.Children, components, counts, instances, prefix, parentStateMap)
	case *template.IfNode:
		walkNodes(n.Then, components, counts, instances, prefix, parentStateMap)
		walkNodes(n.Else, components, counts, instances, prefix, parentStateMap)
	case *template.ForNode:
		walkNodes(n.Children, components, counts, instances, prefix, parentStateMap)
	}
}

// addComponentInstance creates an instance entry if the component is registered.
// If parentStateMap is non-nil, the component's attrs are namespaced through it
// so that references to the parent's scope resolve to root-scope variables.
// If the component has a Doc (will be inlined), recursively walks its template
// to collect nested component references with prefixed variable names.
func addComponentInstance(n *template.ComponentElement, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance, prefix string, parentStateMap map[string]string) {
	key := template.TagRegistryKey(n.Name)
	info, ok := components[key]
	if !ok {
		return
	}
	varBase := tagVarName(n.Name)
	idx := counts[prefix+key]
	counts[prefix+key]++
	varName := fmt.Sprintf("%s%s%d", prefix, varBase, idx)

	attrs := n.Attributes
	if parentStateMap != nil {
		attrs = namespaceExprAttrs(attrs, parentStateMap)
	}

	inst := componentInstance{
		VarName: varName,
		Info:    info,
		Attrs:   attrs,
	}
	*instances = append(*instances, inst)

	// Recurse into the component's template to find nested component references.
	// Build this instance's stateMap so children's attrs get namespaced correctly.
	if info.Doc != nil {
		var childStateMap map[string]string
		if needsNameMap(info) {
			childStateMap = buildStateNameMap(&inst)
		}
		nestedPrefix := varName + "_"
		walkNodes(info.Doc.Children, components, counts, instances, nestedPrefix, childStateMap)
	}
}

// writeComponentInits writes component instantiation statements.
// Skips instances that will be inlined (have Doc available).
func writeComponentInits(buf *bytes.Buffer, instances []componentInstance) {
	wrote := false
	for _, inst := range instances {
		if inst.Info.Doc != nil {
			continue // inlined, no constructor needed
		}
		args := buildComponentArgs(inst)
		fmt.Fprintf(buf, "\t%s := New%sComponent(%s)\n", inst.VarName, inst.Info.ExportedName, args)
		wrote = true
	}
	if wrote {
		buf.WriteString("\n")
	}
}

// buildComponentArgs builds the argument list for a component constructor call.
func buildComponentArgs(inst componentInstance) string {
	args := make([]string, len(inst.Info.Props))
	for i, prop := range inst.Info.Props {
		val := inst.Attrs[prop]
		args[i] = fmt.Sprintf("%q", val)
	}
	return strings.Join(args, ", ")
}

// instanceTracker assigns variable names to ComponentElements in document order.
type instanceTracker struct {
	instances []componentInstance
	index     int
}

// newInstanceTracker creates a tracker for the given instances.
func newInstanceTracker(instances []componentInstance) *instanceTracker {
	return &instanceTracker{instances: instances}
}

// next returns the variable name for the next ComponentElement.
func (t *instanceTracker) next() string {
	if t == nil || t.index >= len(t.instances) {
		return ""
	}
	name := t.instances[t.index].VarName
	t.index++
	return name
}

// nextInstance returns the full componentInstance for the next ComponentElement.
func (t *instanceTracker) nextInstance() *componentInstance {
	if t == nil || t.index >= len(t.instances) {
		return nil
	}
	inst := &t.instances[t.index]
	t.index++
	return inst
}
