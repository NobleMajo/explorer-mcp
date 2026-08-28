package workspacescan

import (
	"os"
	"path/filepath"
	"slices"
)

// WalkFunc is called for each visited child directory under startDir.
type WalkFunc func(listing Listing) error

// WalkDown traverses directories under startDir up to opts.Depth levels.
// startDir itself is not passed to fn; only child directories are visited.
func WalkDown(startDir string, opts Options, fn WalkFunc) error {
	opts = opts.normalized()
	if opts.Depth < 1 {
		return nil
	}
	return walkDown(startDir, opts, 1, fn)
}

func walkDown(currentDir string, opts Options, currentDepth int, fn WalkFunc) error {
	if currentDepth > opts.Depth {
		return nil
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if opts.shouldIgnoreEntry(entry.Name(), true) {
			continue
		}
		if len(opts.SkipDirs) > 0 && slices.Contains(opts.SkipDirs, entry.Name()) {
			continue
		}

		absPath := filepath.Clean(filepath.Join(currentDir, entry.Name()))
		if opts.shouldSkipPath(absPath) {
			continue
		}

		subfiles, subdirs, err := ReadListing(absPath)
		if err != nil {
			continue
		}

		relPath, err := filepath.Rel(opts.Root, absPath)
		if err != nil {
			continue
		}
		relPath = filepath.ToSlash(relPath)

		flags, err := CollectFlags(opts, absPath, subfiles, subdirs)
		if err != nil {
			return err
		}

		if !opts.ShowNonFlag && len(flags) == 0 {
			continue
		}

		listing := Listing{
			AbsPath:  absPath,
			RelPath:  relPath,
			Subfiles: subfiles,
			Subdirs:  subdirs,
			Flags:    flags,
		}

		if fn != nil {
			if err := fn(listing); err != nil {
				return err
			}
		}

		if opts.StopOnFlags && len(flags) > 0 {
			continue
		}

		if err := walkDown(absPath, opts, currentDepth+1, fn); err != nil {
			return err
		}
	}

	return nil
}
