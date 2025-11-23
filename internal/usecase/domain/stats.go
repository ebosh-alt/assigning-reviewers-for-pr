// Package domain contains application services orchestrating domain logic by statistics.
package domain

import (
	"context"
	"fmt"

	"assigning-reviewers-for-pr/internal/entities"
)

// Stats returns aggregated stats.
func (u *Usecase) Stats(ctx context.Context) (entities.Stats, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()
	return u.repo.Stats(ctx)
}

// SummaryStats returns filtered stats snapshot.
func (u *Usecase) SummaryStats(ctx context.Context, filter entities.StatsFilter) (entities.StatsSummary, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return u.repo.StatsSummary(ctx, filter)
}

// ReviewerStats returns stats for a specific reviewer.
func (u *Usecase) ReviewerStats(ctx context.Context, userID string, limit int) (entities.ReviewerStats, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if userID == "" {
		return entities.ReviewerStats{}, fmt.Errorf("%w: user_id is required", entities.ErrInvalidArgument)
	}
	if limit <= 0 {
		limit = 10
	}
	return u.repo.ReviewerStats(ctx, userID, limit)
}

// PRStats returns stats for a specific pull request.
func (u *Usecase) PRStats(ctx context.Context, prID string) (entities.PRStats, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if prID == "" {
		return entities.PRStats{}, fmt.Errorf("%w: pr_id is required", entities.ErrInvalidArgument)
	}
	return u.repo.PRStats(ctx, prID)
}
