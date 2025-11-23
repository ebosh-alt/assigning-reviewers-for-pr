// Package repository provides factory for repositories.
package repository

import (
	"context"
	"fmt"

	"assigning-reviewers-for-pr/config"
	"assigning-reviewers-for-pr/internal/repository/postgres"

	"go.uber.org/zap"
)

// Repository aggregates all persistence interfaces.
type Repository interface {
	LifecycleInterface
	UserInterface
	TeamInterface
	PullRequestInterface
	StatsInterface
}

// New constructs repository backend by name.
func New(ctx context.Context, name string, log *zap.SugaredLogger, cfg *config.Config) (Repository, error) {
	switch name {
	case "postgres":
		return postgres.New(ctx, log, cfg), nil
	default:
		return nil, fmt.Errorf("unknown repo backend: %s", name)
	}
}
