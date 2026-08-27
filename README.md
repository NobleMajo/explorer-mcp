<div align="center">

# 📡 explorer-mcp

![MIT](https://img.shields.io/badge/license-MIT-blue.svg)
![CI/CD](https://github.com/NobleMajo/explorer-mcp/actions/workflows/go-bin-release.yml/badge.svg)
![CI/CD](https://github.com/NobleMajo/explorer-mcp/actions/workflows/go-test-build.yml/badge.svg)  
![](https://img.shields.io/badge/dynamic/json?color=green&label=watchers&query=watchers&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2FNobleMajo%2Fexplorer-mcp)
![](https://img.shields.io/badge/dynamic/json?color=yellow&label=stars&query=stargazers_count&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2FNobleMajo%2Fexplorer-mcp)
![](https://img.shields.io/badge/dynamic/json?color=navy&label=forks&query=forks&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2FNobleMajo%2Fexplorer-mcp)

</div>

_Why should an AI agent collect context piece by piece when an MCP can collect more information faster and present it in a simple, self-explanatory, compact format?_

## About

explorer-mcp is a lightweight, MCP server that gives AI quick read-only access to Git repos, folder structures, and context. It cuts time and token usage by handling exploration internally and feeding results to AI agents.

<details><summary><strong>Explore resources</strong></summary>

## Explore resources

`explorer-mcp` collects different resources and provides them to the requesting AI Agent.

Resources can be enabled or disabled via CLI flags and environment variables. See `explorer-mcp -h` or `explorer-mcp print -h`.
Some resources are omitted at runtime when prerequisites are missing (no git repo, no container CLI, no `opencode` binary, no `gh`, no agent context files, etc.).

### Enabled by default

- **structure**: File tree under `projectRootPath` up to scan depth; relative paths, `/**` for truncated dirs
- **git**: Branch or detached head, dirty files, recent commits, diffstat
- **workspace**: Parent and sibling projects around `projectRootPath`
- **dependencies**: Dependencies from manifests (`go.mod`, `package.json`, `requirements.txt`, …)
- **container**: Container CLIs in PATH and running containers linked to the project
- **tools**: Makefile targets, `package.json` scripts, root shell scripts<!--  -->

### Opt-in

- **cli**: Common CLI tools available in PATH. This could mislead an AI agent and cause complications.
- **opencode**: Effective OpenCode permission rules and MCP server names via `opencode debug agent build`
- **gh**: Matching GitHub repos in user and org namespaces via `gh`
- **agentc**: Existing agent/project instruction file paths (`--agentc` / `ENABLE_AGENTC`). Path-only string array for root files like `AGENTS.md` / `CLAUDE.md` / `CONTRIBUTING.md`, flat `.cursor/rules/*`, and top-level `docs/*.{md,mdx}`; never reads contents. Omitted when disabled or none found.
- **behavior**: `agentBehaviorMainInstruction` and per-domain `agentBehaviorInstructions` for present sections

Quick help:

```sh
go run github.com/NobleMajo/explorer-mcp@latest -h
```

Use `explorer-mcp print [projectRootPath]` to dump the same JSON the MCP `explore` tool returns:

```sh
go run github.com/NobleMajo/explorer-mcp@latest print
```

</details>

<details><summary><strong>Response design</strong></summary>

## Response design

The `explore` JSON follows a few consistent rules:

- **Only show what is there**: lists and maps use `omitempty`; empty arrays are omitted when a scan ran but found nothing.
- **Do not show what is not found**: whole sections are omitted when disabled by flag or when prerequisites are missing (e.g. no `git` binary, not a git repo, no container CLI).
- **Combine details into string arrays**: dependencies, container rows, git status lines, and sibling paths are compact encoded strings instead of nested objects.
- **Use small flags for metadata**: booleans like `parentScanPerformed`, `recentCommitsListed`, and `repoScanDepthLimit` tell the agent whether a scan ran vs. what was found.
- **Behavior hints follow data**: `agentBehaviorInstructions` only includes domains whose section is present and non-empty; use `-B` / `--enable-behavior` to include behavior text.
- **At least one explore resource required**: if every explore resource is disabled, `print` and `explore` return an error.

Depth/count flags (`-c`, `-p`, `-d`) control how much is collected; disable flags (`-S`, `-G`, …) skip entire resources.
MCP note: the `explore` tool requires a **mandatory** input parameter `projectRootPath` (absolute or relative path to project root directory). The path is validated and passed through to all explore resource collectors.

</details>

<details><summary><strong>User Guide</strong></summary>

# User Guide

## Requirements

Linux- or macos-like systems with `go` or `wget & tar` installed.

## Getting Started

Start the latest repo version directly without leaving stuff in the current working dir:

```sh
go run github.com/NobleMajo/explorer-mcp@latest
```

## Quick help

```sh
go run github.com/NobleMajo/explorer-mcp@latest -h
```

## Install via go

###### _For this section go is required, check out the [install go guide](#install-go)._

```sh
go install github.com/NobleMajo/explorer-mcp@latest
```

## Install via wget

```sh
export CUSTOM_BIN_DIR="/usr/local/bin" # <- change if needed
export CUSTOM_VERSION="" # <- set latest version here

rm -rf $CUSTOM_BIN_DIR/explorer-mcp
wget https://github.com/NobleMajo/explorer-mcp/releases/download/v$CUSTOM_VERSION/explorer-mcp-v$CUSTOM_VERSION-linux-amd64.tar.gz -O /tmp/explorer-mcp.tar.gz
tar -xzvf /tmp/explorer-mcp.tar.gz -C $CUSTOM_BIN_DIR/ explorer-mcp
rm /tmp/explorer-mcp.tar.gz
```

# Build

## Build requirements

To build, you need to install go.
The required go version is in the `go.mod` file.

## Build Instructions

###### _For this section go is required, check out the [install go guide](#install-go)._

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

</details>

<details><summary><strong>Development</strong></summary>

# Development

###### _For this section go is required, check out the [install go guide](#install-go)._

This part is work in progress, I want to use 'AIR' as auto-reload tool:

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

</details>

<div align="center">

# 🤝 Contributing

Contributions to this project are welcome!  
Follow the [CONTRIBUTING.md](CONTRIBUTING.md) for more infos.

# ⚠️ Disclaimer

This project is provided without warranties.

# 📜 License

Licensed under the [MIT license](LICENSE).

<a href="https://discord.coreunit.net">
    <img alt="CoreUnit.NET Discord Banner" src="https://discord.com/api/guilds/422136748294930443/widget.png?style=banner2">
</a>

</div>