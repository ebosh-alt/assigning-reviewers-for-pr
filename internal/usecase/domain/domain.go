package domain

import (
	"context"
	"time"

	"assigning-reviewers-for-pr/internal/repository"

	"go.uber.org/zap"
)

// Usecase struct implements all usecase interfaces.
type Usecase struct {
	ctx     context.Context
	log     *zap.SugaredLogger
	repo    repository.Repository
	timeout time.Duration
}

// New constructs a new usecase layer with its dependencies.
func New(
	log *zap.SugaredLogger,
	ctx context.Context,
	repo repository.Repository,
	timeout time.Duration,
) *Usecase {
	return &Usecase{
		ctx:     ctx,
		log:     log,
		repo:    repo,
		timeout: timeout,
	}
}
