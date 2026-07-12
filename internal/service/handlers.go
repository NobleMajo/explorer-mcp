package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerRepoStructureTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "repo_structure",
		Description: "Repo file tree overview (max depth 3, ignores .git node_modules vendor .tmp)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return textToolResult(RepoStructure)
	})
}

func registerGitOverviewTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "git_overview",
		Description: "Git status, branch, commits and diffs (not implemented)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return textToolResult(GitOverview)
	})
}

func registerDependenciesTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dependencies",
		Description: "Package manager dependencies (not implemented)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return textToolResult(Dependencies)
	})
}

func registerWorkspaceContextTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "workspace_context",
		Description: "Parent directory and sibling projects (not implemented)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return textToolResult(WorkspaceContext)
	})
}

func textToolResult(fn func() (string, error)) (*mcp.CallToolResult, any, error) {
	text, err := fn()
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, nil, nil
}
