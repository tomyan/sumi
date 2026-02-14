package pty_test

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/tomyan/sumi/runtime/pty"
)

func TestOpenAndReadWrite(t *testing.T) {
	// Given — a PTY pair
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer master.Close()
	defer slave.Close()

	// When — write to slave, read from master
	msg := "hello pty\n"
	go func() {
		slave.WriteString(msg)
	}()

	buf := make([]byte, 64)
	master.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := master.Read(buf)
	if err != nil {
		t.Fatalf("read from master: %v", err)
	}

	// Then — master receives what slave wrote (PTY may echo, so check contains)
	got := string(buf[:n])
	if !strings.Contains(got, "hello pty") {
		t.Errorf("master read: got %q, want substring %q", got, "hello pty")
	}
}

func TestSetSize(t *testing.T) {
	// Given — a PTY pair
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer master.Close()
	defer slave.Close()

	// When — set size
	err = pty.SetSize(master, 24, 80)

	// Then — no error
	if err != nil {
		t.Errorf("SetSize: %v", err)
	}
}

func TestStartEchoCommand(t *testing.T) {
	// Given — a command that prints to stdout
	cmd := exec.Command("echo", "pty-test-output")

	// When — start on PTY
	master, err := pty.Start(cmd, 24, 80)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer master.Close()

	// Then — read output from master
	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(master)
		done <- string(data)
	}()

	cmd.Wait()

	select {
	case got := <-done:
		if !strings.Contains(got, "pty-test-output") {
			t.Errorf("master output: got %q, want substring %q", got, "pty-test-output")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout reading from master")
	}
}
