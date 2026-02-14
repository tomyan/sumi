package sumitest

import "flag"

var updateSnapshots bool
var previewMode bool
var serveMode bool

func init() {
	flag.BoolVar(&updateSnapshots, "update", false, "update snapshot files")
	flag.BoolVar(&previewMode, "preview", false, "preview scenario frames interactively")
	flag.BoolVar(&serveMode, "serve", false, "enter serve mode for external preview control")
}

// UpdateMode returns true when the -update flag is set.
func UpdateMode() bool {
	return updateSnapshots
}

// PreviewMode returns true when the -preview flag is set.
func PreviewMode() bool {
	return previewMode
}

// ServeMode returns true when the -serve flag is set.
func ServeMode() bool {
	return serveMode
}
