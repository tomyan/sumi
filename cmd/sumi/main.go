package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sumi generate [path]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if err := generateDir(dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
