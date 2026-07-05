package layout

import (
	"strconv"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/css"
)

// ResolveStyles resolves CSS for every node in the input tree against the
// component stylesheet, using the true runtime path and sibling context
// (so structural pseudo-classes work inside {if}/{for} content).
// Inline template attributes always win over CSS; CSS only sets a field
// when a matching declaration provides a value.
// Spliced child-component subtrees (identified by Tag "root") are skipped:
// they resolve against their own component's stylesheet.
func ResolveStyles(root *Input, ss *style.Stylesheet, viewportW, viewportH int) {
	if root == nil || ss == nil {
		return
	}
	css.SetViewport(viewportW, viewportH)
	rootEl := css.Element{Tag: "root"}
	props := css.Resolve(ss, []css.Element{rootEl})
	vars := customPropsFrom(nil, props)
	applyResolvedProps(root, expandProps(props, vars), nil, nil)
	resolveChildren(root, ss, []css.Element{rootEl}, vars)
}

const (
	pseudoBeforeTag = "::before"
	pseudoAfterTag  = "::after"
)

func resolveChildren(parent *Input, ss *style.Stylesheet, path []css.Element, vars map[string]string) {
	synthesizePseudoElements(parent, ss, path, vars)
	siblings := elementSiblings(parent.Children)
	elemIdx := 0
	for _, child := range parent.Children {
		if child == nil || child.Tag == "" || child.Tag == "root" || child.Tag[0] == ':' {
			continue // placeholder, unidentified, component subtree, or synthetic
		}
		el := siblings[elemIdx]
		el.Siblings = siblings
		el.Index = elemIdx
		el.ContainerW, el.ContainerH = parent.LastW, parent.LastH
		elemIdx++

		p := make([]css.Element, len(path), len(path)+1)
		copy(p, path)
		p = append(p, el)

		props := css.Resolve(ss, p)
		childVars := customPropsFrom(vars, props)
		applyResolvedProps(child,
			expandProps(props, childVars),
			expandProps(css.ResolveHover(ss, p), childVars),
			expandProps(css.ResolveFocus(ss, p), childVars))
		resolveChildren(child, ss, p, childVars)
	}
}

// customPropsFrom merges --custom-property declarations into the inherited
// variable scope; returns the inherited map unchanged when a node declares
// none (copy-on-write).
func customPropsFrom(inherited, props map[string]string) map[string]string {
	var merged map[string]string
	for k, v := range props {
		if !strings.HasPrefix(k, "--") {
			continue
		}
		if merged == nil {
			merged = make(map[string]string, len(inherited)+1)
			for ik, iv := range inherited {
				merged[ik] = iv
			}
		}
		merged[k] = v
	}
	if merged == nil {
		return inherited
	}
	return merged
}

