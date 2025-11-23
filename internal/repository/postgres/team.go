package postgres

import (
	"context"
	"errors"
	"fmt"

	"assigning-reviewers-for-pr/internal/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

const (
	insertTeamQuery = "INSERT INTO teams(name) VALUES($1) RETURNING id"
	upsertUserQuery = `
INSERT INTO users(id, username, team_id, is_active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username, team_id = EXCLUDED.team_id, is_active = EXCLUDED.is_active
`
	selectTeamIDQuery         = "SELECT id FROM teams WHERE name=$1"
	selectTeamMembersQuery    = "SELECT id, username, is_active FROM users WHERE team_id=$1"
	selectTeamIDForDeactivate = `SELECT id FROM teams WHERE name=$1`
	deactivateUsersQuery      = `UPDATE users SET is_active=false WHERE team_id=$1 AND is_active=true RETURNING id`
	selectImpactedPRsQuery    = `
SELECT pr.id, pr.author_id
FROM pull_requests pr
WHERE pr.status='OPEN' AND EXISTS (
    SELECT 1 FROM pr_reviewers r WHERE r.pr_id = pr.id AND r.reviewer_id = ANY($1::text[])
) FOR UPDATE`
	selectPRStatusQuery         = `SELECT status FROM pull_requests WHERE id=$1`
	deleteReviewerForDeactivate = `DELETE FROM pr_reviewers WHERE pr_id=$1 AND reviewer_id=$2`
	insertReviewerForDeactivate = `INSERT INTO pr_reviewers(pr_id, reviewer_id) VALUES ($1,$2)`
	activeReplacementQuery      = `SELECT id FROM users WHERE is_active=true AND team_id <> $1 AND id <> $2`
)

// CreateTeam inserts a team and upserts its members.
func (p *Postgres) CreateTeam(ctx context.Context, team entities.Team) (*entities.Team, error) {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var teamID int64
	if err := tx.QueryRow(ctx, insertTeamQuery, team.Name).Scan(&teamID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, entities.ErrTeamExists
		}
		return nil, fmt.Errorf("insert team: %w", err)
	}

	for _, m := range team.Members {
		if _, err := tx.Exec(ctx, upsertUserQuery, m.ID, m.Username, teamID, m.IsActive); err != nil {
			return nil, fmt.Errorf("upsert user: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	p.log.Infow("team created", "team", team.Name, "members", len(team.Members))
	return p.GetTeam(ctx, team.Name)
}

// GetTeam fetches team with members by name.
func (p *Postgres) GetTeam(ctx context.Context, name string) (*entities.Team, error) {
	var teamID int64
	if err := p.db.QueryRow(ctx, selectTeamIDQuery, name).Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrTeamNotFound
		}
		return nil, fmt.Errorf("get team: %w", err)
	}

	rows, err := p.db.Query(ctx, selectTeamMembersQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}
	defer rows.Close()

	members := make([]entities.User, 0)
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan members: %w", err)
		}
		u.TeamName = name
		members = append(members, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return &entities.Team{Name: name, Members: members}, nil
}

// DeactivateTeam bulk deactivates team users and reassigns their open PRs to active users from other teams.
func (p *Postgres) DeactivateTeam(ctx context.Context, teamName string) (entities.DeactivateResult, error) {
	res := entities.DeactivateResult{}

	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return res, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var teamID int64
	if err := tx.QueryRow(ctx, selectTeamIDForDeactivate, teamName).Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, entities.ErrTeamNotFound
		}
		return res, fmt.Errorf("team lookup: %w", err)
	}

	rows, err := tx.Query(ctx, deactivateUsersQuery, teamID)
	if err != nil {
		return res, fmt.Errorf("deactivate users: %w", err)
	}
	defer rows.Close()
	deactivated := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return res, err
		}
		deactivated = append(deactivated, id)
	}
	if err := rows.Err(); err != nil {
		return res, err
	}
	res.DeactivatedUsers = len(deactivated)

	if len(deactivated) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return res, err
		}
		return res, nil
	}

	prRows, err := tx.Query(ctx, selectImpactedPRsQuery, deactivated)
	if err != nil {
		return res, fmt.Errorf("select affected prs: %w", err)
	}
	defer prRows.Close()

	type impactedPR struct {
		id       string
		authorID string
	}
	impacted := make([]impactedPR, 0)
	for prRows.Next() {
		var prID, authorID string
		if err := prRows.Scan(&prID, &authorID); err != nil {
			return res, err
		}
		impacted = append(impacted, impactedPR{id: prID, authorID: authorID})
	}
	if err := prRows.Err(); err != nil {
		return res, err
	}

	for _, pr := range impacted {
		var status string
		if err := tx.QueryRow(ctx, selectPRStatusQuery, pr.id).Scan(&status); err != nil {
			return res, fmt.Errorf("status check: %w", err)
		}
		if status != string(entities.StatusOpen) {
			continue
		}

		reviewers, err := p.readReviewers(ctx, tx, pr.id)
		if err != nil {
			return res, err
		}

		existing := make(map[string]struct{}, len(reviewers))
		for _, r := range reviewers {
			existing[r] = struct{}{}
		}

		for _, r := range reviewers {
			if !contains(deactivated, r) {
				continue
			}

			if _, err := tx.Exec(ctx, deleteReviewerForDeactivate, pr.id, r); err != nil {
				return res, fmt.Errorf("delete old reviewer: %w", err)
			}
			delete(existing, r)

			candidate, ok, err := p.pickReplacement(ctx, tx, teamID, pr.authorID, existing)
			if err != nil {
				return res, err
			}
			if !ok {
				if err := p.insertReassignmentHistory(ctx, tx, pr.id, r, nil); err != nil {
					return res, err
				}
				res.Removed++
				continue
			}
			if _, err := tx.Exec(ctx, insertReviewerForDeactivate, pr.id, candidate); err != nil {
				return res, fmt.Errorf("insert replacement: %w", err)
			}
			if err := p.insertReassignmentHistory(ctx, tx, pr.id, r, &candidate); err != nil {
				return res, err
			}
			existing[candidate] = struct{}{}
			res.Reassigned++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return res, err
	}

	p.log.Infow("team deactivated", "team", teamName, "deactivated_users", res.DeactivatedUsers, "reassigned", res.Reassigned, "removed", res.Removed)
	return res, nil
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func (p *Postgres) pickReplacement(ctx context.Context, tx pgx.Tx, deactivatedTeamID int64, authorID string, existing map[string]struct{}) (string, bool, error) {
	rows, err := tx.Query(ctx, activeReplacementQuery, deactivatedTeamID, authorID)
	if err != nil {
		return "", false, fmt.Errorf("select candidates: %w", err)
	}
	defer rows.Close()
	pool := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", false, err
		}
		if _, ok := existing[id]; ok {
			continue
		}
		pool = append(pool, id)
	}
	if err := rows.Err(); err != nil {
		return "", false, err
	}
	if len(pool) == 0 {
		return "", false, nil
	}
	candidate := pickRandom(pool, 1)[0]
	return candidate, true, nil
}
