package style

import (
	"fmt"
	"sort"
	"strings"
)

// Serialize renders a stylesheet back to CSS text that Parse round-trips.
// Used by codegen to embed the component stylesheet in generated Go for
// runtime resolution. Adjacent rules sharing a media query regroup into one
// @media block; property order is normalised alphabetically.
func Serialize(ss *Stylesheet) string {
	var b strings.Builder
	for i := 0; i < len(ss.Rules); {
		media := ss.Rules[i].Media
		j := i
		for j < len(ss.Rules) && ss.Rules[j].Media == media {
			j++
		}
		indent := ""
		if media != "" {
			fmt.Fprintf(&b, "@media %s {\n", media)
			indent = "\t"
		}
		for _, rule := range ss.Rules[i:j] {
			serializeRule(&b, indent, rule)
		}
		if media != "" {
			b.WriteString("}\n")
		}
		i = j
	}
	for _, kf := range ss.Keyframes {
		serializeKeyframe(&b, kf)
	}
	return b.String()
}

func serializeRule(b *strings.Builder, indent string, rule Rule) {
	selector := rule.Selector
	if rule.Pseudo != "" {
		selector += ":" + rule.Pseudo
	}
	fmt.Fprintf(b, "%s%s {\n", indent, selector)
	serializeProperties(b, indent+"\t", rule.Properties)
	fmt.Fprintf(b, "%s}\n", indent)
}

func serializeProperties(b *strings.Builder, indent string, props map[string]string) {
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(b, "%s%s: %s;\n", indent, name, props[name])
	}
}

func serializeKeyframe(b *strings.Builder, kf Keyframe) {
	fmt.Fprintf(b, "@keyframes %s {\n", kf.Name)
	for _, stop := range kf.Stops {
		fmt.Fprintf(b, "\t%g%% {\n", stop.Percent*100)
		serializeProperties(b, "\t\t", stop.Properties)
		b.WriteString("\t}\n")
	}
	b.WriteString("}\n")
}
