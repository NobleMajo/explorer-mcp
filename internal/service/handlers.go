package service

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	schemaVersion    = 1
	maxToolJSONBytes = 32 * 1024
)

type responseMeta struct {
	ToolName      string `json:"toolName"`
	SchemaVersion int    `json:"schemaVersion"`
}

func marshalResponse(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var readOnlyToolAnnotations = &mcp.ToolAnnotations{
	ReadOnlyHint: true,
}

func registerExploreTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "explore",
		Description: "Default workspace overview as JSON — combines repoStructure, gitOverview, workspaceContext, and dependencies in one call",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return jsonToolResult(Explore)
	})
}

func registerRepoStructureTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "repo_structure",
		Description: "Folder tree as JSON (max depth 3, ignores .git node_modules vendor cache tmp dirs)",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return jsonToolResult(RepoStructure)
	})
}

func registerGitOverviewTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "git_overview",
		Description: "Git repo status as JSON (branch, isWorkingTreeClean, commits, diff stat)",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return jsonToolResult(GitOverview)
	})
}

func registerDependenciesTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dependencies",
		Description: "Package ecosystems and dependency names as JSON (go.mod, package.json)",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return jsonToolResult(Dependencies)
	})
}

func registerWorkspaceContextTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "workspace_context",
		Description: "Workspace layout as JSON (parent directory, sibling projects, isGitRepo per sibling)",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return jsonToolResult(WorkspaceContext)
	})
}

type truncatedToolResponse struct {
	responseMeta
	IsOutputTruncated bool   `json:"isOutputTruncated"`
	OutputByteCount   int    `json:"outputByteCount"`
	TruncatedJsonText string `json:"truncatedJsonText"`
}

func limitJSONOutput(jsonText string) (string, error) {
	if len(jsonText) <= maxToolJSONBytes {
		return jsonText, nil
	}

	var meta struct {
		ToolName string `json:"toolName"`
	}
	_ = json.Unmarshal([]byte(jsonText), &meta)

	return marshalResponse(truncatedToolResponse{
		responseMeta: responseMeta{
			ToolName:      meta.ToolName,
			SchemaVersion: schemaVersion,
		},
		IsOutputTruncated: true,
		OutputByteCount:   len(jsonText),
		TruncatedJsonText: jsonText[:maxToolJSONBytes],
	})
}

func jsonToolResult(fn func() (string, error)) (*mcp.CallToolResult, any, error) {
	jsonText, err := fn()
	if err != nil {
		return nil, nil, err
	}

	jsonText, err = limitJSONOutput(jsonText)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: jsonText},
		},
	}, nil, nil
}
