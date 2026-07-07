//go:build js

package main

import "github.com/tomyan/sumi/runtime/webterm"

func main() {
	webterm.Run(NewTodo(TodoProps{}))
}
