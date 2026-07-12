package service

import (
	"context"

	"github.com/NobleMajo/explorer-mcp/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func InitMcpService(cfg *config.AppConfig, name, version string) error {
	server := mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, nil)
	registerRoutes(server, cfg)
	return server.Run(context.Background(), &mcp.StdioTransport{})
}
