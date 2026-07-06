package main

import (
	"fmt"
	"os"

	"github.com/tomyan/sumi/cmd/sumi/dev"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sumi <command> [args]")
		fmt.Fprintln(os.Stderr, "commands: init, dev, generate, test-preview")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := initCommand(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "dev":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if err := dev.RunDev(dir, generateDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "generate":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if err := generateDir(dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "test-preview":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: sumi test-preview <component-dir>")
			os.Exit(1)
		}
		if err := testPreview(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
