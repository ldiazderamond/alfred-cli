package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

type Client struct {
	BaseURL string
	Model   string
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
	}
}

func (c *Client) GenerateCommand(systemPrompt, userQuery, context string) (string, error) {
	// Build the full prompt with context
	var fullPrompt string
	if context != "" {
		fullPrompt = fmt.Sprintf("%s\n\n%s\n\nUser: %s\nAssistant:", systemPrompt, context, userQuery)
	} else {
		fullPrompt = fmt.Sprintf("%s\n\nUser: %s\nAssistant:", systemPrompt, userQuery)
	}

	reqBody := GenerateRequest{
		Model:  c.Model,
		Prompt: fullPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var generateResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if generateResp.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", generateResp.Error)
	}

	// Clean up the response - remove leading/trailing whitespace and newlines
	command := strings.TrimSpace(generateResp.Response)
	
	// Remove markdown code blocks if present
	command = strings.TrimPrefix(command, "```")
	command = strings.TrimPrefix(command, "```bash")
	command = strings.TrimPrefix(command, "```sh")
	command = strings.TrimSuffix(command, "```")
	command = strings.TrimSpace(command)

	return command, nil
}

func (c *Client) GenerateFix(originalQuery, failedCommand, errorOutput string) (string, error) {
	// Build a prompt asking for a fix
	fixPrompt := fmt.Sprintf(`The following command failed:
Command: %s
Error output: %s

Original request: %s

Please generate a corrected command that will fix the issue. Only respond with the corrected command(s), no explanations.`, failedCommand, errorOutput, originalQuery)

	reqBody := GenerateRequest{
		Model:  c.Model,
		Prompt: fixPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var generateResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if generateResp.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", generateResp.Error)
	}

	// Clean up the response
	command := strings.TrimSpace(generateResp.Response)
	command = strings.TrimPrefix(command, "```")
	command = strings.TrimPrefix(command, "```bash")
	command = strings.TrimPrefix(command, "```sh")
	command = strings.TrimSuffix(command, "```")
	command = strings.TrimSpace(command)

	return command, nil
}
