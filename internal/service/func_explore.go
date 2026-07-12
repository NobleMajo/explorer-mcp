package service

import "encoding/json"

type exploreResponse struct {
	responseMeta
	RepoStructure    json.RawMessage `json:"repoStructure"`
	GitOverview      json.RawMessage `json:"gitOverview"`
	WorkspaceContext json.RawMessage `json:"workspaceContext"`
	Dependencies     json.RawMessage `json:"dependencies"`
}

func Explore() (string, error) {
	repoStructure, err := toolSection(RepoStructure)
	if err != nil {
		return "", err
	}

	gitOverview, err := toolSection(GitOverview)
	if err != nil {
		return "", err
	}

	workspaceContext, err := toolSection(WorkspaceContext)
	if err != nil {
		return "", err
	}

	dependencies, err := toolSection(Dependencies)
	if err != nil {
		return "", err
	}

	return marshalResponse(exploreResponse{
		responseMeta: responseMeta{
			ToolName:      "explore",
			SchemaVersion: schemaVersion,
		},
		RepoStructure:    repoStructure,
		GitOverview:      gitOverview,
		WorkspaceContext: workspaceContext,
		Dependencies:     dependencies,
	})
}

func toolSection(fn func() (string, error)) (json.RawMessage, error) {
	text, err := fn()
	if err != nil {
		return nil, err
	}
	return json.RawMessage(text), nil
}
