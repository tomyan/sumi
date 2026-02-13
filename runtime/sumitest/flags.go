package sumitest

import "flag"

var updateSnapshots bool
var previewMode bool

func init() {
	flag.BoolVar(&updateSnapshots, "update", false, "update snapshot files")
	flag.BoolVar(&previewMode, "preview", false, "preview scenario frames interactively")
}

// UpdateMode returns true when the -update flag is set.
func UpdateMode() bool {
	return updateSnapshots
}

// PreviewMode returns true when the -preview flag is set.
func PreviewMode() bool {
	return previewMode
}
