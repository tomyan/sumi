// Package components provides embedded built-in .sumi component sources.
package components

import "embed"

// FS embeds all .sumi component files including subdirectories.
//
//go:embed *.sumi sumi/*.sumi
var FS embed.FS
