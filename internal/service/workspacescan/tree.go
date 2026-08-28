package workspacescan

import (
	"os"
	"path/filepath"
	"sort"
)

// WalkTree traverses a project tree and returns relative path entries for files,
// truncated directories, collapsed output/deps dirs, and flagged project roots.
func WalkTree(startDir string, opts Options) ([]string, error) {
	opts = opts.normalized()
	if opts.Depth < 1 {
		return nil, nil
	}

	entries := make([]string, 0)
	if err := walkTree(startDir, startDir, 0, opts, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func walkTree(root, dir string, depth int, opts Options, entries *[]string) error {
	if depth >= opts.Depth {
		return nil
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	for _, entry := range dirEntries {
		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(relPath)

		if entry.IsDir() && opts.shouldCollapseDir(entry.Name()) {
			*entries = append(*entries, relSlash+opts.truncationSuffix())
			continue
		}

		if opts.shouldIgnoreEntry(entry.Name(), entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			subfiles, subdirs, listErr := ReadListing(fullPath)
			if listErr == nil {
				flags, flagErr := CollectFlags(opts, fullPath, subfiles, subdirs)
				if flagErr != nil {
					return flagErr
				}
				if opts.StopOnFlags && len(flags) > 0 {
					*entries = append(*entries, FormatPathWithFlags(relSlash, flags))
					continue
				}
			}

			if depth+1 >= opts.Depth {
				hasMore, err := hasVisibleDescendants(fullPath, opts)
				if err != nil {
					return err
				}
				if hasMore {
					*entries = append(*entries, relSlash+opts.truncationSuffix())
				}
				continue
			}
			if err := walkTree(root, fullPath, depth+1, opts, entries); err != nil {
				return err
			}
			continue
		}

		if !opts.IncludeFiles {
			continue
		}
		*entries = append(*entries, relSlash)
	}

	return nil
}

func hasVisibleDescendants(dir string, opts Options) (bool, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range dirEntries {
		if entry.IsDir() && opts.shouldCollapseDir(entry.Name()) {
			return true, nil
		}

		if opts.shouldIgnoreEntry(entry.Name(), entry.IsDir()) {
			continue
		}

		if !entry.IsDir() {
			return true, nil
		}

		childPath := filepath.Join(dir, entry.Name())
		hasMore, err := hasVisibleDescendants(childPath, opts)
		if err != nil {
			return false, err
		}
		if hasMore {
			return true, nil
		}
	}

	return false, nil
}
