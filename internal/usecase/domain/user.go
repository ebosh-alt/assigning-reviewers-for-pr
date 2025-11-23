// Package domain contains application Usecases orchestrating domain logic by user.
package domain

import (
	"context"
	"fmt"

	"assigning-reviewers-for-pr/internal/entities"
)

// SetActiveUser toggles user activity flag and returns updated user.
func (u *Usecase) SetActiveUser(ctx context.Context, userID string, isActive bool) (*entities.User, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if userID == "" {
		return nil, fmt.Errorf("%w: userID is required", entities.ErrInvalidArgument)
	}

	return u.repo.SetUserActive(ctx, userID, isActive)
}

// GetReviewList returns PRs where the user is assigned as reviewer.
func (u *Usecase) GetReviewList(ctx context.Context, userID string) ([]entities.PullRequestShort, error) {
	ctx, cancel := withTimeout(ctx, u.timeout)
	defer cancel()

	if userID == "" {
		return nil, fmt.Errorf("%w: userID is required", entities.ErrInvalidArgument)
	}

	return u.repo.GetUserReviews(ctx, userID)
}
