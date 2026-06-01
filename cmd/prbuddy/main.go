// Package main provides the prbuddy executable.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pangobit/prbuddy/api/cli"
	"github.com/pangobit/prbuddy/app"
	"github.com/pangobit/prbuddy/connectors/gemini"
	"github.com/pangobit/prbuddy/connectors/grok"
	"github.com/pangobit/prbuddy/internal/render"
	"github.com/pangobit/prbuddy/internal/review"
)

func main() {
	env := app.NewEnvironment(os.Environ())
	useCase := app.NewGenerateReviewUseCase(
		review.NewGitCollector("."),
		app.LLMFactoryFunc(newLLMClient),
		render.NewHTMLRenderer(),
		env,
	)
	cmd := cli.NewRootCommand(
		os.Stdout,
		os.Stderr,
		useCase,
	)

	if err := cmd.Execute(); err != nil {
		if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
			os.Exit(1)
		}
		os.Exit(1)
	}
}

func newLLMClient(ctx context.Context, config app.ProviderConfig) (app.LLMClient, error) {
	switch config.Provider {
	case app.ProviderGemini:
		return gemini.NewClient(ctx, config.APIKey, config.Model, config.ProviderURL)
	case app.ProviderGrok:
		return grok.NewClient(config.APIKey, config.Model), nil
	default:
		return nil, errors.New("unknown provider")
	}
}
