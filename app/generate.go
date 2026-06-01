// Package app coordinates PR Buddy use cases.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/pangobit/prbuddy/internal/llm"
	"github.com/pangobit/prbuddy/internal/review"
)

// Provider identifies the selected LLM provider.
type Provider string

const (
	// ProviderGemini selects the Gemini adapter.
	ProviderGemini Provider = "gemini"
	// ProviderGrok selects the Grok adapter.
	ProviderGrok Provider = "grok"
)

// GenerateReviewOptions controls one review generation run.
type GenerateReviewOptions struct {
	// RepositoryPath is the filesystem path to the git repository.
	RepositoryPath string
	// BaseRef is the optional base git ref for the review.
	BaseRef string
	// HeadRef is the optional head git ref for the review.
	HeadRef string
	// EventPath is the optional path to a GitHub Actions event payload.
	EventPath string
	// Provider is the required LLM provider name.
	Provider string
	// Model is an optional provider model override.
	Model string
	// ProviderURL is an optional provider endpoint override.
	ProviderURL string
}

// ReviewCollector builds review context for a render request.
type ReviewCollector interface {
	Collect(context.Context, review.CollectOptions) (review.Context, error)
}

// LLMClient generates review content from a prompt.
type LLMClient interface {
	GenerateMarkdown(context.Context, llm.Prompt) (string, error)
}

// LLMFactory creates a provider client for a generation run.
type LLMFactory interface {
	Client(context.Context, ProviderConfig) (LLMClient, error)
}

// LLMFactoryFunc adapts a function into an LLM factory.
type LLMFactoryFunc func(context.Context, ProviderConfig) (LLMClient, error)

// Client creates a provider client for a generation run.
func (f LLMFactoryFunc) Client(ctx context.Context, config ProviderConfig) (LLMClient, error) {
	return f(ctx, config)
}

// ProviderConfig contains provider adapter configuration.
type ProviderConfig struct {
	// Provider is the selected LLM provider.
	Provider Provider
	// APIKey is the required provider credential.
	APIKey string
	// Model is the optional provider model override.
	Model string
	// ProviderURL is the optional provider endpoint override.
	ProviderURL string
}

// DocumentRenderer writes a complete HTML review document.
type DocumentRenderer interface {
	Render(context.Context, io.Writer, review.Context, string) error
}

// GenerateReviewUseCase coordinates context collection, LLM generation, and rendering.
type GenerateReviewUseCase struct {
	collector ReviewCollector
	factory   LLMFactory
	renderer  DocumentRenderer
	env       Environment
}

// NewGenerateReviewUseCase creates the review generation use case.
func NewGenerateReviewUseCase(collector ReviewCollector, factory LLMFactory, renderer DocumentRenderer, env Environment) *GenerateReviewUseCase {
	return &GenerateReviewUseCase{
		collector: collector,
		factory:   factory,
		renderer:  renderer,
		env:       env,
	}
}

// GenerateReview writes one complete HTML review document to out.
func (u *GenerateReviewUseCase) GenerateReview(ctx context.Context, out io.Writer, opts GenerateReviewOptions) error {
	resolved, err := u.resolveOptions(opts)
	if err != nil {
		return err
	}

	reviewContext, err := u.collector.Collect(ctx, review.CollectOptions{
		RepositoryPath: resolved.RepositoryPath,
		BaseRef:        resolved.BaseRef,
		HeadRef:        resolved.HeadRef,
		EventPath:      resolved.EventPath,
		CI:             resolved.CI,
	})
	if err != nil {
		return fmt.Errorf("collect review context: %w", err)
	}

	client, err := u.factory.Client(ctx, ProviderConfig{
		Provider:    resolved.Provider,
		APIKey:      resolved.APIKey,
		Model:       resolved.Model,
		ProviderURL: resolved.ProviderURL,
	})
	if err != nil {
		return fmt.Errorf("create %s client: %w", resolved.Provider, err)
	}

	markdown, err := client.GenerateMarkdown(ctx, BuildPrompt(reviewContext))
	if err != nil {
		return fmt.Errorf("generate review markdown: %w", err)
	}

	if strings.TrimSpace(markdown) == "" {
		return errors.New("generate review markdown: provider returned empty content")
	}

	if err := u.renderer.Render(ctx, out, reviewContext, markdown); err != nil {
		return fmt.Errorf("render html: %w", err)
	}

	return nil
}

