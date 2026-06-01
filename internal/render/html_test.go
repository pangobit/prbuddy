package render

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/pangobit/prbuddy/internal/review"
)

func TestHTMLRendererWritesCompleteDocumentFromMarkdown(t *testing.T) {
	t.Parallel()

	reviewContext := review.Context{
		Repository: review.Repository{
			Name:      "prbuddy",
			Root:      "/tmp/prbuddy",
			RemoteURL: "git@github.com:pangobit/prbuddy.git",
		},
		Comparison: review.Comparison{
			BaseRef:   "main",
			HeadRef:   "feature",
			MergeBase: "1234567890abcdef",
		},
		PullRequest: &review.PullRequestMetadata{
			Number: 42,
			Title:  "Add renderer",
			URL:    "https://github.com/pangobit/prbuddy/pull/42",
			Author: "ray",
		},
		Commits: []review.Commit{{Hash: "abcdef1234567890"}},
		Files:   []review.ChangedFile{{Path: "internal/render/html.go"}},
		Diff:    "diff --git a/a b/a\n+hello",
	}

	var out bytes.Buffer
	renderer := NewHTMLRenderer()
	if err := renderer.Render(context.Background(), &out, reviewContext, "## Review\n\n- Check tests."); err != nil {
		t.Fatalf("render html: %v", err)
	}

	got := out.String()
	required := []string{
		"<!doctype html>",
		"<html lang=\"en\">",
		"PR Buddy Review Guide",
		"#42 Add renderer",
		"<h2>Review</h2>",
		"<li>Check tests.</li>",
		"1 commits, 1 changed files.",
		"</html>",
	}

	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered html missing %q in:\n%s", want, got)
		}
	}
}

func TestHTMLRendererEscapesRawHTMLInMarkdown(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	renderer := NewHTMLRenderer()
	err := renderer.Render(context.Background(), &out, review.Context{}, "<script>alert('x')</script>")
	if err != nil {
		t.Fatalf("render html: %v", err)
	}

	if strings.Contains(out.String(), "<script>") {
		t.Fatalf("rendered html contains raw script:\n%s", out.String())
	}
}

func TestHTMLRendererSupportsTablesAndMermaidBlocks(t *testing.T) {
	t.Parallel()

	markdown := strings.Join([]string{
		"| Risk | Impact |",
		"| --- | --- |",
		"| Side effect | High |",
		"",
		"```mermaid",
		"flowchart TD",
		"  A --> B",
		"```",
	}, "\n")

	var out bytes.Buffer
	renderer := NewHTMLRenderer()
	err := renderer.Render(context.Background(), &out, review.Context{}, markdown)
	if err != nil {
		t.Fatalf("render html: %v", err)
	}

	got := out.String()
	required := []string{
		"<table>",
		"<th>Risk</th>",
		"<td>Side effect</td>",
		"language-mermaid",
		"mermaid.run",
		"securityLevel: \"strict\"",
	}

	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered html missing %q in:\n%s", want, got)
		}
	}
}
