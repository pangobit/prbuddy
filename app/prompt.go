package app

import (
	"fmt"
	"strings"

	"github.com/pangobit/prbuddy/internal/llm"
	"github.com/pangobit/prbuddy/internal/review"
)

// BuildPrompt creates the temporary v1 prompt for an LLM review guide.
func BuildPrompt(ctx review.Context) llm.Prompt {
	return llm.Prompt{
		System: placeholderSystemPrompt(),
		User:   placeholderUserPrompt(ctx),
	}
}

func placeholderSystemPrompt() string {
	return strings.TrimSpace(`
You are PR Buddy, an expert software reviewer.
Return Markdown only.
Create a human-ready pull request review document that helps a reviewer understand the change beyond the raw diff.
Use concrete evidence from the supplied commits, changed files, and diff.
Use concise prose, Markdown tables, fenced code snippets, and Mermaid diagrams when they clarify flow or dependencies.
Explain execution flow, side effects, data or state changes, operational risk, and testing implications.
Include structured review comments a human can paste into a PR when specific concerns are visible.
Do not invent facts. If evidence is missing, call that out as an open question.
Do not wrap the response in HTML and do not include raw HTML.
`)
}

func placeholderUserPrompt(ctx review.Context) string {
	var b strings.Builder

	writeLine(&b, "# Pull Request Review Context")
	writeLine(&b, "")
	writeLine(&b, "## Repository")
	writeLine(&b, fmt.Sprintf("- Name: %s", ctx.Repository.Name))
	writeLine(&b, fmt.Sprintf("- Root: %s", ctx.Repository.Root))
	writeLine(&b, fmt.Sprintf("- Remote: %s", emptyFallback(ctx.Repository.RemoteURL, "unknown")))
	writeLine(&b, fmt.Sprintf("- GitHub repository: %s", emptyFallback(ctx.CI.Repository, "unknown")))
	writeLine(&b, "")
	writeLine(&b, "## Comparison")
	writeLine(&b, fmt.Sprintf("- Base ref: %s", ctx.Comparison.BaseRef))
	writeLine(&b, fmt.Sprintf("- Head ref: %s", ctx.Comparison.HeadRef))
	writeLine(&b, fmt.Sprintf("- Merge base: %s", ctx.Comparison.MergeBase))
	writeLine(&b, fmt.Sprintf("- GitHub SHA: %s", emptyFallback(ctx.CI.SHA, "unknown")))

	if ctx.PullRequest != nil {
		writeLine(&b, "")
		writeLine(&b, "## Pull Request Metadata")
		writeLine(&b, fmt.Sprintf("- Number: %d", ctx.PullRequest.Number))
		writeLine(&b, fmt.Sprintf("- Title: %s", emptyFallback(ctx.PullRequest.Title, "unknown")))
		writeLine(&b, fmt.Sprintf("- Author: %s", emptyFallback(ctx.PullRequest.Author, "unknown")))
		writeLine(&b, fmt.Sprintf("- URL: %s", emptyFallback(ctx.PullRequest.URL, "unknown")))
		writeLine(&b, fmt.Sprintf("- Base: %s", emptyFallback(ctx.PullRequest.BaseRef, "unknown")))
		writeLine(&b, fmt.Sprintf("- Head: %s", emptyFallback(ctx.PullRequest.HeadRef, "unknown")))
		writeLine(&b, "")
		writeLine(&b, "### Pull Request Body")
		writeLine(&b, emptyFallback(ctx.PullRequest.Body, "No PR body provided."))
	}

	writeLine(&b, "")
	writeLine(&b, "## Commits")
	if len(ctx.Commits) == 0 {
		writeLine(&b, "- No commits found.")
	} else {
		for _, commit := range ctx.Commits {
			writeLine(&b, fmt.Sprintf("- %s %s by %s", shortHash(commit.Hash), commit.Subject, commit.Author))
		}
	}

	writeLine(&b, "")
	writeLine(&b, "## Changed Files")
	if len(ctx.Files) == 0 {
		writeLine(&b, "- No changed files found.")
	} else {
		for _, file := range ctx.Files {
			writeLine(&b, fmt.Sprintf("- %s %s (+%d/-%d)", file.Status, file.Path, file.Additions, file.Deletions))
		}
	}

	writeLine(&b, "")
	writeLine(&b, "## Diff")
	writeLine(&b, "```diff")
	writeLine(&b, ctx.Diff)
	writeLine(&b, "```")
	writeLine(&b, "")
	writeLine(&b, "## Requested Review Document Shape")
	writeLine(&b, "Return Markdown with these sections when evidence supports them:")
	writeLine(&b, "- Executive Summary")
	writeLine(&b, "- Change Flow")
	writeLine(&b, "- Side Effects and Risks")
	writeLine(&b, "- Key Code Snippets")
	writeLine(&b, "- Review Comments")
	writeLine(&b, "- Verification Guidance")
	writeLine(&b, "- Open Questions")
	writeLine(&b, "")
	writeLine(&b, "For review comments, use this structure:")
	writeLine(&b, "```text")
	writeLine(&b, "File: <path>")
	writeLine(&b, "Concern: <specific concern>")
	writeLine(&b, "Suggested comment: <paste-ready review comment>")
	writeLine(&b, "```")

	return b.String()
}

func writeLine(b *strings.Builder, value string) {
	b.WriteString(value)
	b.WriteByte('\n')
}

func emptyFallback(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func shortHash(hash string) string {
	if len(hash) <= 7 {
		return hash
	}

	return hash[:7]
}
