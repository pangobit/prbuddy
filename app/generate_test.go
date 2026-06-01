package app

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/pangobit/prbuddy/internal/llm"
	"github.com/pangobit/prbuddy/internal/review"
)

func TestGenerateReviewUsesProviderAndRendersMarkdown(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	collector := &fakeReviewCollector{
		context: review.Context{
			Repository: review.Repository{Name: "prbuddy"},
		},
	}
	factory := &fakeLLMFactory{
		client: fakeLLMClient{markdown: "## Review\n\nLooks good."},
	}
	renderer := fakeDocumentRenderer{}
	env := MapEnvironment{
		"GEMINI_API_KEY":       "gemini-key",
		"GITHUB_ACTIONS":       "true",
		"GITHUB_BASE_REF":      "main",
		"GITHUB_HEAD_REF":      "feature",
		"GITHUB_EVENT_PATH":    "event.json",
		"GITHUB_REPOSITORY":    "pangobit/prbuddy",
		"GITHUB_SHA":           "abc123",
		"GITHUB_EVENT_NAME":    "pull_request",
		"GITHUB_SERVER_URL":    "https://github.com",
		"PRBUDDY_GEMINI_MODEL": "gemini-test",
	}
	useCase := NewGenerateReviewUseCase(collector, factory, renderer, env)

	err := useCase.GenerateReview(context.Background(), &stdout, GenerateReviewOptions{
		RepositoryPath: ".",
		Provider:       "gemini",
		ProviderURL:    "https://ai",
	})
	if err != nil {
		t.Fatalf("generate review: %v", err)
	}

	if collector.got.BaseRef != "origin/main" {
		t.Fatalf("base ref = %q", collector.got.BaseRef)
	}

	if collector.got.HeadRef != "HEAD" {
		t.Fatalf("head ref = %q", collector.got.HeadRef)
	}

	if collector.got.EventPath != "event.json" {
		t.Fatalf("event path = %q", collector.got.EventPath)
	}

	if factory.got.APIKey != "gemini-key" {
		t.Fatalf("api key = %q", factory.got.APIKey)
	}

	if factory.got.Model != "gemini-test" {
		t.Fatalf("model = %q", factory.got.Model)
	}

	if factory.got.ProviderURL != "https://ai" {
		t.Fatalf("provider url = %q", factory.got.ProviderURL)
	}

	if !strings.Contains(factory.client.got.User, "Pull Request Review Context") {
		t.Fatalf("prompt user = %q", factory.client.got.User)
	}

	if got := stdout.String(); got != "## Review\n\nLooks good." {
		t.Fatalf("stdout = %q", got)
	}
}

func TestGenerateReviewFailsBeforeCollectingWithoutProviderKey(t *testing.T) {
	t.Parallel()

	collector := &fakeReviewCollector{}
	useCase := NewGenerateReviewUseCase(collector, &fakeLLMFactory{}, fakeDocumentRenderer{}, MapEnvironment{})

	err := useCase.GenerateReview(context.Background(), io.Discard, GenerateReviewOptions{
		Provider: "grok",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "XAI_API_KEY must be set") {
		t.Fatalf("error = %q", err)
	}

	if collector.called {
		t.Fatal("collector was called")
	}
}

func TestGenerateReviewAllowsGeminiCustomURLWithoutProviderKey(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	collector := &fakeReviewCollector{}
	factory := &fakeLLMFactory{
		client: fakeLLMClient{markdown: "## Review"},
	}
	useCase := NewGenerateReviewUseCase(collector, factory, fakeDocumentRenderer{}, MapEnvironment{})

	err := useCase.GenerateReview(context.Background(), &stdout, GenerateReviewOptions{
		Provider:    "gemini",
		ProviderURL: "https://ai",
	})
	if err != nil {
		t.Fatalf("generate review: %v", err)
	}

	if factory.got.APIKey != "" {
		t.Fatalf("api key = %q", factory.got.APIKey)
	}

	if factory.got.ProviderURL != "https://ai" {
		t.Fatalf("provider url = %q", factory.got.ProviderURL)
	}
}

func TestGenerateReviewRejectsUnsupportedProvider(t *testing.T) {
	t.Parallel()

	useCase := NewGenerateReviewUseCase(&fakeReviewCollector{}, &fakeLLMFactory{}, fakeDocumentRenderer{}, MapEnvironment{})

	err := useCase.GenerateReview(context.Background(), io.Discard, GenerateReviewOptions{
		Provider: "other",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("error = %q", err)
	}
}

func TestBuildPromptRequestsRichMarkdownReviewDocument(t *testing.T) {
	t.Parallel()

	prompt := BuildPrompt(review.Context{})
	required := []string{
		"Return Markdown only.",
		"sitting next to another engineer",
		"walking them through the code",
		"Show, do not tell.",
		"Tell the story of the code",
		"Do not repeat what the reader can already see in the diff",
		"fenced code snippets",
		"Mermaid diagrams",
		"flowchart TD",
		"Side effects and risks",
		"Review comments",
		"Suggested comment",
	}

	for _, want := range required {
		if !strings.Contains(prompt.System+"\n"+prompt.User, want) {
			t.Fatalf("prompt missing %q in:\n%s\n%s", want, prompt.System, prompt.User)
		}
	}
}

type fakeReviewCollector struct {
	context review.Context
	got     review.CollectOptions
	called  bool
}

func (c *fakeReviewCollector) Collect(_ context.Context, opts review.CollectOptions) (review.Context, error) {
	c.got = opts
	c.called = true
	return c.context, nil
}

type fakeLLMFactory struct {
	client fakeLLMClient
	got    ProviderConfig
}

func (f *fakeLLMFactory) Client(_ context.Context, config ProviderConfig) (LLMClient, error) {
	f.got = config
	return &f.client, nil
}

type fakeLLMClient struct {
	markdown string
	got      llm.Prompt
}

func (c *fakeLLMClient) GenerateMarkdown(_ context.Context, prompt llm.Prompt) (string, error) {
	c.got = prompt
	return c.markdown, nil
}

type fakeDocumentRenderer struct{}

func (r fakeDocumentRenderer) Render(_ context.Context, out io.Writer, _ review.Context, markdown string) error {
	_, err := io.WriteString(out, markdown)
	return err
}
