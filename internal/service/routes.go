package service

import (
	"github.com/NobleMajo/explorer-mcp/internal/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerRoutes(server *mcpsdk.Server, cfg *config.AppConfig) {
	registerExploreTool(server, cfg.Verbose)
}
