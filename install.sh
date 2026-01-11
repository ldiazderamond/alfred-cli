#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="ldiazderamond/alfred-cli"
BINARY_NAME="alfred-cli"
INSTALL_PATH="/usr/local/bin"
VERSION="${VERSION:-latest}"

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        *)          printf "${RED}Error: Unsupported operating system: $(uname -s)${NC}\n" >&2; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)             printf "${RED}Error: Unsupported architecture: $(uname -m)${NC}\n" >&2; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Download binary from GitHub releases
download_binary() {
    local platform=$1
    local version=$2
    local download_url=""
    
    if [ "$version" = "latest" ]; then
        # Try to get latest stable release first
        local latest_tag=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        
        # If no stable release, try to get the latest release (including pre-releases)
        if [ -z "$latest_tag" ]; then
            latest_tag=$(curl -s "https://api.github.com/repos/${REPO}/releases" | grep '"tag_name":' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')
        fi
        
        if [ -z "$latest_tag" ]; then
            printf "${RED}Error: Could not fetch latest release. Make sure releases are available on GitHub.${NC}\n" >&2
            printf "${YELLOW}Falling back to building from source...${NC}\n" >&2
            return 1
        fi
        version=$latest_tag
    fi
    
    # Construct download URL
    if [ "$platform" = "linux-amd64" ]; then
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-linux-amd64"
    elif [ "$platform" = "linux-arm64" ]; then
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-linux-arm64"
    elif [ "$platform" = "darwin-amd64" ]; then
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-darwin-amd64"
    elif [ "$platform" = "darwin-arm64" ]; then
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-darwin-arm64"
    else
        printf "${RED}Error: Unsupported platform: ${platform}${NC}\n" >&2
        return 1
    fi
    
    printf "${GREEN}Downloading ${BINARY_NAME} ${version} for ${platform}...${NC}\n" >&2
    
    # Download to temporary location
    local tmp_file=$(mktemp)
    if ! curl -fsSL "$download_url" -o "$tmp_file"; then
        printf "${RED}Error: Failed to download binary from GitHub releases.${NC}\n" >&2
        printf "${YELLOW}This might mean:${NC}\n" >&2
        printf "${YELLOW}  1. No releases are available yet${NC}\n" >&2
        printf "${YELLOW}  2. The release doesn't have a binary for your platform${NC}\n" >&2
        printf "${YELLOW}Falling back to building from source...${NC}\n" >&2
        rm -f "$tmp_file"
        return 1
    fi
    
    # Make it executable
    chmod +x "$tmp_file"
    echo "$tmp_file"
}

# Build from source
build_from_source() {
    printf "${YELLOW}Building ${BINARY_NAME} from source...${NC}\n"
    
    if ! command -v go &> /dev/null; then
        printf "${RED}Error: Go is not installed. Please install Go 1.21 or later.${NC}\n" >&2
        exit 1
    fi
    
    # Detect if we're on Linux and set CGO_ENABLED=0 for static linking
    local cgo_enabled=""
    if [ "$(uname -s)" = "Linux" ]; then
        cgo_enabled="CGO_ENABLED=0 "
        printf "${YELLOW}Building with static linking for Linux compatibility...${NC}\n"
    fi
    
    # Check if we're in a git repository (local install)
    if [ -f "go.mod" ]; then
        local tmp_file=$(mktemp)
        ${cgo_enabled}go build -ldflags "-s -w" -o "$tmp_file" .
        chmod +x "$tmp_file"
        echo "$tmp_file"
    else
        # We're being run via curl | sh, need to clone the repo
        printf "${YELLOW}Cloning repository to build from source...${NC}\n"
        local tmp_dir=$(mktemp -d)
        trap "rm -rf $tmp_dir" EXIT
        
        git clone --depth 1 "https://github.com/${REPO}.git" "$tmp_dir" || {
            printf "${RED}Error: Failed to clone repository. Please install Go and clone the repository manually.${NC}\n" >&2
            exit 1
        }
        
        cd "$tmp_dir"
        local tmp_file=$(mktemp)
        ${cgo_enabled}go build -ldflags "-s -w" -o "$tmp_file" .
        chmod +x "$tmp_file"
        echo "$tmp_file"
    fi
}

# Main installation function
main() {
    printf "${GREEN}Installing ${BINARY_NAME}...${NC}\n"
    printf "\n"
    
    # Detect platform
    local platform=$(detect_platform)
    printf "Detected platform: ${platform}\n"
    
    # Try to download binary, fall back to building from source
    local binary_path=""
    if binary_path=$(download_binary "$platform" "$VERSION"); then
        printf "${GREEN}Successfully downloaded binary${NC}\n"
    else
        printf "${YELLOW}Building from source instead...${NC}\n"
        binary_path=$(build_from_source)
    fi
    
    # Install to /usr/local/bin
    printf "${GREEN}Installing ${BINARY_NAME} to ${INSTALL_PATH}...${NC}\n"
    if [ -w "$INSTALL_PATH" ]; then
        cp "$binary_path" "${INSTALL_PATH}/${BINARY_NAME}"
    else
        sudo cp "$binary_path" "${INSTALL_PATH}/${BINARY_NAME}"
    fi
    
    # Clean up temporary file
    rm -f "$binary_path"
    
    # Create config directory
    printf "${GREEN}Creating config directory...${NC}\n"
    mkdir -p ~/.alfred-cli
    
    CONFIG_FILE=~/.alfred-cli/config.yaml
    if [ ! -f "$CONFIG_FILE" ]; then
        printf "${GREEN}Creating default config file...${NC}\n"
        cat > "$CONFIG_FILE" << 'EOF'
ollama:
  url: "http://192.168.1.117"
  port: 11434
  model: "llama3.2"
system_prompt: "You are a helpful assistant that generates terminal commands based on the user's query and the provided environment context. Use the environment information (platform, current directory, OS, shell) to generate accurate and platform-appropriate commands. For Windows, generate PowerShell or cmd.exe commands. For macOS and Linux, generate bash/sh commands. If multiple steps are needed, generate multiple commands separated by newlines (one command per line). Each command will be executed sequentially, so you can use separate commands for operations like changing directories. Only respond with the command(s), no explanations unless asked."
EOF
        printf "${GREEN}Config file created at $CONFIG_FILE${NC}\n"
    else
        printf "${YELLOW}Config file already exists at $CONFIG_FILE (skipping)${NC}\n"
    fi
    
    printf "\n"
    printf "${GREEN}Installation complete!${NC}\n"
    printf "\n"
    printf "You can now use '${BINARY_NAME}' from anywhere.\n"
    printf "\n"
    printf "Configuration file: $CONFIG_FILE\n"
    printf "You can edit it or use environment variables to override settings.\n"
    printf "See README.md for configuration options.\n"
}

# Run main function
main
