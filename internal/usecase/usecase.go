package usecase

import (
	"context"
	"time"

	"assigning-reviewers-for-pr/internal/repository"
	"assigning-reviewers-for-pr/internal/usecase/domain"

	"go.uber.org/zap"
)

// InterfaceUsecase aggregates all usecase interfaces.
type InterfaceUsecase interface {
	UserUsecaseInterface
	TeamUsecaseInterface
	PullRequestUsecaseInterface
	StatsUsecaseInterface
}

// New constructs a new usecase layer with its dependencies.
func New(log *zap.SugaredLogger, ctx context.Context, repo repository.Repository, timeout time.Duration) InterfaceUsecase {
	return domain.New(log, ctx, repo, timeout)
}
