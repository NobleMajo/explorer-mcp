package workspacescan

import (
	"path/filepath"
	"slices"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

// Options configures directory traversal for workspace and structure scans.
type Options struct {
	// Root is the path relative paths are computed from.
	Root string
	// Depth is the maximum directory depth to traverse below each walk start.
	// Values below 1 disable traversal.
	Depth int

	// CheckFlags enables project flag detection via globals.CollectSiblingProjectFlags.
	CheckFlags bool
	// StopOnFlags prevents descending into directories that have project flags.
	StopOnFlags bool
	// ShowNonFlag when false emits only directories that have project flags.
	ShowNonFlag bool

	// IgnoreHiddenDirs skips directory names prefixed with "." when true.
	IgnoreHiddenDirs bool
	// IgnoreHiddenFiles skips file names prefixed with "." when true.
	IgnoreHiddenFiles bool
	// IgnoreFiles skips file and directory names in this list.
	IgnoreFiles []string
	// ApplyGlobalIgnores applies globals.IsScanIgnored when true.
	ApplyGlobalIgnores bool
	// IgnorePaths skips absolute paths during traversal (cleaned before compare).
	IgnorePaths []string
	// SkipDirs skips directory basenames during traversal.
	SkipDirs []string

	// ExpandOutDirs lists dist/out/output contents instead of collapsing them.
	ExpandOutDirs bool
	// ExpandDepsDirs lists node_modules/vendor contents instead of collapsing them.
	ExpandDepsDirs bool
	// TruncateMarker is appended to directory paths truncated at Depth (default /**).
	TruncateMarker string

	// IncludeFiles enables file entries during tree walks (structure scan).
	IncludeFiles bool
}

func (o Options) normalized() Options {
	out := o
	out.Root = filepath.Clean(out.Root)
	out.IgnorePaths = cleanPaths(out.IgnorePaths)
	return out
}

func cleanPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.Clean(path)
		if path == "" {
			continue
		}
		out = append(out, path)
	}
	return out
}

func (o Options) shouldIgnoreEntry(name string, isDir bool) bool {
	if isDir {
		if o.IgnoreHiddenDirs && isHiddenName(name) {
			return true
		}
	} else if o.IgnoreHiddenFiles && isHiddenName(name) {
		return true
	}
	if len(o.IgnoreFiles) > 0 && slices.Contains(o.IgnoreFiles, name) {
		return true
	}
	if o.ApplyGlobalIgnores && globals.IsScanIgnored(name) {
		if isDir && o.ExpandDepsDirs && IsDepsDir(name) {
			return false
		}
		return true
	}
	return false
}

func (o Options) shouldSkipPath(absPath string) bool {
	absPath = filepath.Clean(absPath)
	for _, ignored := range o.IgnorePaths {
		if absPath == ignored {
			return true
		}
	}
	return false
}

func isHiddenName(name string) bool {
	return len(name) > 0 && name[0] == '.'
}
