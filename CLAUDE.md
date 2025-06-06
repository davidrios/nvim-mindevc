# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

nvim-mindevc is a Go CLI tool that automates the setup of Neovim development environments inside Docker containers (devcontainers). It downloads and configures Neovim and related tools with proper security verification.

## Development Commands

```bash
# Build the project
go build -o nvim-mindevc

# Run all tests
go test ./...

# Run specific package tests
go test ./setup/

# Show current configuration (useful for debugging)
./nvim-mindevc show-config

# Run setup with verbose logging
./nvim-mindevc -v setup

# Build test tools (requires network access)
./scripts/compile-zig.sh
```

## Architecture

The codebase follows a clean modular design:

- **cmd/** - Cobra CLI commands (root, setup, show-config)
- **config/** - Configuration loading with hierarchical precedence: flags > env vars > config files > defaults
- **docker/** - Docker Compose integration for container operations
- **setup/** - Core installation logic with secure tool downloading and SHA256 verification

## Key Implementation Details

### Configuration System
Uses Viper with hierarchical loading:
1. Command-line flags (highest priority)
2. Environment variables (NVIM_MINDEVC_ prefix)
3. Config files: `./.nvim-mindevc.yaml` → `.devcontainer/nvim-mindevc.yaml` → `~/.config/nvim-mindevc.yaml`
4. Hardcoded defaults (lowest priority)

### Security Model
All tool downloads require SHA256 hash verification. The system supports multi-architecture binaries (x86_64, aarch64) with architecture-specific hashes defined in config.

### Docker Integration
Heavy integration with Docker Compose for executing commands inside devcontainers. Uses `docker compose ps` and `docker compose exec` patterns.

### Tool Management
Downloads tools to `~/.cache/nvim-mindevc/` and creates symlinks to `/usr/local/bin/` or custom locations. Implements proper cleanup and extraction logic.

## Code style
- Do not add comments for trivial lines of code. Add comments explaining why, not what's being done.

## Testing Infrastructure

- Comprehensive unit tests in `setup/tools_test.go` with HTTP mocking
- Test containers for multiple Linux distributions in `testdata/test_project/`
- Mock Zig implementations of tools (fd, rg) for integration testing

## Known Implementation Gaps

The `ExtractAndLinkTool` function in `setup/tools.go:` is currently a stub and needs full implementation for archive extraction and symlink creation.
