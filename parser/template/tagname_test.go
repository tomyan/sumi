package template

import "testing"

func TestClassifyPrimitiveTags(t *testing.T) {
	// Given / When / Then
	for _, tag := range []string{"box", "text", "title"} {
		tier := ClassifyTag(tag)
		if tier != TierPrimitive {
			t.Errorf("ClassifyTag(%q) = %v, want TierPrimitive", tag, tier)
		}
	}
}

func TestClassifyStdlibTags(t *testing.T) {
	// Given / When / Then
	tests := []string{"sumi:TextInput", "sumi:Select", "sumi:Whatever"}
	for _, tag := range tests {
		tier := ClassifyTag(tag)
		if tier != TierStdlib {
			t.Errorf("ClassifyTag(%q) = %v, want TierStdlib", tag, tier)
		}
	}
}

func TestClassifyUserTags(t *testing.T) {
	// Given / When / Then
	tests := []string{"myui:Button", "lib:Card"}
	for _, tag := range tests {
		tier := ClassifyTag(tag)
		if tier != TierUser {
			t.Errorf("ClassifyTag(%q) = %v, want TierUser", tag, tier)
		}
	}
}

func TestClassifyFundamentalTags(t *testing.T) {
	// Given / When / Then
	tests := []string{"textedit", "scrollbar", "counter", "mywidget"}
	for _, tag := range tests {
		tier := ClassifyTag(tag)
		if tier != TierFundamental {
			t.Errorf("ClassifyTag(%q) = %v, want TierFundamental", tag, tier)
		}
	}
}

func TestTagRegistryKey(t *testing.T) {
	// Given / When / Then
	tests := []struct {
		input string
		want  string
	}{
		{"counter", "counter"},
		{"textedit", "textedit"},
		{"sumi:TextInput", "sumi:textinput"},
		{"sumi:Select", "sumi:select"},
		{"myui:Button", "myui:button"},
	}
	for _, tt := range tests {
		got := TagRegistryKey(tt.input)
		if got != tt.want {
			t.Errorf("TagRegistryKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTagComponentFile(t *testing.T) {
	// Given / When / Then
	tests := []struct {
		input string
		want  string
	}{
		{"counter", "counter.sumi"},
		{"textedit", "textedit.sumi"},
		{"sumi:TextInput", "text-input.sumi"},
		{"sumi:Select", "select.sumi"},
		{"sumi:FancyButton", "fancy-button.sumi"},
		{"myui:Button", "button.sumi"},
	}
	for _, tt := range tests {
		got := TagComponentFile(tt.input)
		if got != tt.want {
			t.Errorf("TagComponentFile(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
