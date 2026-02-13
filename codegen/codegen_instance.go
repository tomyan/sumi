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
func collectComponentInstances(doc *template.Document, components map[string]*ComponentInfo) []componentInstance {
	if len(components) == 0 {
		return nil
	}
	var instances []componentInstance
	counts := map[string]int{}
	walkNodes(doc.Children, components, counts, &instances, "")
	return instances
}

// walkNodes recursively collects ComponentElement instances from a node list.
func walkNodes(nodes []template.Node, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance, prefix string) {
	for _, node := range nodes {
		walkNode(node, components, counts, instances, prefix)
	}
}

// walkNode collects ComponentElement instances from a single node.
func walkNode(node template.Node, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance, prefix string) {
	switch n := node.(type) {
	case *template.ComponentElement:
		addComponentInstance(n, components, counts, instances, prefix)
	case *template.BoxElement:
		walkNodes(n.Children, components, counts, instances, prefix)
	case *template.IfNode:
		walkNodes(n.Then, components, counts, instances, prefix)
		walkNodes(n.Else, components, counts, instances, prefix)
	case *template.ForNode:
		walkNodes(n.Children, components, counts, instances, prefix)
	}
}

// addComponentInstance creates an instance entry if the component is registered.
// If the component has a Doc (will be inlined), recursively walks its template
// to collect nested component references with prefixed variable names.
func addComponentInstance(n *template.ComponentElement, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance, prefix string) {
	key := template.TagRegistryKey(n.Name)
	info, ok := components[key]
	if !ok {
		return
	}
	varBase := tagVarName(n.Name)
	idx := counts[prefix+key]
	counts[prefix+key]++
	varName := fmt.Sprintf("%s%s%d", prefix, varBase, idx)
	*instances = append(*instances, componentInstance{
		VarName: varName,
		Info:    info,
		Attrs:   n.Attributes,
	})

	// Recurse into the component's template to find nested component references
	if info.Doc != nil {
		nestedPrefix := varName + "_"
		walkNodes(info.Doc.Children, components, counts, instances, nestedPrefix)
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
