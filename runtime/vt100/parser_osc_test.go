package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestOSCSentinelWithBEL(t *testing.T) {
	// Given — sentinel terminated with BEL (0x07)
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b]999;done\x07"))

	// Then
	if !screen.SentinelSeen() {
		t.Error("SentinelSeen() = false, want true")
	}
}

func TestOSCSentinelWithST(t *testing.T) {
	// Given — sentinel terminated with ST (ESC\)
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b]999;done\x1b\\"))

	// Then
	if !screen.SentinelSeen() {
		t.Error("SentinelSeen() = false, want true after ST terminator")
	}
}

func TestOSCSTDoesNotCorruptFollowing(t *testing.T) {
	// Given — OSC with ST followed by text
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b]2;title\x1b\\X"))

	// Then — X at (0,0)
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestOSCSTSplitAcrossCalls(t *testing.T) {
	// Given — OSC payload in first call, ESC\ in second
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b]999;done"))
	screen.Write([]byte("\x1b\\"))

	// Then
	if !screen.SentinelSeen() {
		t.Error("SentinelSeen() = false, want true after split ST")
	}
}
