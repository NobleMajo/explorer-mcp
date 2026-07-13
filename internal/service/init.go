package service

import (
	"context"
	"os"

	"github.com/NobleMajo/explorer-mcp/internal/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func InitMcpService(cfg *config.AppConfig, name, version string) error {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: name, Version: version}, nil)
	registerRoutes(server, cfg)
	return server.Run(context.Background(), &mcpsdk.StdioTransport{})
}

func DirectJsonResult(cfg *config.AppConfig) (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return DirectJsonResultAt(cfg, root)
}

func DirectJsonResultAt(cfg *config.AppConfig, projectRootPath string) (string, error) {
	return buildExploreResponse(projectRootPath, exploreSettingsFromConfig(cfg))
}
