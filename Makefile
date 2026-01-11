.PHONY: build install uninstall clean help

BINARY_NAME=alfred-cli
INSTALL_PATH=/usr/local/bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "-s -w -X 'main.version=dev' -X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BINARY_NAME) .

install: build ## Build and install to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	@echo "Creating config directory..."
	@mkdir -p ~/.alfred-cli
	@if [ ! -f ~/.alfred-cli/config.yaml ]; then \
		echo "Creating default config file..."; \
		echo 'ollama:' > ~/.alfred-cli/config.yaml; \
		echo '  url: "http://192.168.1.117"' >> ~/.alfred-cli/config.yaml; \
		echo '  port: 11434' >> ~/.alfred-cli/config.yaml; \
		echo '  model: "llama3.2"' >> ~/.alfred-cli/config.yaml; \
		echo 'system_prompt: "You are a helpful assistant that generates terminal commands based on the user'\''s query and the provided environment context. Use the environment information (platform, current directory, OS, shell) to generate accurate and platform-appropriate commands. For Windows, generate PowerShell or cmd.exe commands. For macOS and Linux, generate bash/sh commands. If multiple steps are needed, generate multiple commands separated by newlines (one command per line). Each command will be executed sequentially, so you can use separate commands for operations like changing directories. Only respond with the command(s), no explanations unless asked."' >> ~/.alfred-cli/config.yaml; \
		echo "Config file created at ~/.alfred-cli/config.yaml"; \
	else \
		echo "Config file already exists (skipping)"; \
	fi
	@echo ""
	@echo "Installation complete!"
	@echo "You can now use '$(BINARY_NAME)' from anywhere."

uninstall: ## Remove the binary from /usr/local/bin
	@echo "Removing $(BINARY_NAME) from $(INSTALL_PATH)..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete!"

clean: ## Remove the built binary
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	@echo "Clean complete!"

test: build ## Build and run a test query
	@echo "Running test query..."
	./$(BINARY_NAME) "list files in current directory"

build-windows: ## Build Windows executable (amd64)
	@echo "Building Windows executable..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X 'main.version=dev' -X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BINARY_NAME).exe .
	@echo "Windows executable created: $(BINARY_NAME).exe"

build-windows-arm: ## Build Windows executable (arm64)
	@echo "Building Windows ARM64 executable..."
	@GOOS=windows GOARCH=arm64 go build -ldflags "-s -w -X 'main.version=dev' -X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BINARY_NAME)-arm64.exe .
	@echo "Windows ARM64 executable created: $(BINARY_NAME)-arm64.exe"

build-macos: ## Build macOS executable (amd64)
	@echo "Building macOS executable..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X 'main.version=dev' -X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BINARY_NAME)-darwin-amd64 .
	@echo "macOS executable created: $(BINARY_NAME)-darwin-amd64"

build-macos-arm: ## Build macOS executable (Apple Silicon)
	@echo "Building macOS ARM64 executable..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X 'main.version=dev' -X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BINARY_NAME)-darwin-arm64 .
	@echo "macOS ARM64 executable created: $(BINARY_NAME)-darwin-arm64"

build-linux: ## Build Linux executable (amd64)
	@echo "Building Linux executable..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X 'main.version=dev' -X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BINARY_NAME)-linux-amd64 .
	@echo "Linux executable created: $(BINARY_NAME)-linux-amd64"

build-all: ## Build executables for all platforms
	@echo "Building for all platforms..."
	@$(MAKE) build-windows
	@$(MAKE) build-macos
	@$(MAKE) build-macos-arm
	@$(MAKE) build-linux
	@echo "All builds complete!"
