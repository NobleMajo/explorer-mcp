package service

import (
	"github.com/NobleMajo/explorer-mcp/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerRoutes(server *mcp.Server, cfg *config.AppConfig) {
	_ = cfg
	registerExploreTool(server)
	registerRepoStructureTool(server)
	registerGitOverviewTool(server)
	registerDependenciesTool(server)
	registerWorkspaceContextTool(server)
}
