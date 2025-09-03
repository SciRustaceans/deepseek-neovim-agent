package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"deepseek-nvim-agent/config"
)

type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiKey:  cfg.DeepSeekAPIKey,
		baseURL: cfg.DeepSeekBaseURL,
		client:  &http.Client{},
	}
}

func (c *Client) ChatCompletion(messages []Message) (string, error) {
	request := ChatCompletionRequest{
		Model:    "deepseek-coder",
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *Client) GenerateCodeEdit(prompt, currentCode string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: "You are an expert programming assistant. Always respond with only the modified code, no explanations. If you need to show changes, use code comments.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Current code:\n```\n%s\n```\n\nTask: %s\n\nReturn only the modified code:", currentCode, prompt),
		},
	}

	return c.ChatCompletion(messages)
}

func (c *Client) ExplainCode(code, language string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: "You are an expert programming teacher. Explain the code clearly and concisely.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Explain this %s code:\n```%s\n%s\n```", language, language, code),
		},
	}

	return c.ChatCompletion(messages)
}
