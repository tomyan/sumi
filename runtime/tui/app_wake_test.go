package tui

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
)

func TestWakeTriggersEventFrame(t *testing.T) {
	// Given
	eventCh := make(chan input.Event, 1)

	var gotFrame bool
	app := &App{
		OnRender: func() {},
	}
	app.initQuit()
	app.OnEvent = func(evt input.Event) {
		if evt.Kind == input.EventFrame {
			gotFrame = true
			app.Quit()
		}
	}

	// When — Wake sends immediately (no delay like RequestFrame)
	app.Wake()
	app.runLoop(eventCh, nil, nil)

	// Then
	if !gotFrame {
		t.Fatal("Wake did not trigger EventFrame dispatch")
	}
}

func TestWakeIsNonBlocking(t *testing.T) {
	// Given — wakeCh has capacity 1
	app := &App{}
	app.initQuit()

	// When — calling Wake twice should not block
	app.Wake()
	app.Wake()

	// Then — no deadlock, test completes
}
