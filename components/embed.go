// Package components provides embedded built-in .sumi component sources.
package components

import "embed"

// FS embeds all .sumi component files.
//
//go:embed *.sumi
var FS embed.FS
