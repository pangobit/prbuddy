package review

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitCollectorCollectsReviewContext(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	runGitTest(t, dir, "init")
	runGitTest(t, dir, "config", "user.name", "PR Buddy")
	runGitTest(t, dir, "config", "user.email", "prbuddy@example.com")
	runGitTest(t, dir, "checkout", "-b", "main")
	writeTestFile(t, filepath.Join(dir, "README.md"), "hello\n")
	runGitTest(t, dir, "add", "README.md")
	runGitTest(t, dir, "commit", "-m", "Initial commit")
	runGitTest(t, dir, "checkout", "-b", "feature")
	writeTestFile(t, filepath.Join(dir, "README.md"), "hello\nworld\n")
	runGitTest(t, dir, "add", "README.md")
	runGitTest(t, dir, "commit", "-m", "Expand readme")

	collector := NewGitCollector(dir)
	ctx, err := collector.Collect(context.Background(), CollectOptions{
		RepositoryPath: dir,
		BaseRef:        "main",
		HeadRef:        "feature",
	})
	if err != nil {
		t.Fatalf("collect review context: %v", err)
	}

	if ctx.Repository.Name != filepath.Base(dir) {
		t.Fatalf("repository name = %q", ctx.Repository.Name)
	}

	if ctx.Comparison.BaseRef != "main" {
		t.Fatalf("base ref = %q", ctx.Comparison.BaseRef)
	}

	if len(ctx.Commits) != 1 {
		t.Fatalf("commit count = %d", len(ctx.Commits))
	}

	if ctx.Commits[0].Subject != "Expand readme" {
		t.Fatalf("commit subject = %q", ctx.Commits[0].Subject)
	}

	if len(ctx.Files) != 1 {
		t.Fatalf("file count = %d", len(ctx.Files))
	}

	if ctx.Files[0].Path != "README.md" {
		t.Fatalf("file path = %q", ctx.Files[0].Path)
	}

	if ctx.Files[0].Additions != 1 {
		t.Fatalf("additions = %d", ctx.Files[0].Additions)
	}

	if !strings.Contains(ctx.Diff, "+world") {
		t.Fatalf("diff = %q", ctx.Diff)
	}
}

func TestParseNameStatusRejectsMalformedLine(t *testing.T) {
	t.Parallel()

	_, err := parseNameStatus("M")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseNumstatHandlesBinaryFiles(t *testing.T) {
	t.Parallel()

	stats, err := parseNumstat("-\t-\timage.png")
	if err != nil {
		t.Fatalf("parse numstat: %v", err)
	}

	if stats["image.png"].additions != 0 {
		t.Fatalf("additions = %d", stats["image.png"].additions)
	}

	if stats["image.png"].deletions != 0 {
		t.Fatalf("deletions = %d", stats["image.png"].deletions)
	}
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
