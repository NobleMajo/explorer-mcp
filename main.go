package main

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/config"
	"github.com/NobleMajo/explorer-mcp/internal/service"
	"github.com/joho/godotenv"
)

var DisplayName string = "Unset"
var ShortName string = "unset"
var Version string = "?.?.?"
var Commit string = "???????"

func main() {
	_ = godotenv.Load()

	appConfig := config.ParseConfig(DisplayName, ShortName, Version, Commit)

	log.Printf(
		"starting %s (%s) v%s build %s on MCP stdio\n",
		DisplayName,
		ShortName,
		Version,
		Commit,
	)

	if err := service.InitMcpService(appConfig, DisplayName, Version); err != nil {
		if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "EOF") {
			return
		}
		log.Printf("MCP server failed: %v\n", err)
		os.Exit(1)
	}
}
