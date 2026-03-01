# sota

CLI for [sota.io](https://sota.io) -- deploy web apps from the terminal.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/sota-deploy/cli)](https://github.com/sota-deploy/cli/releases/latest)

## Quick Start

```bash
# Install
curl -fsSL https://sota.io/install.sh | sh

# Login
sota login

# Deploy
cd my-app
sota deploy
```

## Installation

### Install Script (recommended)

```bash
curl -fsSL https://sota.io/install.sh | sh
```

The script detects your OS and architecture, downloads the correct binary, and places it in your PATH. It also handles upgrades in-place.

### Manual Download

Download the binary for your platform from [GitHub Releases](https://github.com/sota-deploy/cli/releases/latest):

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `sota-darwin-arm64` |
| macOS (Intel) | `sota-darwin-amd64` |
| Linux (x86_64) | `sota-linux-amd64` |
| Linux (ARM64) | `sota-linux-arm64` |

Then make it executable and move to your PATH:

```bash
chmod +x sota-darwin-arm64
sudo mv sota-darwin-arm64 /usr/local/bin/sota
```

### Build from Source

Prerequisites: [Go 1.25+](https://go.dev/dl/)

```bash
git clone https://github.com/sota-deploy/cli.git
cd cli
make build
```

The binary will be at `bin/sota`. To install to your GOPATH:

```bash
make install
```

To build for all platforms:

```bash
make build-all
```

Version injection via ldflags:

```bash
make build VERSION=1.0.0
```

## Authentication

### Device Code Flow (recommended)

```bash
sota login
```

This opens your browser with a device code. Enter the code at [sota.io/auth/device](https://sota.io/auth/device) to authenticate. No localhost redirect needed.

### API Key

Set an API key directly:

```bash
sota auth set-key <key>
```

Or via environment variable:

```bash
export SOTA_API_KEY=sota_your_api_key_here
```

### Check Status

```bash
sota auth status
```

## Commands

| Command | Description |
|---------|-------------|
| `sota login` | Authenticate via device code flow |
| `sota logout` | Remove stored credentials |
| `sota deploy` | Deploy current directory |
| `sota logs` | View build and runtime logs |
| `sota logs -f` | Follow logs in real-time |
| `sota env set KEY=VALUE` | Set environment variable |
| `sota env get KEY` | Get environment variable |
| `sota env list` | List all environment variables |
| `sota rollback` | Rollback to previous deployment |
| `sota status` | Show deployment status |
| `sota projects list` | List all projects |
| `sota projects create NAME` | Create a new project |
| `sota auth set-key KEY` | Set API key directly |
| `sota auth status` | Show auth status |

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--no-color` | Disable colored output | `false` |
| `--json` | Output in JSON format | `false` |
| `--api-url` | Custom API base URL | `https://api.sota.io` |

## Links

- [Website](https://sota.io)
- [Documentation](https://sota.io/docs)
- [Dashboard](https://sota.io/dashboard)
- [MCP Server](https://github.com/sota-deploy/mcp-server) -- AI agent integration
- [Issues](https://github.com/sota-deploy/cli/issues)

## License

MIT
