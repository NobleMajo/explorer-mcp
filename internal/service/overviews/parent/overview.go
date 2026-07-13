package parent

import (
	"os"
	"path/filepath"
	"sort"
)

type workspaceContextResponse struct {
	ParentScanPerformed bool     `json:"parentScanPerformed"`
	SiblingProjectCount *int     `json:"siblingProjectCount,omitempty"`
	SiblingProjects     []string `json:"siblingProjects,omitempty"`
}

func buildWorkspaceContext(projectRootPath string, verbose bool, settings ScanSettings) (workspaceContextResponse, error) {
	_ = verbose
	cwd := projectRootPath

	resp := workspaceContextResponse{
		ParentScanPerformed: settings.Depth > 0,
	}
	if settings.Depth < 1 {
		return resp, nil
	}

	siblings, err := listParentProjects(cwd, settings)
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
