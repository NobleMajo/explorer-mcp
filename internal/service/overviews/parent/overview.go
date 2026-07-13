package parent

import (
	"os"
	"path/filepath"
	"sort"
)

type workspaceContextResponse struct {
	CurrentWorkingDirectoryPath string   `json:"currentWorkingDirectoryPath"`
	ParentDirectoryPath         string   `json:"parentDirectoryPath"`
	ParentScanPerformed         bool     `json:"parentScanPerformed"`
	SiblingProjectCount         *int     `json:"siblingProjectCount,omitempty"`
	SiblingProjects             []string `json:"siblingProjects,omitempty"`
}

func buildWorkspaceContext(verbose bool, parentScanDepth int) (workspaceContextResponse, error) {
	_ = verbose
	cwd, err := os.Getwd()
	if err != nil {
		return workspaceContextResponse{}, err
	}

	resp := workspaceContextResponse{
		CurrentWorkingDirectoryPath: cwd,
		ParentDirectoryPath:         filepath.Dir(cwd),
		ParentScanPerformed:         parentScanDepth > 0,
	}
	if parentScanDepth < 1 {
		return resp, nil
	}

	siblings, err := listParentProjects(cwd, parentScanDepth)
	if err != nil {
		return workspaceContextResponse{}, err
	}

	sort.Strings(siblings)

	count := len(siblings)
	resp.SiblingProjectCount = &count
	if len(siblings) > 0 {
		resp.SiblingProjects = siblings
	}

	return resp, nil
}

func hasGitMetadata(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
