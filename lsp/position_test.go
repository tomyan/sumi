package lsp

import "testing"

func TestOffsetToPositionBasic(t *testing.T) {
	// Given
	text := "abc\ndefg\nhi"

	// When: offset at 'f' (index 6)
	pos := offsetToPosition(text, 6)

	// Then
	if pos.Line != 1 || pos.Character != 2 {
		t.Errorf("got %+v, want {Line:1 Character:2}", pos)
	}
}

func TestOffsetToPositionMultiByteBMP(t *testing.T) {
	// Given: 'é' is two UTF-8 bytes but one UTF-16 code unit
	text := "é<div>"

	// When: offset at '<' (byte index 2)
	pos := offsetToPosition(text, 2)

	// Then
	if pos.Line != 0 || pos.Character != 1 {
		t.Errorf("got %+v, want {Line:0 Character:1}", pos)
	}
}

func TestOffsetToPositionSupplementaryPlane(t *testing.T) {
	// Given: '😀' is four UTF-8 bytes and two UTF-16 code units
	text := "a😀b\nc"

	// When: offset at 'b' (byte index 5)
	pos := offsetToPosition(text, 5)

	// Then
	if pos.Line != 0 || pos.Character != 3 {
		t.Errorf("got %+v, want {Line:0 Character:3}", pos)
	}
}

func TestOffsetToPositionClampsOutOfRange(t *testing.T) {
	// Given
	text := "hi"

	// When
	pos := offsetToPosition(text, 999)

	// Then
	if pos.Line != 0 || pos.Character != 2 {
		t.Errorf("got %+v, want {Line:0 Character:2}", pos)
	}
}
