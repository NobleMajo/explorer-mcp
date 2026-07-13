package parent

import (
	"os"
	"path/filepath"
	"sort"
)

type workspaceContextResponse struct {
	CurrentWorkingDirectoryPath string   `json:"currentWorkingDirectoryPath"`
	ParentDirectoryPath         string   `json:"parentDirectoryPath"`
	SiblingProjectCount         int      `json:"siblingProjectCount"`
	GitSiblingProjects          []string `json:"gitSiblingProjects"`
	SiblingProjects             []string `json:"siblingProjects"`
}

func buildWorkspaceContext(verbose bool) (workspaceContextResponse, error) {
	_ = verbose
	cwd, err := os.Getwd()
	if err != nil {
		return workspaceContextResponse{}, err
	}

	parent := filepath.Dir(cwd)
	gitSiblings, siblings, err := listSiblingProjects(parent, cwd)
	if err != nil {
		return workspaceContextResponse{}, err
	}

	return workspaceContextResponse{
		CurrentWorkingDirectoryPath: cwd,
		ParentDirectoryPath:         parent,
		SiblingProjectCount:         len(gitSiblings) + len(siblings),
		GitSiblingProjects:          gitSiblings,
		SiblingProjects:             siblings,
	}, nil
}

func listSiblingProjects(parent, cwd string) (gitSiblings, siblings []string, err error) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, nil, err
	}

	gitSiblings = make([]string, 0)
	siblings = make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		absPath := filepath.Join(parent, entry.Name())
		if absPath == cwd {
			continue
		}

		relPath, err := filepath.Rel(cwd, absPath)
		if err != nil {
			return nil, nil, err
		}

		path := filepath.ToSlash(relPath)
		if hasGitMetadata(absPath) {
			gitSiblings = append(gitSiblings, path)
			continue
		}
		siblings = append(siblings, path)
	}

	sort.Strings(gitSiblings)
	sort.Strings(siblings)
	return gitSiblings, siblings, nil
}

func hasGitMetadata(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
