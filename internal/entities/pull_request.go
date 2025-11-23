// Package entities contains core business entities.
package entities

import "time"

// PullRequestStatus enumerates PR lifecycle states.
type PullRequestStatus string

const (
	// StatusOpen marks PR as open.
	StatusOpen PullRequestStatus = "OPEN"
	// StatusMerged marks PR as merged.
	StatusMerged PullRequestStatus = "MERGED"
)

// PullRequest is a domain model of a PR.
type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PullRequestStatus
	Reviewers []string
	CreatedAt *time.Time
	MergedAt  *time.Time
}

// PullRequestShort is a compact projection for reviewer listings.
type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   PullRequestStatus
}
