package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"assigning-reviewers-for-pr/internal/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

const (
	selectAuthorQuery                = `SELECT u.team_id, u.is_active FROM users u WHERE u.id=$1`
	insertPRQuery                    = `INSERT INTO pull_requests(id, name, author_id, status) VALUES ($1,$2,$3,'OPEN')`
	selectCandidatesQuery            = `SELECT id FROM users WHERE team_id=$1 AND is_active=true AND id <> $2`
	selectPRForUpdateQuery           = `SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id=$1 FOR UPDATE`
	updatePRMergedQuery              = `UPDATE pull_requests SET status='MERGED', merged_at=NOW() WHERE id=$1 RETURNING merged_at`
	selectReviewersQuery             = `SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1`
	deleteReviewerQuery              = `DELETE FROM pr_reviewers WHERE pr_id=$1 AND reviewer_id=$2`
	insertReviewerQuery              = `INSERT INTO pr_reviewers(pr_id, reviewer_id) VALUES ($1,$2)`
	selectReviewerTeamQuery          = `SELECT team_id FROM users WHERE id=$1`
	selectReplacementCandidatesQuery = `SELECT id FROM users WHERE team_id=$1 AND is_active=true AND id <> $2`
)

// CreatePR creates PR and assigns up to two reviewers.
func (p *Postgres) CreatePR(ctx context.Context, pr entities.PullRequest) (res *entities.PullRequest, err error) {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var authorTeamID int64
	var authorActive bool
	if err := tx.QueryRow(ctx, selectAuthorQuery, pr.AuthorID).Scan(&authorTeamID, &authorActive); err != nil {
		p.log.Errorw("failed to query author team", "error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("author lookup: %w", err)
	}

	if !authorActive {
		return nil, fmt.Errorf("%w: author inactive", entities.ErrInvalidArgument)
	}

	if _, err := tx.Exec(ctx, insertPRQuery, pr.ID, pr.Name, pr.AuthorID); err != nil {
		var pgErr *pgconn.PgError
		p.log.Errorw("failed to insert pull request", "error", err, "id", pr.ID)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, entities.ErrPRExists
		}
		return nil, fmt.Errorf("insert pr: %w", err)
	}

	candidatesRows, err := tx.Query(ctx, selectCandidatesQuery, authorTeamID, pr.AuthorID)
	if err != nil {
		p.log.Errorw("failed to select candidates", "error", err)
		return nil, fmt.Errorf("select candidates: %w", err)
	}
	defer candidatesRows.Close()
	candidates := make([]string, 0)
	for candidatesRows.Next() {
		var id string
		if err := candidatesRows.Scan(&id); err != nil {
			p.log.Errorw("failed to scan candidate", "error", err)
			return nil, err
		}
		candidates = append(candidates, id)
	}
	if err := candidatesRows.Err(); err != nil {
		p.log.Errorw("error iterating candidates", "error", err)
		return nil, err
	}

	reviewers := pickRandom(candidates, 2)
	for _, r := range reviewers {
		if _, err := tx.Exec(ctx, insertReviewerQuery, pr.ID, r); err != nil {
			p.log.Errorw("failed to insert reviewer", "error", err, "reviewer_id", r)
			return nil, fmt.Errorf("insert reviewer: %w", err)
		}
	}

	var createdAt time.Time
	if err := tx.QueryRow(ctx, `SELECT created_at FROM pull_requests WHERE id=$1`, pr.ID).Scan(&createdAt); err != nil {
		p.log.Errorw("failed to select created_at", "error", err, "pr_id", pr.ID)
		return nil, fmt.Errorf("select created_at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	pr.Reviewers = reviewers
	pr.Status = entities.StatusOpen
	pr.CreatedAt = &createdAt
	p.log.Infow("pr created", "pr_id", pr.ID, "reviewers", reviewers)
	return &pr, nil
}

// MergePR marks PR merged idempotently.
func (p *Postgres) MergePR(ctx context.Context, prID string) (res *entities.PullRequest, err error) {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var pr entities.PullRequest
	var createdAt time.Time
	var mergedAt *time.Time
	if err := tx.QueryRow(ctx, selectPRForUpdateQuery, prID).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt); err != nil {
		p.log.Errorw("failed to select pr for update", "error", err, "pr_id", prID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrPRNotFound
		}
		return nil, fmt.Errorf("get pr: %w", err)
	}

	pr.CreatedAt = &createdAt
	pr.MergedAt = mergedAt

	if pr.Status != entities.StatusMerged {
		var now time.Time
		if err := tx.QueryRow(ctx, updatePRMergedQuery, prID).Scan(&now); err != nil {
			p.log.Errorw("failed to update pr merged", "error", err, "pr_id", prID)
			return nil, fmt.Errorf("merge pr: %w", err)
		}
		pr.Status = entities.StatusMerged
		pr.MergedAt = &now
	}

	reviewers, err := p.readReviewers(ctx, tx, prID)
	if err != nil {

		return nil, err
	}
	pr.Reviewers = reviewers

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	p.log.Infow("pr merged", "pr_id", prID)
	return &pr, nil
}

// ReassignReviewer replaces reviewer with another active member of same team.
func (p *Postgres) ReassignReviewer(ctx context.Context, prID, oldUserID string) (res *entities.PullRequest, repl string, err error) {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var pr entities.PullRequest
	var createdAt time.Time
	if err := tx.QueryRow(ctx, selectPRForUpdateQuery, prID).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &createdAt, &pr.MergedAt); err != nil {
		p.log.Errorw("failed to select pr for update", "error", err, "pr_id", prID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", entities.ErrPRNotFound
		}
		return nil, "", fmt.Errorf("get pr: %w", err)
	}
	pr.CreatedAt = &createdAt

	if pr.Status == entities.StatusMerged {
		return nil, "", entities.ErrPRMerged
	}

	reviewers, err := p.readReviewers(ctx, tx, prID)
	if err != nil {
		return nil, "", err
	}
	pr.Reviewers = reviewers

	assigned := false
	for _, r := range reviewers {
		if r == oldUserID {
			assigned = true
			break
		}
	}
	if !assigned {
		p.log.Errorw("old reviewer not assigned to PR", "pr_id", prID, "old_reviewer", oldUserID)
		return nil, "", entities.ErrNotAssigned
	}

	var teamID int64
	if err := tx.QueryRow(ctx, selectReviewerTeamQuery, oldUserID).Scan(&teamID); err != nil {
		p.log.Errorw("failed to select old reviewer team", "error", err, "old_reviewer", oldUserID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", entities.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("old reviewer lookup: %w", err)
	}

	rows, err := tx.Query(ctx, selectReplacementCandidatesQuery, teamID, pr.AuthorID)
	if err != nil {
		p.log.Errorw("failed to select replacements", "error", err, "pr_id", prID)
		return nil, "", fmt.Errorf("select replacements: %w", err)
	}
	defer rows.Close()

	candidates := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, "", err
		}
		already := false
		for _, r := range reviewers {
			if r == id {
				already = true
				break
			}
		}
		if !already {
			candidates = append(candidates, id)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	if len(candidates) == 0 {
		return nil, "", entities.ErrNoCandidate
	}

	repl = pickRandom(candidates, 1)[0]

	if _, err := tx.Exec(ctx, deleteReviewerQuery, prID, oldUserID); err != nil {
		return nil, "", fmt.Errorf("delete old reviewer: %w", err)
	}
	if _, err := tx.Exec(ctx, insertReviewerQuery, prID, repl); err != nil {
		return nil, "", fmt.Errorf("insert replacement: %w", err)
	}
	if err := p.insertReassignmentHistory(ctx, tx, prID, oldUserID, &repl); err != nil {
		return nil, "", err
	}

	reviewers = append(filterOut(reviewers, oldUserID), repl)
	pr.Reviewers = reviewers

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	p.log.Infow("reviewer reassigned", "pr_id", prID, "old", oldUserID, "new", repl)
	return &pr, repl, nil
}

func (p *Postgres) readReviewers(ctx context.Context, tx pgx.Tx, prID string) ([]string, error) {
	rows, err := tx.Query(ctx, selectReviewersQuery, prID)
	if err != nil {
		p.log.Errorw("failed to select reviewers", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("select reviewers: %w", err)
	}
	defer rows.Close()
	revs := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			p.log.Errorw("failed to scan reviewer", "error", err)
			return nil, err
		}
		revs = append(revs, id)
	}
	if err := rows.Err(); err != nil {
		p.log.Errorw("error iterating reviewers", "error", err)
		return nil, err
	}
	return revs, nil
}

func (p *Postgres) insertReassignmentHistory(ctx context.Context, tx pgx.Tx, prID, oldReviewer string, newReviewer *string) error {
	if _, err := tx.Exec(ctx, `INSERT INTO pr_reassignment_history(pr_id, old_reviewer_id, new_reviewer_id) VALUES ($1,$2,$3)`, prID, oldReviewer, newReviewer); err != nil {
		return fmt.Errorf("insert reassignment history: %w", err)
	}
	return nil
}

func filterOut(list []string, target string) []string {
	res := make([]string, 0, len(list))
	for _, v := range list {
		if v != target {
			res = append(res, v)
		}
	}
	return res
}

func pickRandom(src []string, n int) []string {
	if n >= len(src) {
		return append([]string(nil), src...)
	}
	pool := append([]string(nil), src...)
	res := make([]string, 0, n)
	for i := 0; i < n; i++ {
		limit := big.NewInt(int64(len(pool)))
		idxBig, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return pool[:n] // fallback deterministic slice
		}
		idx := idxBig.Int64()
		res = append(res, pool[idx])
		pool = append(pool[:idx], pool[idx+1:]...)
	}
	return res
}
