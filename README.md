# nvim-mindevc

A Go CLI tool that automates the setup of Neovim inside Docker devcontainers. It downloads, configures, and compiles Neovim with all necessary development tools.

Demo:

<a href="https://asciinema.org/a/juXKraph4GARMnTtfsGhVyd44" target="_blank"><img src="https://asciinema.org/a/juXKraph4GARMnTtfsGhVyd44.svg" /></a>

## Features

- **Automated Neovim Setup**: Downloads and compiles Neovim nightly builds inside devcontainers
- **Tool Management**: Installs essential development tools (ripgrep, fd, zig, etc.) with multi-architecture support
- **devcontainer.json**: Basic integration with `devcontainer.json` file. For now only supports configs using `dockerComposeFile`
- **Flexible Configuration**: Hierarchical configuration with multiple sources
- **Cross-Platform**: Supports x86_64 and aarch64 architectures


## Installation


### Pre-built binaries

See the releases page: https://github.com/davidrios/nvim-mindevc/releases


### From Source

Prerequisites:

- Go 1.24.3 or later

```bash
git clone https://github.com/davidrios/nvim-mindevc.git
cd nvim-mindevc
go build -o nvim-mindevc
sudo mv nvim-mindevc /usr/local/bin
```

## Quick Start

```bash
# Setup Neovim in your devcontainer
# cd to folder where there's a .devcontainer.json or .devcontainer/devcontainer.json file
nvim-mindevc setup

# Show current configuration
nvim-mindevc show-config

# Enable verbose logging
nvim-mindevc -v setup
```


## Configuration

nvim-mindevc uses a hierarchical configuration system with the following precedence:

1. **Command-line**: `-c path/to/config.yaml` and/or `-d path/to/devcontainer.json`
2. **Environment variables** (with `NVIM_MINDEVC_` prefix)
3. **Configuration files**:
   - `./.nvim-mindevc.yaml`
   - `.devcontainer/nvim-mindevc.yaml`
   - `~/.config/nvim-mindevc.yaml`
4. **Built-in defaults** (lowest priority)

### Example Configuration

```yaml
cacheDir: "~/.cache/nvim-mindevc"
installTools:
  - zig
  - ripgrep
  - fd

neovim:
  configURI: "file://~/.config/nvim"
  runscript: "/opt/nvim-mindevc/bin/nvim"

remote:
  workdir: "/opt/nvim-mindevc"
```

## Usage

### Basic Setup

```bash
# Setup with default configuration
nvim-mindevc setup

# Use custom config file
nvim-mindevc -c custom-config.yaml setup

# Specify devcontainer file
nvim-mindevc -d .devcontainer/devcontainer.json setup
```

### Configuration Management

```bash
# View current configuration
nvim-mindevc show-config

# Debug configuration loading
nvim-mindevc -v show-config
```

### Git Emulation

The tool includes built-in git commands for use within devcontainers, so no git installation is required for basic plugin managers like Lazy:

```bash
# symlink to `git`
ln -s nvim-mindevc git
./git log
./git fetch
./git checkout <branch>
./git ls-files
```


### Key Components

- **Configuration System**: Uses Viper for hierarchical configuration loading
- **Docker Integration**: Executes commands inside devcontainers via Docker Compose
- **Tool Management**: Downloads, extracts, and links development tools with SHA256 verification


## Development

### Running Tests

```bash
# Run all tests
go test ./... -short

# Run specific package tests
go test ./setup/ -short

# Test with verbose output
go test -v ./... -short
```


### Testing with Multiple Distributions

The project includes test containers for various Linux distributions:

- Ubuntu (16.04, 18.04, 22.04, 24.04)
- Debian (Buster, Bookworm)
- Fedora
- CentOS
- Alpine (3.20, 3.21)


## Remote Tools

Tools that will be installed by default in the container:

- **Neovim**: Latest nightly builds, compiled from source
- **Zig**: Cross-platform build system
- **ripgrep**: Fast text search tool
- **fd**: Modern find replacement
- **make**: Static build of GNU Make

It's also possible to configure custom tools.


## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.
