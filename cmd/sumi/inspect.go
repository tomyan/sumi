package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/tomyan/sumi/cmd/sumi/dev"
	"github.com/tomyan/sumi/runtime/tui"
)

// sumi inspect — dump the element tree / geometry of a running app
// (dev-launched apps listen on <appdir>/.sumi-dev.sock).

func inspectCommand(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	socket := fs.String("socket", "", "inspect socket (default: the sumi dev socket for --dir)")
	dir := fs.String("dir", ".", "app directory whose dev session to inspect")
	jsonOut := fs.Bool("json", false, "emit raw JSON")
	// Accept the subcommand in any position: flag parsing stops at the
	// first positional, so lift non-flag words out first.
	cmd := "boxes"
	var flagArgs []string
	for _, a := range args {
		if !strings.HasPrefix(a, "-") && (a == "tree" || a == "boxes") {
			cmd = a
			continue
		}
		flagArgs = append(flagArgs, a)
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if *socket == "" {
		*socket = dev.DevSocketPath(*dir)
	}
	conn, err := net.Dial("unix", *socket)
	if err != nil {
		return fmt.Errorf("connect %s: %w (is `sumi dev` running here?)", *socket, err)
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(map[string]string{"cmd": cmd}); err != nil {
		return err
	}
	var resp struct {
		Tree  *tui.InspectNode `json:"tree"`
		Error string           `json:"error"`
	}
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	if *jsonOut {
		out, _ := json.MarshalIndent(resp.Tree, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	printInspectNode(os.Stdout, resp.Tree, 0)
	return nil
}

// printInspectNode renders the dump as an indented tree.
func printInspectNode(w *os.File, n *tui.InspectNode, depth int) {
	if n == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	label := n.Tag
	if label == "" {
		label = "#" + n.Kind
	}
	if n.ID != "" {
		label += "#" + n.ID
	}
	for _, c := range n.Classes {
		label += "." + c
	}
	var extras []string
	if n.Content != "" {
		extras = append(extras, fmt.Sprintf("%q", n.Content))
	}
	if n.Box != nil {
		extras = append(extras, fmt.Sprintf("@%d,%d %dx%d", n.Box.X, n.Box.Y, n.Box.W, n.Box.H))
	}
	if n.Style != "" {
		extras = append(extras, "{"+n.Style+"}")
	}
	if n.Focused {
		extras = append(extras, ":focus")
	}
	if n.Hidden {
		extras = append(extras, "[hidden]")
	}
	fmt.Fprintf(w, "%s%s %s\n", indent, label, strings.Join(extras, " "))
	for _, c := range n.Children {
		printInspectNode(w, c, depth+1)
	}
}
