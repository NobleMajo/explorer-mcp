package parent

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
)

type workspaceContextResponse struct {
	jsonresp.Meta
	CurrentWorkingDirectoryPath string           `json:"currentWorkingDirectoryPath"`
	ParentDirectoryPath         string           `json:"parentDirectoryPath"`
	SiblingProjectCount         int              `json:"siblingProjectCount"`
	SiblingProjects             []siblingProject `json:"siblingProjects"`
}

type siblingProject struct {
	RelativePath string `json:"relativePath"`
	IsGitRepo    bool   `json:"isGitRepo"`
}

func buildWorkspaceContext(verbose bool) (workspaceContextResponse, error) {
	_ = verbose
	cwd, err := os.Getwd()
	if err != nil {
		return workspaceContextResponse{}, err
	}

	parent := filepath.Dir(cwd)
	siblings, err := listSiblingProjects(parent, cwd)
	if err != nil {
		return workspaceContextResponse{}, err
	}

	return workspaceContextResponse{
		Meta: jsonresp.Meta{
			ToolName:      "workspace_context",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		CurrentWorkingDirectoryPath: cwd,
		ParentDirectoryPath:         parent,
		SiblingProjectCount:         len(siblings),
		SiblingProjects:             siblings,
	}, nil
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
		if absPath == cwd {
			continue
		}

		relPath, err := filepath.Rel(cwd, absPath)
		if err != nil {
			return nil, err
		}

		siblings = append(siblings, siblingProject{
			RelativePath: filepath.ToSlash(relPath),
			IsGitRepo:    hasGitMetadata(absPath),
		})
	}

	sort.Slice(siblings, func(i, j int) bool {
		return siblings[i].RelativePath < siblings[j].RelativePath
	})

	return siblings, nil
}

func hasGitMetadata(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
