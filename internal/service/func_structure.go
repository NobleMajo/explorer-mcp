package service

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const structureMaxDepth = 3

var structureIgnore = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".tmp":         true,
}

func RepoStructure() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	lines := []string{root}
	if err := appendStructureLines(root, 0, &lines); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

func appendStructureLines(dir string, depth int, lines *[]string) error {
	if depth >= structureMaxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if structureIgnore[entry.Name()] {
			continue
		}

		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}

		*lines = append(*lines, strings.Repeat("  ", depth+1)+name)

		if entry.IsDir() {
			if err := appendStructureLines(filepath.Join(dir, entry.Name()), depth+1, lines); err != nil {
				return err
			}
		}
	}

	return nil
}