type resolvedOptions struct {
	RepositoryPath string
	BaseRef        string
	HeadRef        string
	EventPath      string
	Provider       Provider
	APIKey         string
	Model          string
	ProviderURL    string
	CI             review.CIContext
}

func (u *GenerateReviewUseCase) resolveOptions(opts GenerateReviewOptions) (resolvedOptions, error) {
	provider, err := parseProvider(opts.Provider)
	if err != nil {
		return resolvedOptions{}, err
	}

	keyName := providerAPIKeyName(provider)
	apiKey := envValue(u.env, keyName)
	if !providerAllowsBlankKey(provider, opts.ProviderURL) && strings.TrimSpace(apiKey) == "" {
		return resolvedOptions{}, fmt.Errorf("%s must be set when --provider %s is used", keyName, provider)
	}

	ci := u.githubActionsContext()
	resolved := resolvedOptions{
		RepositoryPath: defaultString(opts.RepositoryPath, "."),
		BaseRef:        resolveBaseRef(opts.BaseRef, ci.BaseRef),
		HeadRef:        defaultString(opts.HeadRef, "HEAD"),
		EventPath:      defaultString(opts.EventPath, envValue(u.env, "GITHUB_EVENT_PATH")),
		Provider:       provider,
		APIKey:         apiKey,
		Model:          resolveModel(opts.Model, provider, u.env),
		ProviderURL:    opts.ProviderURL,
		CI:             ci,
	}

	return resolved, nil
}

func providerAllowsBlankKey(provider Provider, providerURL string) bool {
	return provider == ProviderGemini && providerURL != ""
}

func parseProvider(value string) (Provider, error) {
	switch Provider(strings.TrimSpace(value)) {
	case ProviderGemini:
		return ProviderGemini, nil
	case ProviderGrok:
		return ProviderGrok, nil
	case "":
		return "", errors.New("provider must be one of: gemini, grok")
	default:
		return "", fmt.Errorf("unsupported provider %q; expected one of: gemini, grok", value)
	}
}

func providerAPIKeyName(provider Provider) string {
	switch provider {
	case ProviderGemini:
		return "GEMINI_API_KEY"
	case ProviderGrok:
		return "XAI_API_KEY"
	default:
		return ""
	}
}

func resolveBaseRef(flagValue string, githubBaseRef string) string {
	if flagValue != "" {
		return flagValue
	}

	if githubBaseRef != "" {
		return "origin/" + githubBaseRef
	}

	return "origin/main"
}

func resolveModel(flagValue string, provider Provider, env Environment) string {
	if flagValue != "" {
		return flagValue
	}

	if value := envValue(env, "PRBUDDY_MODEL"); value != "" {
		return value
	}

	switch provider {
	case ProviderGemini:
		return envValue(env, "PRBUDDY_GEMINI_MODEL")
	case ProviderGrok:
		return envValue(env, "PRBUDDY_GROK_MODEL")
	default:
		return ""
	}
}

func (u *GenerateReviewUseCase) githubActionsContext() review.CIContext {
	return review.CIContext{
		GitHubActions: envValue(u.env, "GITHUB_ACTIONS") == "true",
		Repository:    envValue(u.env, "GITHUB_REPOSITORY"),
		SHA:           envValue(u.env, "GITHUB_SHA"),
		EventName:     envValue(u.env, "GITHUB_EVENT_NAME"),
		ServerURL:     envValue(u.env, "GITHUB_SERVER_URL"),
		BaseRef:       envValue(u.env, "GITHUB_BASE_REF"),
		HeadRef:       envValue(u.env, "GITHUB_HEAD_REF"),
	}
}

func defaultString(value string, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}

func envValue(env Environment, name string) string {
	value, ok := env.LookupEnv(name)
	if !ok {
		return ""
	}

	return value
}
