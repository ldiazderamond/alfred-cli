package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ollama       OllamaConfig `yaml:"ollama"`
	SystemPrompt string       `yaml:"system_prompt"`
}

type OllamaConfig struct {
	URL   string `yaml:"url"`
	Port  int    `yaml:"port"`
	Model string `yaml:"model"`
}

var defaultSystemPrompt = "You are a helpful assistant that generates terminal commands based on the user's query and the provided environment context. Use the environment information (platform, current directory, OS, shell) to generate accurate and platform-appropriate commands. For Windows, generate PowerShell or cmd.exe commands. For macOS and Linux, generate bash/sh commands. If multiple steps are needed, generate multiple commands separated by newlines (one command per line). Each command will be executed sequentially, so you can use separate commands for operations like changing directories. Only respond with the command(s), no explanations unless asked."

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".alfred-cli")
	configPath := filepath.Join(configDir, "config.yaml")

	// Start with defaults
	cfg := &Config{
		Ollama: OllamaConfig{
			URL:   "http://192.168.1.117",
			Port:  11434,
			Model: "llama3.2",
		},
		SystemPrompt: defaultSystemPrompt,
	}

	// Read config file if it exists
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Override with environment variables if set
	if url := os.Getenv("ALFRED_OLLAMA_URL"); url != "" {
		cfg.Ollama.URL = url
	}
	if portStr := os.Getenv("ALFRED_OLLAMA_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Ollama.Port = port
		}
	}
	if model := os.Getenv("ALFRED_OLLAMA_MODEL"); model != "" {
		cfg.Ollama.Model = model
	}
	if prompt := os.Getenv("ALFRED_SYSTEM_PROMPT"); prompt != "" {
		cfg.SystemPrompt = prompt
	}

	// Ensure system prompt has a value
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = defaultSystemPrompt
	}

	return cfg, nil
}

func GetOllamaURL(cfg *Config) string {
	port := cfg.Ollama.Port
	if port == 0 {
		port = 11434
	}
	return fmt.Sprintf("%s:%d", cfg.Ollama.URL, port)
}
