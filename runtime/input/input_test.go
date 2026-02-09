package input

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestReadKey_SingleASCII(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte("a"))

	// When
	got, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 'a' {
		t.Errorf("got %q, want 'a'", got)
	}
}

func TestReadKey_MultiByteUTF8(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte("é"))

	// When
	got, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 'é' {
		t.Errorf("got %q, want 'é'", got)
	}
}

func TestReadKey_ThreeByteUTF8(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte("日"))

	// When
	got, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != '日' {
		t.Errorf("got %q, want '日'", got)
	}
}

func TestReadKey_FourByteUTF8(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte("😀"))

	// When
	got, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != '😀' {
		t.Errorf("got %q, want '😀'", got)
	}
}

func TestReadKey_EmptyReader(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte{})

	// When
	_, err := ReadKey(r)

	// Then
	if err != io.EOF {
		t.Errorf("got error %v, want io.EOF", err)
	}
}

func TestReadKey_ReturnsFirstCharOnly(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte("abc"))

	// When
	got, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 'a' {
		t.Errorf("got %q, want 'a'", got)
	}
}

func TestReadKey_SequentialReads(t *testing.T) {
	// Given
	r := bytes.NewReader([]byte("ab"))

	// When
	first, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error on first read: %v", err)
	}
	if first != 'a' {
		t.Errorf("first: got %q, want 'a'", first)
	}

	// When
	second, err := ReadKey(r)

	// Then
	if err != nil {
		t.Fatalf("unexpected error on second read: %v", err)
	}
	if second != 'b' {
		t.Errorf("second: got %q, want 'b'", second)
	}
}

func TestEnableRawMode_NonTerminalFd(t *testing.T) {
	// Given
	// A regular file fd is not a terminal — EnableRawMode should return an error.
	f, err := os.CreateTemp("", "input-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// When
	restore, err := EnableRawMode(int(f.Fd()))

	// Then
	if err == nil {
		t.Error("expected error for non-terminal fd, got nil")
		if restore != nil {
			restore()
		}
	}
}

func TestEnableRawMode_RestoreCallable(t *testing.T) {
	// Given
	// Even when EnableRawMode fails, if a restore function is returned,
	// it should be safe to call (including multiple times).
	f, err := os.CreateTemp("", "input-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// When
	restore, _ := EnableRawMode(int(f.Fd()))

	// Then
	if restore != nil {
		restore() // should not panic
		restore() // should be safe to call multiple times
	}
}
