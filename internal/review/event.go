package review

import (
	"encoding/json"
	"fmt"
	"os"
)

func collectEventMetadata(path string) (*PullRequestMetadata, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read event payload: %w", err)
	}

	metadata, err := parseGitHubEvent(data)
	if err != nil {
		return nil, fmt.Errorf("parse event payload: %w", err)
	}

	return metadata, nil
}

func parseGitHubEvent(data []byte) (*PullRequestMetadata, error) {
	var event githubEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	if event.PullRequest == nil {
		return nil, nil
	}

	return &PullRequestMetadata{
		Number:  firstNonZero(event.PullRequest.Number, event.Number),
		Title:   event.PullRequest.Title,
		Body:    event.PullRequest.Body,
		URL:     event.PullRequest.HTMLURL,
		Author:  event.PullRequest.User.Login,
		BaseRef: event.PullRequest.Base.Ref,
		HeadRef: event.PullRequest.Head.Ref,
		BaseSHA: event.PullRequest.Base.SHA,
		HeadSHA: event.PullRequest.Head.SHA,
	}, nil
}

type githubEvent struct {
	Number      int                `json:"number"`
	PullRequest *githubPullRequest `json:"pull_request"`
}

type githubPullRequest struct {
	Number  int           `json:"number"`
	Title   string        `json:"title"`
	Body    string        `json:"body"`
	HTMLURL string        `json:"html_url"`
	User    githubUser    `json:"user"`
	Base    githubPullRef `json:"base"`
	Head    githubPullRef `json:"head"`
}

type githubUser struct {
	Login string `json:"login"`
}

type githubPullRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}
