// Package review models the repository context needed to guide PR review.
package review

// CollectOptions describes the local comparison a collector should inspect.
type CollectOptions struct {
	// RepositoryPath is the filesystem path to the git repository.
	RepositoryPath string
	// BaseRef is the git ref that represents the target branch.
	BaseRef string
	// HeadRef is the git ref that represents the proposed changes.
	HeadRef string
	// EventPath is the optional path to a GitHub Actions event payload.
	EventPath string
	// CI contains environment metadata from the current workflow run.
	CI CIContext
}

// Context contains the structured material used to render a review guide.
type Context struct {
	// Repository describes the checked-out repository.
	Repository Repository
	// Comparison identifies the base and head refs being reviewed.
	Comparison Comparison
	// Commits lists commits included in the comparison.
	Commits []Commit
	// Files lists files changed in the comparison.
	Files []ChangedFile
	// Diff contains the patch for the comparison.
	Diff string
	// PullRequest contains PR metadata when it is available from an event payload.
	PullRequest *PullRequestMetadata
	// CI contains CI metadata when PR Buddy runs inside GitHub Actions.
	CI CIContext
}

// Repository describes the git repository under review.
type Repository struct {
	// Name is the repository name inferred from the git root.
	Name string
	// Root is the absolute git working tree root.
	Root string
	// RemoteURL is the origin remote URL when one is configured.
	RemoteURL string
}

// Comparison identifies the refs and merge base used for review.
type Comparison struct {
	// BaseRef is the requested base git ref.
	BaseRef string
	// HeadRef is the requested head git ref.
	HeadRef string
	// MergeBase is the git merge base between BaseRef and HeadRef.
	MergeBase string
}

// Commit describes one commit included in the comparison.
type Commit struct {
	// Hash is the full git commit hash.
	Hash string
	// Subject is the first line of the commit message.
	Subject string
	// Author is the commit author name.
	Author string
}

// ChangedFile describes a file changed in the comparison.
type ChangedFile struct {
	// Path is the changed file path.
	Path string
	// Status is the git change status.
	Status string
	// Additions is the number of added lines when git can report it.
	Additions int
	// Deletions is the number of deleted lines when git can report it.
	Deletions int
}

// PullRequestMetadata describes a pull request from a local event payload.
type PullRequestMetadata struct {
	// Number is the pull request number.
	Number int
	// Title is the pull request title.
	Title string
	// Body is the pull request description.
	Body string
	// URL is the browser URL for the pull request.
	URL string
	// Author is the pull request author login.
	Author string
	// BaseRef is the target branch name from the event payload.
	BaseRef string
	// HeadRef is the source branch name from the event payload.
	HeadRef string
	// BaseSHA is the target branch SHA from the event payload.
	BaseSHA string
	// HeadSHA is the source branch SHA from the event payload.
	HeadSHA string
}

// CIContext describes metadata available from a CI provider.
type CIContext struct {
	// GitHubActions is true when running in GitHub Actions.
	GitHubActions bool
	// Repository is the repository identifier supplied by CI.
	Repository string
	// SHA is the triggering commit SHA supplied by CI.
	SHA string
	// EventName is the workflow event name supplied by CI.
	EventName string
	// ServerURL is the GitHub server URL supplied by CI.
	ServerURL string
	// BaseRef is the pull request base branch supplied by CI.
	BaseRef string
	// HeadRef is the pull request head branch supplied by CI.
	HeadRef string
}
