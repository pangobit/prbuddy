// Package gemini adapts Google's Gemini SDK to PR Buddy's app port.
package gemini

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/genai"

	"github.com/pangobit/prbuddy/internal/llm"
)

const defaultModel = "gemini-3.5-flash"
const proxyPlaceholderKey = "prbuddy-proxy"

// Client generates review markdown with Gemini.
type Client struct {
	client *genai.Client
	model  string
}

// NewClient creates a Gemini-backed LLM client.
func NewClient(ctx context.Context, apiKey string, model string, providerURL string) (*Client, error) {
	if providerURL != "" && apiKey == "" {
		apiKey = proxyPlaceholderKey
	}

	config := &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	}
	if providerURL != "" {
		config.HTTPOptions = genai.HTTPOptions{BaseURL: providerURL}
	}

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		model:  defaultString(model, defaultModel),
	}, nil
}

// GenerateMarkdown generates markdown review content.
func (c *Client) GenerateMarkdown(ctx context.Context, prompt llm.Prompt) (string, error) {
	response, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(joinPrompt(prompt)), nil)
	if err != nil {
		return "", err
	}

	text := strings.TrimSpace(response.Text())
	if text == "" {
		return "", errors.New("gemini returned empty content")
	}

	return text, nil
}

func joinPrompt(prompt llm.Prompt) string {
	return prompt.System + "\n\n" + prompt.User
}

func defaultString(value string, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}
