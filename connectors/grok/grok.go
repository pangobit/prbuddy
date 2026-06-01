// Package grok adapts xAI's OpenAI-compatible API to PR Buddy's app port.
package grok

import (
	"context"
	"errors"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/pangobit/prbuddy/internal/llm"
)

const (
	defaultBaseURL = "https://api.x.ai/v1"
	defaultModel   = "grok-4.3-latest"
)

// Client generates review markdown with Grok.
type Client struct {
	client openai.Client
	model  string
}

// NewClient creates a Grok-backed LLM client.
func NewClient(apiKey string, model string) *Client {
	return &Client{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(defaultBaseURL),
		),
		model: defaultString(model, defaultModel),
	}
}

// GenerateMarkdown generates markdown review content.
func (c *Client) GenerateMarkdown(ctx context.Context, prompt llm.Prompt) (string, error) {
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.DeveloperMessage(prompt.System),
			openai.UserMessage(prompt.User),
		},
		Model: openai.ChatModel(c.model),
	})
	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", errors.New("grok returned no choices")
	}

	text := strings.TrimSpace(completion.Choices[0].Message.Content)
	if text == "" {
		return "", errors.New("grok returned empty content")
	}

	return text, nil
}

func defaultString(value string, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}
