package workspacescan

import (
	"path/filepath"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

// SiblingWalkOptions returns Options for workspace sibling directory scans.
func SiblingWalkOptions(root string, depth int, scanDotDirs bool, skipDir string) Options {
	opts := Options{
		Root:               root,
		Depth:              depth,
		CheckFlags:         true,
		StopOnFlags:        true,
		ShowNonFlag:        true,
		IgnoreHiddenDirs:   !scanDotDirs,
		IgnoreFiles:        globals.ScanIgnoreFiles,
		ApplyGlobalIgnores: true,
		IncludeFiles:       false,
	}
	if skipDir != "" {
		opts.IgnorePaths = []string{filepath.Clean(skipDir)}
	}
	return opts
}

// StructureWalkOptions returns Options for in-project structure scans.
func StructureWalkOptions(root string, depth int, expandOutDirs, expandDepsDirs bool) Options {
	return Options{
		Root:               root,
		Depth:              depth,
		CheckFlags:         true,
		StopOnFlags:        true,
		ShowNonFlag:        true,
		IgnoreHiddenDirs:   true,
		IgnoreHiddenFiles:  false,
		ApplyGlobalIgnores: true,
		ExpandOutDirs:      expandOutDirs,
		ExpandDepsDirs:     expandDepsDirs,
		TruncateMarker:     "/**",
		IncludeFiles:       true,
	}
}
