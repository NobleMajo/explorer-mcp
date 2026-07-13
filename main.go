package main

import (
	"errors"
	"fmt"
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

	if appConfig.DirectOut {
		out, err := service.DirectJsonResult(appConfig.Verbose)
		if err != nil {
			log.Printf("explore output failed: %v\n", err)
			os.Exit(1)
		}

		if appConfig.Verbose {
			log.Printf(
				"Run explore output of %s (%s) v%s build %s\n",
				DisplayName,
				ShortName,
				Version,
				Commit,
			)
			log.Println(out)
		} else {
			fmt.Println(out)
		}

		return
	}

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
