# alfred-cli

A command-line tool that uses Ollama to generate terminal commands from natural language queries. alfred-cli will prompt you for confirmation before executing any generated commands.

## Features

- Natural language to command conversion using Ollama
- Cross-platform support (Linux, macOS, Windows)
- Platform-aware command generation (PowerShell/cmd on Windows, bash/sh on Unix)
- Environment context gathering (OS, directory, shell, user)
- Configurable Ollama URL, port, and model
- Customizable system prompt
- Configuration via config file and environment variables
- Safe command execution with user confirmation

## Installation

### Prerequisites

- Go 1.21 or later
- Ollama installed and running
- An Ollama model (default: `llama3.2`)

### Build from Source

```bash
git clone <repository-url>
cd cli
go build -o alfred-cli .
```

### Install Globally

**Option 1: Quick install from GitHub (recommended)**

```bash
curl -fsSL https://raw.githubusercontent.com/ldiazderamond/alfred-cli/main/install.sh | sh
```

> **Note:** If you encounter caching issues, you can add a cache-busting parameter:
> ```bash
> curl -fsSL "https://raw.githubusercontent.com/ldiazderamond/alfred-cli/main/install.sh?v=$(date +%s)" | sh
> ```

This will automatically:
- Detect your platform (Linux/macOS, amd64/arm64)
- Download the appropriate pre-built binary from GitHub releases
- Install it to `/usr/local/bin`
- Create the default configuration file

If no pre-built binary is available for your platform, it will automatically build from source.

**Option 2: Using the install script locally**

```bash
./install.sh
```

**Option 3: Using Make**

```bash
make install
```

**Option 4: Manual installation**

```bash
go build -o alfred-cli .
sudo cp alfred-cli /usr/local/bin/
```

**Option 5: Using go install**

```bash
go install
```

### Cross-Platform Builds

Yes! Go supports cross-compilation natively. You can build Windows executables from Linux, macOS executables from Windows, etc.

**Using Makefile (recommended):**
```bash
make build-windows      # Build Windows executable (amd64)
make build-windows-arm  # Build Windows executable (arm64)
make build-macos        # Build macOS executable (amd64)
make build-macos-arm    # Build macOS executable (Apple Silicon)
make build-linux        # Build Linux executable (amd64)
make build-all          # Build for all platforms
```

**Manual cross-compilation:**
```bash
# Windows (from Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o alfred-cli.exe .

# macOS (from Linux/Windows)
GOOS=darwin GOARCH=amd64 go build -o alfred-cli-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o alfred-cli-darwin-arm64 .  # Apple Silicon

# Linux (from Windows/macOS)
GOOS=linux GOARCH=amd64 go build -o alfred-cli-linux-amd64 .
```

The binary will automatically detect the platform and use the appropriate shell:
- **Windows**: PowerShell (preferred) or cmd.exe
- **macOS/Linux**: bash or sh

## Configuration

### Config File

Create a configuration file at `~/.alfred-cli/config.yaml`:

```yaml
ollama:
  url: "http://192.168.1.117"
  port: 11434
  model: "llama3.2"
system_prompt: "You are a helpful assistant that generates terminal commands based on the user's query and the provided environment context. Use the environment information (platform, current directory, OS, shell) to generate accurate and platform-appropriate commands. For Windows, generate PowerShell or cmd.exe commands. For macOS and Linux, generate bash/sh commands. If multiple steps are needed, generate multiple commands separated by newlines (one command per line). Each command will be executed sequentially, so you can use separate commands for operations like changing directories. Only respond with the command(s), no explanations unless asked."
```

### Environment Variables

You can override any configuration using environment variables:

- `ALFRED_OLLAMA_URL` - Ollama server URL (default: `http://192.168.1.117`)
- `ALFRED_OLLAMA_PORT` - Ollama server port (default: `11434`)
- `ALFRED_OLLAMA_MODEL` - Ollama model to use (default: `llama3.2`)
- `ALFRED_SYSTEM_PROMPT` - System prompt for the AI model

Environment variables take precedence over config file values.

## Development

### Makefile Commands

The project includes a Makefile with useful commands:

```bash
make build           # Build the binary
make install         # Build and install globally
make uninstall       # Remove from /usr/local/bin
make clean           # Remove the built binary
make test            # Build and run a test query
make build-windows   # Build Windows executable
make build-macos     # Build macOS executable
make build-linux     # Build Linux executable
make build-all       # Build for all platforms
make help            # Show all available commands
```

### Creating Releases

To create a new release with pre-built binaries:

1. Create and push a version tag (following semantic versioning):
   ```bash
   # For a pre-release (recommended for initial releases):
   git tag v0.1.0-alpha
   git push origin v0.1.0-alpha
   
   # Or for a stable release:
   git tag v0.1.0
   git push origin v0.1.0
   ```

   **Versioning guidelines:**
   - Start with pre-releases like `v0.1.0-alpha`, `v0.1.0-beta`, or `v0.1.0-rc.1` for initial testing
   - Use `v0.1.0` for the first stable release
   - Use `v0.x.x` for pre-1.0 releases (may have breaking changes)
   - Use `v1.0.0` when the API is stable and production-ready
   - Follow semantic versioning: `MAJOR.MINOR.PATCH[-PRERELEASE]`
   - Pre-release examples: `v0.1.0-alpha`, `v0.1.0-beta.1`, `v0.1.0-rc.1`

2. GitHub Actions will automatically:
   - Build binaries for all platforms (Linux, macOS, Windows, both amd64 and arm64)
   - Create a GitHub release with all binaries attached

3. Users can then install using:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/ldiazderamond/alfred-cli/main/install.sh | sh
   ```

The install script will automatically download the latest release binary for their platform.

## Usage

### Basic Usage

```bash
alfred-cli "give me a command to initialise a new quarkus project called 'kkpo'"
```

### Examples

```bash
# Generate a command to create a new directory
alfred-cli "create a directory called myproject"

# Generate a command to list files
alfred-cli "show me all python files in the current directory"

# Generate a command with specific requirements
alfred-cli "give me a command to clone a git repository from github"
```

## How It Works

1. You provide a natural language query describing what command you need
2. alfred-cli sends your query to Ollama with a system prompt
3. Ollama generates the appropriate terminal command
4. alfred-cli displays the generated command and asks for confirmation
5. If confirmed, the command is executed in your terminal

## Troubleshooting

### Ollama Connection Issues

If you get connection errors, make sure:
- Ollama is running: `ollama serve`
- The URL and port in your config match your Ollama setup
- The specified model is available: `ollama list`

### Model Not Found

If the specified model is not available, pull it first:

```bash
ollama pull llama3.2
```

### Command Generation Issues

If commands are not generated correctly, try:
- Adjusting the system prompt in your config
- Using a different model
- Being more specific in your query

## License

[Specify your license here]
