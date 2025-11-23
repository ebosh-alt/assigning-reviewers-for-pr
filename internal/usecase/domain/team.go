// Package domain contains application Usecases orchestrating domain logic by team.
package domain

import (
	"context"
	"fmt"

	"assigning-reviewers-for-pr/internal/entities"
)

// CreateTeam creates a team with members.
func (u *Usecase) CreateTeam(ctx context.Context, team entities.Team) (*entities.Team, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if team.Name == "" {
		u.log.Errorw("failed to create team: missing team_name")
		return nil, fmt.Errorf("%w: team_name is required", entities.ErrInvalidArgument)
	}
	return u.repo.CreateTeam(ctx, team)
}

// Team returns team by name.
func (u *Usecase) Team(ctx context.Context, name string) (*entities.Team, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if name == "" {
		u.log.Errorw("failed to get team: missing team_name")
		return nil, fmt.Errorf("%w: team_name is required", entities.ErrInvalidArgument)
	}
	return u.repo.GetTeam(ctx, name)
}

// DeactivateTeam deactivates users of a team and cleans reviewer assignments.
func (u *Usecase) DeactivateTeam(ctx context.Context, teamName string) (entities.DeactivateResult, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if teamName == "" {
		u.log.Errorw("failed to deactivate team: missing team_name")
		return entities.DeactivateResult{}, fmt.Errorf("%w: team_name is required", entities.ErrInvalidArgument)
	}
	return u.repo.DeactivateTeam(ctx, teamName)
}
