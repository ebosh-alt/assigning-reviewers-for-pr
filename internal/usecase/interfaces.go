package usecase

import (
	"context"

	"assigning-reviewers-for-pr/internal/entities"
)

// UserUsecaseInterface abstracts user-related operations for delivery layer.
type UserUsecaseInterface interface {
	SetActiveUser(ctx context.Context, userID string, isActive bool) (*entities.User, error)
	GetReviewList(ctx context.Context, userID string) ([]entities.PullRequestShort, error)
}

// TeamUsecaseInterface abstracts team-related operations.
type TeamUsecaseInterface interface {
	CreateTeam(ctx context.Context, team entities.Team) (*entities.Team, error)
	Team(ctx context.Context, name string) (*entities.Team, error)
	DeactivateTeam(ctx context.Context, teamName string) (entities.DeactivateResult, error)
}

// PullRequestUsecaseInterface abstracts PR-related operations.
type PullRequestUsecaseInterface interface {
	CreatePullRequest(ctx context.Context, pr entities.PullRequest) (*entities.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*entities.PullRequest, error)
	ReassignPullRequest(ctx context.Context, prID, oldUserID string) (*entities.PullRequest, string, error)
}

// StatsUsecaseInterface abstracts statistics operations.
type StatsUsecaseInterface interface {
	Stats(ctx context.Context) (entities.Stats, error)
	SummaryStats(ctx context.Context, filter entities.StatsFilter) (entities.StatsSummary, error)
	ReviewerStats(ctx context.Context, userID string, limit int) (entities.ReviewerStats, error)
	PRStats(ctx context.Context, prID string) (entities.PRStats, error)
}
