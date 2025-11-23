// Package repository contains repository interfaces for persistence layers.
package repository

import (
	"context"

	"assigning-reviewers-for-pr/internal/entities"
)

// LifecycleInterface describes storage startup/shutdown hooks.
type LifecycleInterface interface {
	OnStart(_ context.Context) error
	OnStop(_ context.Context) error
}

// UserInterface exposes user-related operations.
type UserInterface interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*entities.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]entities.PullRequestShort, error)
}

// TeamInterface exposes team-related operations.
type TeamInterface interface {
	CreateTeam(ctx context.Context, team entities.Team) (*entities.Team, error)
	GetTeam(ctx context.Context, name string) (*entities.Team, error)
}

// PullRequestInterface exposes PR-related operations.
type PullRequestInterface interface {
	CreatePR(ctx context.Context, pr entities.PullRequest) (*entities.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*entities.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*entities.PullRequest, string, error)
}

// StatsInterface exposes aggregated statistics operations.
type StatsInterface interface {
	Stats(ctx context.Context) (entities.Stats, error)
	StatsSummary(ctx context.Context, filter entities.StatsFilter) (entities.StatsSummary, error)
	ReviewerStats(ctx context.Context, userID string, limit int) (entities.ReviewerStats, error)
	PRStats(ctx context.Context, prID string) (entities.PRStats, error)
	DeactivateTeam(ctx context.Context, teamName string) (entities.DeactivateResult, error)
}
