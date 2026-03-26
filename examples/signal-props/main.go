package main

import (
	"github.com/tomyan/sumi/examples/signal-props/greeting"
	"github.com/tomyan/sumi/runtime/tui"
)

func main() {
	comp := greeting.NewGreeting(greeting.GreetingProps{
		Name: "Sumi",
	})
	tui.Run(comp)
}
