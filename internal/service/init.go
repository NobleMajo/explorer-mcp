package service

import (
	"context"

	"github.com/NobleMajo/explorer-mcp/internal/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func InitMcpService(cfg *config.AppConfig, name, version string) error {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: name, Version: version}, nil)
	registerRoutes(server, cfg)
	return server.Run(context.Background(), &mcpsdk.StdioTransport{})
}

func DirectJsonResult(cfg *config.AppConfig) (string, error) {
	return buildExploreResponse(exploreSettingsFromConfig(cfg))
}
