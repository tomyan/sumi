package main

import "github.com/tomyan/sumi/runtime/tui"

//go:generate go run ../../cmd/sumi generate .

func main() {
	tui.Run(NewPanels(PanelsProps{}))
}
