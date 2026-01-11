package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Ollama       OllamaConfig `mapstructure:"ollama"`
	SystemPrompt string       `mapstructure:"system_prompt"`
}

type OllamaConfig struct {
	URL   string `mapstructure:"url"`
	Port  int    `mapstructure:"port"`
	Model string `mapstructure:"model"`
}

var defaultSystemPrompt = "You are a helpful assistant that generates terminal commands based on the user's query and the provided environment context. Use the environment information (platform, current directory, OS, shell) to generate accurate and platform-appropriate commands. For Windows, generate PowerShell or cmd.exe commands. For macOS and Linux, generate bash/sh commands. If multiple steps are needed, generate multiple commands separated by newlines (one command per line). Each command will be executed sequentially, so you can use separate commands for operations like changing directories. Only respond with the command(s), no explanations unless asked."

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".alfred-cli")

	v := viper.New()

	// Set defaults
	v.SetDefault("ollama.url", "http://192.168.1.117")
	v.SetDefault("ollama.port", 11434)
	v.SetDefault("ollama.model", "llama3.2")
	v.SetDefault("system_prompt", defaultSystemPrompt)

	// Set config file path
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is okay, we'll use defaults
	}

	// Set environment variable support
	v.SetEnvPrefix("ALFRED")
	v.AutomaticEnv()
	v.BindEnv("ollama.url", "ALFRED_OLLAMA_URL")
	v.BindEnv("ollama.port", "ALFRED_OLLAMA_PORT")
	v.BindEnv("ollama.model", "ALFRED_OLLAMA_MODEL")
	v.BindEnv("system_prompt", "ALFRED_SYSTEM_PROMPT")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Ensure system prompt has a value
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = defaultSystemPrompt
	}

	return &cfg, nil
}

func GetOllamaURL(cfg *Config) string {
	port := cfg.Ollama.Port
	if port == 0 {
		port = 11434
	}
	return fmt.Sprintf("%s:%d", cfg.Ollama.URL, port)
}
