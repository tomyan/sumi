package layout

import "testing"

// E3b: keyframe stops resolve at cascade time against the animating
// node's context — var() uses the node's custom-property scope,
// light-dark() keeps its ColorPair for scheme-aware emission.

func TestKeyframeStopsResolveVarFromNodeScope(t *testing.T) {
	// Given: --glow set on the parent, referenced inside @keyframes.
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "div", Kind: KindBox, Classes: []string{"pulse"}, Children: []*Input{
			{Kind: KindText, Content: "x"},
		}},
	}}
	ss := sheet(t, `
root { --glow: red; }
.pulse { animation: pulse 1s; }
@keyframes pulse {
  from { color: var(--glow); }
  to { color: blue; }
}`)

	// When
	ResolveStyles(tree, ss, 40, 10)

	// Then
	spec := tree.Children[0].AnimationSpec
	if spec == nil {
		t.Fatal("no AnimationSpec stamped")
	}
	if len(spec.Stops) != 2 {
		t.Fatalf("stops = %d, want 2: %+v", len(spec.Stops), spec.Stops)
	}
	if got := spec.Stops[0].Style.FG.Name; got != "red" {
		t.Errorf("from color = %q, want red (var resolved from node scope)", got)
	}
	if got := spec.Stops[1].Style.FG.Name; got != "blue" {
		t.Errorf("to color = %q, want blue", got)
	}
}

func TestKeyframeStopsKeepLightDarkPairs(t *testing.T) {
	// Given
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "div", Kind: KindBox, Classes: []string{"fade"}},
	}}
	ss := sheet(t, `
.fade { animation: fade 1s; }
@keyframes fade {
  from { color: light-dark(white, black); }
  to { color: red; }
}`)

	// When
	ResolveStyles(tree, ss, 40, 10)

	// Then
	spec := tree.Children[0].AnimationSpec
	if spec == nil || len(spec.Stops) != 2 {
		t.Fatalf("spec = %+v", spec)
	}
	pair := spec.Stops[0].Style.FG.Pair
	if pair == nil {
		t.Fatalf("from color should carry a light-dark pair: %+v", spec.Stops[0].Style.FG)
	}
	if pair.Light.Name != "white" || pair.Dark.Name != "black" {
		t.Errorf("pair = %+v, want white/black", pair)
	}
}

func TestNoAnimationLeavesStopsEmpty(t *testing.T) {
	// Given
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "div", Kind: KindBox},
	}}
	ss := sheet(t, `div { color: red; }`)

	// When
	ResolveStyles(tree, ss, 40, 10)

	// Then
	if spec := tree.Children[0].AnimationSpec; spec != nil {
		t.Errorf("spec = %+v, want nil", spec)
	}
}
