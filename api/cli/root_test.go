package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/pangobit/prbuddy/app"
)

func TestRenderCommandRunsGenerator(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	generator := &fakeGenerator{}
	cmd := NewRootCommand(&stdout, &stderr, generator)
	cmd.SetArgs([]string{
		"render",
		"--repo", ".",
		"--base", "main",
		"--head", "feature",
		"--event-path", "event.json",
		"--provider", "gemini",
		"--provider-url", "https://ai",
		"--model", "gemini-test",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}

	if got := stdout.String(); got != "<!doctype html><html></html>" {
		t.Fatalf("stdout = %q", got)
	}

	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q", got)
	}

	if generator.got.BaseRef != "main" {
		t.Fatalf("base ref = %q", generator.got.BaseRef)
	}

	if generator.got.Provider != "gemini" {
		t.Fatalf("provider = %q", generator.got.Provider)
	}

	if generator.got.Model != "gemini-test" {
		t.Fatalf("model = %q", generator.got.Model)
	}

	if generator.got.ProviderURL != "https://ai" {
		t.Fatalf("provider url = %q", generator.got.ProviderURL)
	}
}

func TestRenderCommandRejectsEmptyRepo(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := NewRootCommand(&stdout, &stderr, &fakeGenerator{})
	cmd.SetArgs([]string{"render", "--repo", ""})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "repo must not be empty") {
		t.Fatalf("error = %q", err)
	}

	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q", got)
	}
}

type fakeGenerator struct {
	got app.GenerateReviewOptions
}

func (g *fakeGenerator) GenerateReview(_ context.Context, out io.Writer, opts app.GenerateReviewOptions) error {
	g.got = opts
	_, err := io.WriteString(out, "<!doctype html><html></html>")
	return err
}
