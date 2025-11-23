// Package domain contains application services orchestrating domain logic by pull request.
package domain

import (
	"context"
	"fmt"

	"assigning-reviewers-for-pr/internal/entities"
)

// CreatePullRequest creates PR and auto-assigns reviewers.
func (u *Usecase) CreatePullRequest(ctx context.Context, pr entities.PullRequest) (*entities.PullRequest, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if pr.ID == "" || pr.Name == "" || pr.AuthorID == "" {
		return nil, fmt.Errorf("%w: missing required fields", entities.ErrInvalidArgument)
	}
	res, err := u.repo.CreatePR(ctx, pr)
	if err != nil {
		return nil, err
	}
	u.log.Infow("pr create", "pr_id", pr.ID)
	return res, nil
}

// MergePullRequest marks PR as merged idempotently.
func (u *Usecase) MergePullRequest(ctx context.Context, prID string) (*entities.PullRequest, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if prID == "" {
		return nil, fmt.Errorf("%w: pull_request_id is required", entities.ErrInvalidArgument)
	}
	return u.repo.MergePR(ctx, prID)
}

// ReassignPullRequest swaps reviewer.
func (u *Usecase) ReassignPullRequest(ctx context.Context, prID, oldUserID string) (*entities.PullRequest, string, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if prID == "" || oldUserID == "" {
		return nil, "", fmt.Errorf("%w: missing required fields", entities.ErrInvalidArgument)
	}
	return u.repo.ReassignReviewer(ctx, prID, oldUserID)
}
