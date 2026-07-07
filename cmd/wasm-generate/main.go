// Command wasm-generate is the js/wasm build of `sumi generate` for a single
// file: it reads a .sumi source path from argv, runs the parse+codegen
// pipeline, and writes the generated Go beside it. It is invoked inside the
// browser toolchain host (see sumi-site's engine) as codegen.wasm.
package main

import (
	"fmt"
	"os"

	"github.com/tomyan/sumi/gen"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: wasm-generate <file.sumi>")
		os.Exit(2)
	}
	if err := generate(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generate reads the .sumi file, compiles it, and writes the _sumi.go output.
func generate(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	out, err := gen.Generate(path, string(src))
	if err != nil {
		return err
	}
	outPath := gen.OutputPath(path)
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return fmt.Errorf("%s: %w", outPath, err)
	}
	return nil
}
