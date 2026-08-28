package workspacescan

import (
	"os"
	"sort"
)

// Listing is a single directory's immediate children and optional project flags.
type Listing struct {
	AbsPath  string
	RelPath  string
	Subfiles []string
	Subdirs  []string
	Flags    []string
}

// ReadListing reads one directory and returns sorted immediate child names.
func ReadListing(absPath string) (subfiles, subdirs []string, err error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, nil, err
	}
	subfiles, subdirs = SplitEntries(entries)
	return subfiles, subdirs, nil
}

// SplitEntries splits directory entries into sorted file and directory names.
func SplitEntries(entries []os.DirEntry) (subfiles, subdirs []string) {
	subfiles = make([]string, 0)
	subdirs = make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
			continue
		}
		subfiles = append(subfiles, entry.Name())
	}
	sort.Strings(subfiles)
	sort.Strings(subdirs)
	return subfiles, subdirs
}
