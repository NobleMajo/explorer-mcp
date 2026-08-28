package structure

import (
	"github.com/NobleMajo/explorer-mcp/internal/service/workspacescan"
)

type ScanSettings struct {
	Depth    int
	OutDirs  bool
	DepsDirs bool
}

type repoStructureResponse struct {
	ProjectScanDepthLimit *int     `json:"projectScanDepthLimit,omitempty"`
	EntryCount            *int     `json:"entryCount,omitempty"`
	Entries               []string `json:"entries,omitempty"`
}

func buildRepoStructure(projectRootPath string, verbose bool, settings ScanSettings) (repoStructureResponse, error) {
	_ = verbose
	if settings.Depth < 1 {
		zero := 0
		return repoStructureResponse{ProjectScanDepthLimit: &zero}, nil
	}

	entries, err := workspacescan.WalkTree(
		projectRootPath,
		workspacescan.StructureWalkOptions(projectRootPath, settings.Depth, settings.OutDirs, settings.DepsDirs),
	)
	if err != nil {
		return repoStructureResponse{}, err
	}

	count := len(entries)
	resp := repoStructureResponse{
		ProjectScanDepthLimit: &settings.Depth,
		EntryCount:            &count,
	}
	if len(entries) > 0 {
		resp.Entries = entries
	}

	return resp, nil
}
