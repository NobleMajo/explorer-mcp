# explorer-mcp
![CI/CD](https://github.com/noblemajo/explorer-mcp/actions/workflows/go-bin-release.yml/badge.svg)
![CI/CD](https://github.com/noblemajo/explorer-mcp/actions/workflows/go-test-build.yml/badge.svg)  
![MIT](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/dynamic/json?color=green&label=watchers&query=watchers&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2Fnoblemajo%2Fexplorer-mcp)
![](https://img.shields.io/badge/dynamic/json?color=yellow&label=stars&query=stargazers_count&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2Fnoblemajo%2Fexplorer-mcp)
![](https://img.shields.io/badge/dynamic/json?color=navy&label=forks&query=forks&suffix=x&url=https%3A%2F%2Fapi.github.com%2Frepos%2Fnoblemajo%2Fexplorer-mcp)

`explorer-mcp` is a small, read-only MCP server that provides AI with a fast Git repository, folder structure and other contextual information.
It should reduce the time and token usage for initial and subsequent exploration by handling everything internally and passing it to AI agents.
 
# Table of Contents
- [Requirements](#requirements)
- [Install via go](#install-via-go)
- [Install via wget](#install-via-wget)
- [Build requirements](#build-requirements)
- [Build](#build-1)
- [Install go](#install-go)

# Getting Started

## Requirements
None windows system with `go` or `wget & tar` installed.

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
