// Package cli wires PR Buddy's command-line interface to application behavior.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/pangobit/prbuddy/app"
)

// Generator runs the PR review generation use case.
type Generator interface {
	GenerateReview(context.Context, io.Writer, app.GenerateReviewOptions) error
}

// NewRootCommand creates the PR Buddy root command.
func NewRootCommand(out io.Writer, errOut io.Writer, generator Generator) *cobra.Command {
	root := &cobra.Command{
		Use:           "prbuddy",
		Short:         "Generate PR review artifacts for humans and LLM workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetOut(out)
	root.SetErr(errOut)
	root.AddCommand(newRenderCommand(out, generator))

	return root
}

func newRenderCommand(out io.Writer, generator Generator) *cobra.Command {
	var opts app.GenerateReviewOptions

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Generate and render a PR review guide as HTML",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateRenderOptions(opts); err != nil {
				return err
			}

			if err := generator.GenerateReview(cmd.Context(), out, opts); err != nil {
				return fmt.Errorf("generate review: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.RepositoryPath, "repo", ".", "path to the git repository")
	cmd.Flags().StringVar(&opts.BaseRef, "base", "", "base git ref for the review")
	cmd.Flags().StringVar(&opts.HeadRef, "head", "", "head git ref for the review")
	cmd.Flags().StringVar(&opts.EventPath, "event-path", "", "path to a GitHub Actions event payload")
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "LLM provider to use: gemini or grok")
	cmd.Flags().StringVar(&opts.Model, "model", "", "provider model override")
	cmd.Flags().StringVar(&opts.ProviderURL, "provider-url", "", "provider endpoint override")

	return cmd
}

func validateRenderOptions(opts app.GenerateReviewOptions) error {
	if opts.RepositoryPath == "" {
		return errors.New("repo must not be empty")
	}

	return nil
}
