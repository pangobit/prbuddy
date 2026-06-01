package review

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GitCollector collects review context from a local git checkout.
type GitCollector struct {
	repositoryPath string
}

// NewGitCollector creates a collector rooted at repositoryPath.
func NewGitCollector(repositoryPath string) *GitCollector {
	return &GitCollector{repositoryPath: repositoryPath}
}

// Collect builds a review context by comparing the requested git refs.
func (c *GitCollector) Collect(ctx context.Context, opts CollectOptions) (Context, error) {
	repositoryPath := c.repositoryPath
	if opts.RepositoryPath != "" {
		repositoryPath = opts.RepositoryPath
	}

	root, err := gitOutput(ctx, repositoryPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return Context{}, fmt.Errorf("find git root: %w", err)
	}

	mergeBase, err := gitOutput(ctx, root, "merge-base", opts.BaseRef, opts.HeadRef)
	if err != nil {
		return Context{}, fmt.Errorf("find merge base for %q and %q: %w", opts.BaseRef, opts.HeadRef, err)
	}

	remoteURL, err := gitOutput(ctx, root, "remote", "get-url", "origin")
	if err != nil {
		remoteURL = ""
	}

	commits, err := collectCommits(ctx, root, mergeBase, opts.HeadRef)
	if err != nil {
		return Context{}, err
	}

	files, err := collectFiles(ctx, root, opts.BaseRef, opts.HeadRef)
	if err != nil {
		return Context{}, err
	}

	diff, err := gitRawOutput(ctx, root, "diff", "--find-renames", "--patch", opts.BaseRef+"..."+opts.HeadRef)
	if err != nil {
		return Context{}, fmt.Errorf("collect diff: %w", err)
	}

	pullRequest, err := collectEventMetadata(opts.EventPath)
	if err != nil {
		return Context{}, err
	}

	return Context{
		Repository: Repository{
			Name:      filepath.Base(root),
			Root:      root,
			RemoteURL: remoteURL,
		},
		Comparison: Comparison{
			BaseRef:   opts.BaseRef,
			HeadRef:   opts.HeadRef,
			MergeBase: mergeBase,
		},
		Commits:     commits,
		Files:       files,
		Diff:        diff,
		PullRequest: pullRequest,
		CI:          opts.CI,
	}, nil
}

func collectCommits(ctx context.Context, root string, mergeBase string, headRef string) ([]Commit, error) {
	output, err := gitRawOutput(ctx, root, "log", "--format=%H%x00%s%x00%an", mergeBase+".."+headRef)
	if err != nil {
		return nil, fmt.Errorf("collect commits: %w", err)
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	commits := make([]Commit, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, "\x00")
		if len(fields) != 3 {
			return nil, fmt.Errorf("parse commit line %q", line)
		}

		commits = append(commits, Commit{
			Hash:    fields[0],
			Subject: fields[1],
			Author:  fields[2],
		})
	}

	return commits, nil
}

func collectFiles(ctx context.Context, root string, baseRef string, headRef string) ([]ChangedFile, error) {
	spec := baseRef + "..." + headRef

	statusOutput, err := gitRawOutput(ctx, root, "diff", "--find-renames", "--name-status", spec)
	if err != nil {
		return nil, fmt.Errorf("collect file status: %w", err)
	}

	files, err := parseNameStatus(statusOutput)
	if err != nil {
		return nil, err
	}

	numstatOutput, err := gitRawOutput(ctx, root, "diff", "--find-renames", "--numstat", spec)
	if err != nil {
		return nil, fmt.Errorf("collect file stats: %w", err)
	}

	stats, err := parseNumstat(numstatOutput)
	if err != nil {
		return nil, err
	}

	for i := range files {
		stat, ok := stats[files[i].Path]
		if !ok {
			continue
		}

		files[i].Additions = stat.additions
		files[i].Deletions = stat.deletions
	}

	return files, nil
}

func parseNameStatus(output string) ([]ChangedFile, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	files := make([]ChangedFile, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			return nil, fmt.Errorf("parse name status line %q", line)
		}

		status := fields[0]
		path := fields[len(fields)-1]
		files = append(files, ChangedFile{
			Path:   path,
			Status: status,
		})
	}

	return files, nil
}

type fileStat struct {
	additions int
	deletions int
}

func parseNumstat(output string) (map[string]fileStat, error) {
	stats := make(map[string]fileStat)
	output = strings.TrimSpace(output)
	if output == "" {
		return stats, nil
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			return nil, fmt.Errorf("parse numstat line %q", line)
		}

		additions, err := parseGitLineCount(fields[0])
		if err != nil {
			return nil, fmt.Errorf("parse additions for %q: %w", line, err)
		}

		deletions, err := parseGitLineCount(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parse deletions for %q: %w", line, err)
		}

		stats[fields[len(fields)-1]] = fileStat{
			additions: additions,
			deletions: deletions,
		}
	}

	return stats, nil
}

func parseGitLineCount(value string) (int, error) {
	if value == "-" {
		return 0, nil
	}

	lineCount, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return lineCount, nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := gitRawOutput(ctx, dir, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func gitRawOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", err
		}

		return "", fmt.Errorf("%w: %s", err, message)
	}

	return string(output), nil
}
