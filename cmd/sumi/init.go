package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// sumi init — scaffold a runnable app in a new or empty directory.

const initAppSumi = `<script>
count := sumi.New(0)

func increment(evt *sumi.DOMEvent) {
	count.Update(func(n int) int { return n + 1 })
}

func handleKey(evt sumi.Event) {
	if evt.Kind == sumi.EventSignal { sumi.Quit(); return }
	if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') { sumi.Quit(); return }
}
</script>

<style>
h1 {
	color: cyan;
}
button:focus {
	color: yellow;
}
.hint {
	opacity: dim;
}
</style>

<div onkey="handleKey">
	<h1>Hello, sumi</h1>
	<p>You have pressed the button <strong>{count}</strong> times.</p>
	<button onclick={increment}>Press me</button>
	<div class="hint">Tab to focus, Enter to press; q quits</div>
</div>
`

const initMainGo = `package main

import "github.com/tomyan/sumi/runtime/tui"

//go:generate sumi generate .

func main() {
	tui.Run(NewApp(AppProps{}))
}
`

// initCommand parses `sumi init [dir]` flags and runs the scaffold.
func initCommand(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	module := fs.String("module", "", "module path for the new app (default: example.com/<dir>)")
	sumiPath := fs.String("sumi-path", "", "local sumi checkout for the replace directive (default: $SUMI_PATH or upward search)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	if *module == "" {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		*module = "example.com/" + filepath.Base(abs)
	}
	if err := runInit(dir, *module, *sumiPath); err != nil {
		return err
	}
	fmt.Printf("scaffolded %s\n\nnext steps:\n  cd %s\n  go run .\n", *module, dir)
	return nil
}

// runInit scaffolds an app in dir with the given module path. sumiPath
// is the local sumi checkout for the replace directive ("" = locate it
// automatically; required until the module is published).
func runInit(dir, modulePath, sumiPath string) error {
	if err := ensureEmptyDir(dir); err != nil {
		return err
	}
	if sumiPath == "" {
		found, err := findSumiCheckout(".")
		if err != nil {
			return err
		}
		sumiPath = found
	}
	goMod := fmt.Sprintf("module %s\n\ngo 1.25\n\nrequire github.com/tomyan/sumi v0.0.0\n\nreplace github.com/tomyan/sumi => %s\n", modulePath, sumiPath)
	files := map[string]string{
		"app.sumi": initAppSumi,
		"main.go":  initMainGo,
		"go.mod":   goMod,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	if err := generateDir(dir); err != nil {
		return fmt.Errorf("generate scaffold: %w", err)
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if out, err := tidy.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, out)
	}
	return nil
}

// ensureEmptyDir creates dir if needed and refuses a non-empty one.
func ensureEmptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("directory %s is not empty", dir)
	}
	return nil
}

// findSumiCheckout locates the sumi source checkout for the go.mod
// replace directive: the SUMI_PATH env var, or an upward walk from
// start looking for sumi's own go.mod.
func findSumiCheckout(start string) (string, error) {
	if p := os.Getenv("SUMI_PATH"); p != "" {
		return p, nil
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		mod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(mod), "module github.com/tomyan/sumi\n") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot locate the sumi checkout: set SUMI_PATH or pass --sumi-path")
		}
		dir = parent
	}
}
