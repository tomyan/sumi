package layout

import "testing"

func TestLayoutPropagatesKey(t *testing.T) {
	// Given
	input := &Input{
		Kind:    KindText,
		Content: "hello",
		Key:     "foo",
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Key != "foo" {
		t.Errorf("box.Key = %q, want %q", box.Key, "foo")
	}
}

func TestLayoutNoKey(t *testing.T) {
	// Given
	input := &Input{
		Kind:    KindText,
		Content: "hello",
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Key != "" {
		t.Errorf("box.Key = %q, want empty string", box.Key)
	}
}

func TestLayoutPropagatesKeyOnChildren(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "a", Key: "k1"},
			{Kind: KindText, Content: "b", Key: "k2"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if len(box.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(box.Children))
	}
	if box.Children[0].Key != "k1" {
		t.Errorf("child[0].Key = %q, want %q", box.Children[0].Key, "k1")
	}
	if box.Children[1].Key != "k2" {
		t.Errorf("child[1].Key = %q, want %q", box.Children[1].Key, "k2")
	}
}
