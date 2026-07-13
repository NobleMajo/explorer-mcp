# explorer-mcp
![CI/CD](https://github.com/noblemajo/explorer-mcp/actions/workflows/go-bin-release.yml/badge.svg)
![CI/CD](https://github.com/noblemajo/explorer-mcp/actions/workflows/go-test-build.yml/badge.svg)  
![MIT](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/dynamic/json?color=green&label=watchers&query=watchers&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2Fnoblemajo%2Fexplorer-mcp)
![](https://img.shields.io/badge/dynamic/json?color=yellow&label=stars&query=stargazers_count&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2Fnoblemajo%2Fexplorer-mcp)
![](https://img.shields.io/badge/dynamic/json?color=navy&label=forks&query=forks&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2Fnoblemajo%2Fexplorer-mcp)

explorer-mcp is a lightweight, read-only MCP server that gives AI quick access to Git repos, folder structures, and context. It cuts time and token usage by handling exploration internally and feeding results to AI agents.

Use `explorer-mcp print` to dump the same JSON the MCP `explore` tool returns.

# Table of Contents
- [Explore response design](#explore-response-design)
- [Requirements](#requirements)
- [Getting Started](#getting-started-1)
- [Install via go](#install-via-go)
- [Install via wget](#install-via-wget)
- [Build requirements](#build-requirements)
- [Build](#build-1)
- [Install go](#install-go)


## Explore response design

The `explore` JSON follows a few consistent rules:

- **Only show what is there** — lists and maps use `omitempty`; empty arrays are omitted when a scan ran but found nothing.
- **Do not show what is not found** — whole sections are omitted when disabled by flag or when prerequisites are missing (e.g. no `git` binary, not a git repo, no container CLI).
- **Combine details into string arrays** — dependencies, container rows, git status lines, and sibling paths are compact encoded strings instead of nested objects.
- **Use small flags for metadata** — booleans like `parentScanPerformed`, `recentCommitsListed`, and `repoScanDepthLimit` tell the agent whether a scan ran vs. what was found.
- **Behavior hints follow data** — `agentBehaviorInstructions` only includes domains whose section is present and non-empty; use `-B` / `--enable-behavior` to include behavior text.
- **At least one overview required** — if every overview is disabled, `print` and `explore` return an error.

Depth/count flags (`-c`, `-p`, `-d`) control how much is collected; disable flags (`-S`, `-G`, …) skip entire sections.

More flags via:
```sh
go run github.com/NobleMajo/explorer-mcp@latest -h
```

Example output for current working dir:
```sh
go run github.com/NobleMajo/explorer-mcp@latest print
```


# Getting Started

## Requirements
Linux- or macos-like systems with `go` or `wget & tar` installed.

## Getting Started

Start the latest repo version directly without leaving stuff in the current working dir: 
```sh
go run github.com/NobleMajo/explorer-mcp@latest
```

## Install via go
###### *For this section go is required, check out the [install go guide](#install-go).*

```sh
go install https://github.com/NobleMajo/explorer-mcp
```

## Install via wget
```sh
export CUSTOM_BIN_DIR="/usr/local/bin" # <- change if needed
export EXPLORER_MCP_VERSION="" # <- set latest version here

rm -rf $CUSTOM_BIN_DIR/explorer-mcp
wget https://github.com/NobleMajo/explorer-mcp/releases/download/v$EXPLORER_MCP_VERSION/explorer-mcp-v$EXPLORER_MCP_VERSION-linux-amd64.tar.gz -O /tmp/explorer-mcp.tar.gz
tar -xzvf /tmp/explorer-mcp.tar.gz -C $CUSTOM_BIN_DIR/ explorer-mcp
rm /tmp/explorer-mcp.tar.gz
```

# Build
## Build requirements
To build, you need to install go. 
The required go version is in the `go.mod` file.

## Build
###### *For this section go is required, check out the [install go guide](#install-go).*

Clone the repo:
```sh
git clone https://github.com/NobleMajo/explorer-mcp.git
cd explorer-mcp
```

Build the explorer-mcp binary from source code:
```sh
make build
./explorer-mcp
```

# Development
###### *For this section go is required, check out the [install go guide](#install-go).*

This part is work in process, i want use 'AIR' as autoreload tool:
```sh
make dev #WIP
```

## Install go
The required go version for this project is in the `go.mod` file.

To install and update go, I can recommend the following repo:
```sh
git clone git@github.com:udhos/update-golang.git golang-updater
cd golang-updater
sudo ./update-golang.sh
```

# Contributing
Contributions to this project are welcome!  
Interested users can refer to the guidelines provided in the [CONTRIBUTING.md](CONTRIBUTING.md) file to contribute to the project and help improve its functionality and features.

# License
This project is licensed under the [MIT license](LICENSE), providing users with flexibility and freedom to use and modify the software according to their needs.

# Disclaimer
This project is provided without warranties.  
Users are advised to review the accompanying license for more information on the terms of use and limitations of liability.
