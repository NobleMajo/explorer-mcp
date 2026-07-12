package service

import (
	"os"
	"path/filepath"
	"sort"
)

type workspaceContextResponse struct {
	responseMeta
	CurrentWorkingDirectoryPath string           `json:"currentWorkingDirectoryPath"`
	ParentDirectoryPath         string           `json:"parentDirectoryPath"`
	SiblingProjectCount         int              `json:"siblingProjectCount"`
	SiblingProjects             []siblingProject `json:"siblingProjects"`
}

type siblingProject struct {
	DirectoryName    string `json:"directoryName"`
	AbsolutePath     string `json:"absolutePath"`
	IsCurrentProject bool   `json:"isCurrentProject"`
	IsGitRepo        bool   `json:"isGitRepo"`
}

func WorkspaceContext() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	parent := filepath.Dir(cwd)
	siblings, err := listSiblingProjects(parent, cwd)
	if err != nil {
		return "", err
	}

	return marshalResponse(workspaceContextResponse{
		responseMeta: responseMeta{
			ToolName:      "workspace_context",
			SchemaVersion: schemaVersion,
		},
		CurrentWorkingDirectoryPath: cwd,
		ParentDirectoryPath:         parent,
		SiblingProjectCount:         len(siblings),
		SiblingProjects:             siblings,
	})
}

func listSiblingProjects(parent, cwd string) ([]siblingProject, error) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}

	siblings := make([]siblingProject, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		absPath := filepath.Join(parent, entry.Name())
		siblings = append(siblings, siblingProject{
			DirectoryName:    entry.Name(),
			AbsolutePath:     absPath,
			IsCurrentProject: absPath == cwd,
			IsGitRepo:        hasGitMetadata(absPath),
		})
	}

	sort.Slice(siblings, func(i, j int) bool {
		return siblings[i].DirectoryName < siblings[j].DirectoryName
	})

	return siblings, nil
}

func hasGitMetadata(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
