package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/template"
)

// componentInstance tracks a single component reference in the template.
type componentInstance struct {
	VarName string         // e.g. "counter0"
	Info    *ComponentInfo // component type info
	Attrs   map[string]string
}

// collectComponentInstances walks the document and assigns unique variable names.
func collectComponentInstances(doc *template.Document, components map[string]*ComponentInfo) []componentInstance {
	if len(components) == 0 {
		return nil
	}
	var instances []componentInstance
	counts := map[string]int{}
	walkNodes(doc.Children, components, counts, &instances)
	return instances
}

// walkNodes recursively collects ComponentElement instances from a node list.
func walkNodes(nodes []template.Node, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance) {
	for _, node := range nodes {
		walkNode(node, components, counts, instances)
	}
}

// walkNode collects ComponentElement instances from a single node.
func walkNode(node template.Node, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance) {
	switch n := node.(type) {
	case *template.ComponentElement:
		addComponentInstance(n, components, counts, instances)
	case *template.BoxElement:
		walkNodes(n.Children, components, counts, instances)
	case *template.IfNode:
		walkNodes(n.Then, components, counts, instances)
		walkNodes(n.Else, components, counts, instances)
	case *template.ForNode:
		walkNodes(n.Children, components, counts, instances)
	}
}

// addComponentInstance creates an instance entry if the component is registered.
func addComponentInstance(n *template.ComponentElement, components map[string]*ComponentInfo, counts map[string]int, instances *[]componentInstance) {
	info, ok := components[n.Name]
	if !ok {
		return
	}
	idx := counts[n.Name]
	counts[n.Name]++
	*instances = append(*instances, componentInstance{
		VarName: fmt.Sprintf("%s%d", n.Name, idx),
		Info:    info,
		Attrs:   n.Attributes,
	})
}

// writeComponentInits writes component instantiation statements.
func writeComponentInits(buf *bytes.Buffer, instances []componentInstance) {
	for _, inst := range instances {
		args := buildComponentArgs(inst)
		fmt.Fprintf(buf, "\t%s := New%sComponent(%s)\n", inst.VarName, inst.Info.ExportedName, args)
	}
	if len(instances) > 0 {
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
