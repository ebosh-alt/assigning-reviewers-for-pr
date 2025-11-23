package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"assigning-reviewers-for-pr/internal/entities"
)

const (
	setUserActiveQuery = `
WITH updated AS (
    UPDATE users u
    SET is_active = $2
    WHERE u.id = $1
    RETURNING u.id, u.username, u.team_id, u.is_active
)
SELECT up.id, up.username, t.name AS team_name, up.is_active
FROM updated up
JOIN teams t ON t.id = up.team_id
`
	userReviewsQuery = `SELECT pr.id, pr.name, pr.author_id, pr.status
FROM pr_reviewers r
JOIN pull_requests pr ON pr.id = r.pr_id
WHERE r.reviewer_id = $1
ORDER BY pr.created_at DESC`
)

// SetUserActive updates the is_active flag and returns the updated domain user with team name.
func (p *Postgres) SetUserActive(ctx context.Context, userID string, isActive bool) (*entities.User, error) {
	var u entities.User
	err := p.db.QueryRow(ctx, setUserActiveQuery, userID, isActive).
		Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		p.log.Errorw("failed to set user active", "error", err, "user_id", userID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrUserNotFound
		}

		return nil, fmt.Errorf("set user active: %w", err)
	}

	p.log.Infow("user active flag updated", "user_id", userID, "is_active", isActive)
	return &u, nil
}

// GetUserReviews returns PRs where the user is assigned as reviewer.
func (p *Postgres) GetUserReviews(ctx context.Context, userID string) ([]entities.PullRequestShort, error) {
	rows, err := p.db.Query(ctx, userReviewsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("get user reviews: %w", err)
	}
	defer rows.Close()

	prs := make([]entities.PullRequestShort, 0)
	for rows.Next() {
		var pr entities.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			p.log.Errorw("failed to scan user reviews", "error", err, "user_id", userID)
			return nil, fmt.Errorf("scan user reviews: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		p.log.Errorw("failed to iterate user reviews", "error", err, "user_id", userID)
		return nil, fmt.Errorf("iterate user reviews: %w", err)
	}

	return prs, nil
}
