package service

import (
	"os"
	"path/filepath"
	"sort"
)

type repoStructureResponse struct {
	responseMeta
	RootPath              string           `json:"rootPath"`
	MaxDepth              int              `json:"maxDepth"`
	IgnoredDirectoryNames []string         `json:"ignoredDirectoryNames"`
	IgnoredFileNames      []string         `json:"ignoredFileNames"`
	EntryCount            int              `json:"entryCount"`
	Entries               []structureEntry `json:"entries"`
}

type structureEntry struct {
	RelativePath string `json:"relativePath"`
	EntryName    string `json:"entryName"`
	IsDirectory  bool   `json:"isDirectory"`
	Depth        int    `json:"depth"`
}

func RepoStructure() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	entries := make([]structureEntry, 0)
	if err := appendStructureEntries(root, root, 0, &entries); err != nil {
		return "", err
	}

	ignoredDirNames := append([]string(nil), ScanIgnoreFiles...)
	sort.Strings(ignoredDirNames)
	ignoredFileNames := append([]string(nil), IgnoreFiles...)
	sort.Strings(ignoredFileNames)

	return marshalResponse(repoStructureResponse{
		responseMeta: responseMeta{
			ToolName:      "repo_structure",
			SchemaVersion: schemaVersion,
		},
		RootPath:              root,
		MaxDepth:              structureScanMaxDepth,
		IgnoredDirectoryNames: ignoredDirNames,
		IgnoredFileNames:      ignoredFileNames,
		EntryCount:            len(entries),
		Entries:               entries,
	})
}

func appendStructureEntries(root, dir string, depth int, entries *[]structureEntry) error {
	if depth >= structureScanMaxDepth {
		return nil
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		if dirEntries[i].IsDir() != dirEntries[j].IsDir() {
			return dirEntries[i].IsDir()
		}
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	for _, entry := range dirEntries {
		if isScanIgnored(entry.Name()) {
			continue
		}
		if !entry.IsDir() && isIgnoredFile(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		if entry.IsDir() {
			relPath += "/"
		}

		*entries = append(*entries, structureEntry{
			RelativePath: relPath,
			EntryName:    entry.Name(),
			IsDirectory:  entry.IsDir(),
			Depth:        depth + 1,
		})

		if entry.IsDir() {
			if err := appendStructureEntries(root, fullPath, depth+1, entries); err != nil {
				return err
			}
		}
	}

	return nil
}
