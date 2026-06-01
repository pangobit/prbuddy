package review

import "testing"

func TestParseGitHubEventReturnsPullRequestMetadata(t *testing.T) {
	t.Parallel()

	metadata, err := parseGitHubEvent([]byte(`{
		"number": 42,
		"pull_request": {
			"title": "Add LLM review",
			"body": "Please review this carefully.",
			"html_url": "https://github.com/pangobit/prbuddy/pull/42",
			"user": {"login": "ray"},
			"base": {"ref": "main", "sha": "base-sha"},
			"head": {"ref": "feature", "sha": "head-sha"}
		}
	}`))
	if err != nil {
		t.Fatalf("parse event: %v", err)
	}

	if metadata == nil {
		t.Fatal("metadata is nil")
	}

	if metadata.Number != 42 {
		t.Fatalf("number = %d", metadata.Number)
	}

	if metadata.Title != "Add LLM review" {
		t.Fatalf("title = %q", metadata.Title)
	}

	if metadata.Author != "ray" {
		t.Fatalf("author = %q", metadata.Author)
	}

	if metadata.BaseSHA != "base-sha" {
		t.Fatalf("base sha = %q", metadata.BaseSHA)
	}
}

func TestParseGitHubEventToleratesNonPullRequestPayload(t *testing.T) {
	t.Parallel()

	metadata, err := parseGitHubEvent([]byte(`{"repository":{"full_name":"pangobit/prbuddy"}}`))
	if err != nil {
		t.Fatalf("parse event: %v", err)
	}

	if metadata != nil {
		t.Fatalf("metadata = %#v", metadata)
	}
}