// expandProps substitutes var() references in every value; values that
// expand to nothing are dropped (graceful drop), and custom properties
// themselves are removed from the applied set.
func expandProps(props, vars map[string]string) map[string]string {
	if props == nil {
		return nil
	}
	out := make(map[string]string, len(props))
	for k, v := range props {
		if strings.HasPrefix(k, "--") {
			continue
		}
		if strings.Contains(v, "var(") {
			v = strings.TrimSpace(css.ExpandVarRefs(v, vars))
			if v == "" {
				continue
			}
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// synthesizePseudoElements rebuilds ::before/::after marker children on a
// box from the current rules. Synthetic nodes (Tag "::before"/"::after")
// are stripped first so repeated resolution stays idempotent; pseudo boxes
// are invisible to sibling matching per spec.
func synthesizePseudoElements(parent *Input, ss *style.Stylesheet, path []css.Element, vars map[string]string) {
	if parent.Kind != KindBox {
		return
	}
	children := parent.Children[:0:0]
	for _, c := range parent.Children {
		if c != nil && (c.Tag == pseudoBeforeTag || c.Tag == pseudoAfterTag) {
			continue
		}
		children = append(children, c)
	}
	if before := pseudoNode(parent, ss, path, vars, "before", pseudoBeforeTag); before != nil {
		children = append([]*Input{before}, children...)
	}
	if after := pseudoNode(parent, ss, path, vars, "after", pseudoAfterTag); after != nil {
		children = append(children, after)
	}
	parent.Children = children
}

func pseudoNode(parent *Input, ss *style.Stylesheet, path []css.Element, vars map[string]string, name, tag string) *Input {
	props := expandProps(css.ResolvePseudoElement(ss, path, name), vars)
	if props == nil {
		return nil
	}
	content, ok := css.ParseContent(props["content"], parent.Attrs)
	if !ok {
		return nil
	}
	return &Input{
		Kind:    KindText,
		Tag:     tag,
		Content: content,
		Style:   css.ToRenderStyle(props),
	}
}

// elementSiblings builds the sibling identity list for structural matching.
func elementSiblings(children []*Input) []css.Element {
	var sibs []css.Element
	for _, c := range children {
		if c == nil || c.Tag == "" || c.Tag == "root" || c.Tag[0] == ':' {
			continue
		}
		sibs = append(sibs, css.Element{
			Tag: c.Tag, ID: c.ID, Classes: c.Classes, Attrs: c.Attrs,
			Empty: len(c.Children) == 0 && c.Content == "",
		})
	}
	return sibs
}

// applyResolvedProps writes resolved CSS onto a node. Visual styles are
// replaced wholesale (CSS is their only source); layout properties are set
// only when CSS declares them and no inline attribute overrides.
func applyResolvedProps(n *Input, props, hover, focus map[string]string) {
	n.Style = css.ToRenderStyle(props)
	if hover != nil {
		n.HoverStyle = css.ToRenderStyle(hover)
	}
	if focus != nil {
		n.FocusStyle = css.ToRenderStyle(focus)
	}
	n.Transitions = css.ParseTransitions(props)
	n.AnimationSpec = css.ParseAnimation(props)
	applyLayoutProps(n, props)
}

// cssValue returns the CSS value for a layout property, unless an inline
// attribute overrides it (inline attrs are emitted by codegen and win).
func cssValue(n *Input, props map[string]string, key string) (string, bool) {
	if _, inline := n.Attrs[key]; inline {
		return "", false
	}
	v, ok := props[key]
	return v, ok
}

func applyLayoutProps(n *Input, props map[string]string) {
	if v, ok := cssValue(n, props, "flex-direction"); ok {
		n.Direction = v
	}
	applySizeProp(n, props, "width", &n.FixedWidth, &n.WidthPct, &n.WidthCalc)
	applySizeProp(n, props, "height", &n.FixedHeight, &n.HeightPct, &n.HeightCalc)
	applyIntProp(n, props, "gap", &n.Gap)
	applyIntProp(n, props, "flex-grow", &n.FlexGrow)
	applyIntProp(n, props, "min-width", &n.MinWidth)
	applyIntProp(n, props, "min-height", &n.MinHeight)
	applyIntProp(n, props, "max-width", &n.MaxWidth)
	applyIntProp(n, props, "max-height", &n.MaxHeight)
	if v, ok := cssValue(n, props, "box-sizing"); ok {
		n.ContentBox = v == "content-box"
	}
	if v, ok := cssValue(n, props, "justify-content"); ok {
		n.Justify = normalizeFlexKeyword(v)
	}
	if v, ok := cssValue(n, props, "align-items"); ok {
		n.Align = normalizeFlexKeyword(v)
	}
	if v, ok := cssValue(n, props, "padding"); ok {
		n.Padding = ParsePadding(v)
	}
	applyMarginProps(n, props)
	if v, ok := cssValue(n, props, "display"); ok {
		n.Display = v
	}
	if v, ok := cssValue(n, props, "overflow"); ok {
		n.Overflow = v
	}
	applyPositionProps(n, props)
	applyBorderProps(n, props)
}

func applyMarginProps(n *Input, props map[string]string) {
	if v, ok := cssValue(n, props, "margin"); ok {
		n.Margin = ParseMargin(v)
	}
	side := func(key string, val *int, auto *bool) {
		v, ok := cssValue(n, props, key)
		if !ok {
			return
		}
		if v == "auto" {
			*auto = true
			return
		}
		*val = ParseCellLength(v)
	}
	side("margin-top", &n.Margin.Top, &n.Margin.AutoTop)
	side("margin-right", &n.Margin.Right, &n.Margin.AutoRight)
	side("margin-bottom", &n.Margin.Bottom, &n.Margin.AutoBottom)
	side("margin-left", &n.Margin.Left, &n.Margin.AutoLeft)
}

func applyPositionProps(n *Input, props map[string]string) {
	if v, ok := cssValue(n, props, "position"); ok {
		n.Position = v
	}
	applyIntProp(n, props, "top", &n.Top)
	applyIntProp(n, props, "left", &n.Left)
	applyIntProp(n, props, "right", &n.Right)
	applyIntProp(n, props, "bottom", &n.Bottom)
	applyIntProp(n, props, "z-index", &n.ZIndex)
}

func applyBorderProps(n *Input, props map[string]string) {
	if v, ok := cssValue(n, props, "border"); ok {
		n.Border = v
	}
	if v, ok := cssValue(n, props, "border-top"); ok {
		n.BorderTop = v
	}
	if v, ok := cssValue(n, props, "border-bottom"); ok {
		n.BorderBottom = v
	}
	if v, ok := cssValue(n, props, "border-title"); ok {
		n.BorderTitle = v
	}
	if v, ok := cssValue(n, props, "border-collapse"); ok {
		n.BorderCollapse = v == "collapse"
	}
}

func applyIntProp(n *Input, props map[string]string, key string, dst *int) {
	v, ok := cssValue(n, props, key)
	if !ok {
		return
	}
	if css.IsCalcValue(v) {
		if r, calcOK := css.EvalCalc(v, -1); calcOK {
			*dst = r
		}
		return
	}
	if strings.HasSuffix(v, "%") {
		return // no percentage meaning for this property
	}
	*dst = ParseCellLength(v)
}

func applySizeProp(n *Input, props map[string]string, key string, fixed, pct *int, calc *string) {
	v, ok := cssValue(n, props, key)
	if !ok {
		return
	}
	if css.IsCalcValue(v) {
		if strings.Contains(v, "%") {
			*calc = v // containing-block size arrives at layout time
			return
		}
		if r, calcOK := css.EvalCalc(v, -1); calcOK {
			*fixed = r
		}
		return
	}
	if p, isPct := strings.CutSuffix(v, "%"); isPct {
		if val, err := strconv.Atoi(p); err == nil {
			*pct = val
		}
		return
	}
	*fixed = ParseCellLength(v)
}

func normalizeFlexKeyword(v string) string {
	switch v {
	case "flex-start":
		return "start"
	case "flex-end":
		return "end"
	}
	return v
}
