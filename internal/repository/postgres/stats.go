package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"assigning-reviewers-for-pr/internal/entities"

	"github.com/jackc/pgx/v5"
)

const (
	statsByUserQuery    = `SELECT reviewer_id, COUNT(*) FROM pr_reviewers GROUP BY reviewer_id`
	statsByPRQuery      = `SELECT pr_id, COUNT(*) FROM pr_reviewers GROUP BY pr_id`
	statsByStatusQuery  = `SELECT status, COUNT(*) FROM pull_requests GROUP BY status`
	statsByTeamQuery    = `SELECT t.name, COUNT(*) FROM pr_reviewers r JOIN users u ON u.id = r.reviewer_id JOIN teams t ON t.id = u.team_id GROUP BY t.name`
	reviewerExistsQuery = `SELECT true FROM users WHERE id=$1`
	reviewerAssigns     = `SELECT COUNT(*) FROM pr_reviewers WHERE reviewer_id=$1`
	reviewerStatus      = `
SELECT pr.status, COUNT(*)
FROM pr_reviewers r
JOIN pull_requests pr ON pr.id = r.pr_id
WHERE r.reviewer_id=$1
GROUP BY pr.status`
	reviewerRecent = `
SELECT pr.id, pr.name, pr.author_id, pr.status
FROM pr_reviewers r
JOIN pull_requests pr ON pr.id = r.pr_id
WHERE r.reviewer_id=$1
ORDER BY pr.created_at DESC
LIMIT $2`
	prStatsQuery     = `SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id=$1`
	prReviewersQuery = `SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1`
	prHistoryQuery   = `SELECT old_reviewer_id, new_reviewer_id, changed_at FROM pr_reassignment_history WHERE pr_id=$1 ORDER BY changed_at DESC`
)

// Stats returns assignments grouped by user and PR.
func (p *Postgres) Stats(ctx context.Context) (entities.Stats, error) {
	res := entities.Stats{}

	rows, err := p.db.Query(ctx, statsByUserQuery)
	if err != nil {
		return res, fmt.Errorf("stats by user: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s entities.UserStat
		if err := rows.Scan(&s.UserID, &s.AssignCnt); err != nil {
			return res, fmt.Errorf("scan user stat: %w", err)
		}
		res.ByUser = append(res.ByUser, s)
	}
	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("iterate user stat: %w", err)
	}

	rows2, err := p.db.Query(ctx, statsByPRQuery)
	if err != nil {
		return res, fmt.Errorf("stats by pr: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var s entities.PRStat
		if err := rows2.Scan(&s.PRID, &s.AssignCnt); err != nil {
			return res, fmt.Errorf("scan pr stat: %w", err)
		}
		res.ByPR = append(res.ByPR, s)
	}
	if err := rows2.Err(); err != nil {
		return res, fmt.Errorf("iterate pr stat: %w", err)
	}

	rows3, err := p.db.Query(ctx, statsByStatusQuery)
	if err != nil {
		return res, fmt.Errorf("stats by status: %w", err)
	}
	defer rows3.Close()
	for rows3.Next() {
		var s entities.StatusStat
		if err := rows3.Scan(&s.Status, &s.PRCount); err != nil {
			return res, fmt.Errorf("scan status stat: %w", err)
		}
		res.ByStatus = append(res.ByStatus, s)
	}
	if err := rows3.Err(); err != nil {
		return res, fmt.Errorf("iterate status stat: %w", err)
	}

	rows4, err := p.db.Query(ctx, statsByTeamQuery)
	if err != nil {
		return res, fmt.Errorf("stats by team: %w", err)
	}
	defer rows4.Close()
	for rows4.Next() {
		var s entities.TeamStat
		if err := rows4.Scan(&s.TeamName, &s.AssignCnt); err != nil {
			return res, fmt.Errorf("scan team stat: %w", err)
		}
		res.ByTeam = append(res.ByTeam, s)
	}
	if err := rows4.Err(); err != nil {
		return res, fmt.Errorf("iterate team stat: %w", err)
	}

	return res, nil
}

// StatsSummary returns filtered stats snapshot.
func (p *Postgres) StatsSummary(ctx context.Context, filter entities.StatsFilter) (entities.StatsSummary, error) {
	res := entities.StatsSummary{}

	whereClause, args := buildPRFilter(filter)
	limitValue := filter.Limit
	if limitValue <= 0 {
		limitValue = 10
	}

	limitIdx := len(args) + 1
	topArgs := append([]any{}, args...)
	topArgs = append(topArgs, limitValue)

	var b strings.Builder
	b.WriteString("SELECT r.reviewer_id, COUNT(*) AS cnt FROM pr_reviewers r JOIN pull_requests pr ON pr.id = r.pr_id")
	if whereClause != "" {
		b.WriteByte(' ')
		b.WriteString(whereClause)
	}
	b.WriteString(" GROUP BY r.reviewer_id ORDER BY cnt DESC LIMIT $")
	b.WriteString(strconv.Itoa(limitIdx))

	rows, err := p.db.Query(ctx, b.String(), topArgs...)
	if err != nil {
		return res, fmt.Errorf("summary top reviewers: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s entities.UserStat
		if err := rows.Scan(&s.UserID, &s.AssignCnt); err != nil {
			return res, fmt.Errorf("scan top reviewers: %w", err)
		}
		res.TopReviewers = append(res.TopReviewers, s)
	}
	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("iterate top reviewers: %w", err)
	}

	statusQuery := strings.Builder{}
	statusQuery.WriteString("SELECT pr.status, COUNT(*) FROM pull_requests pr")
	if whereClause != "" {
		statusQuery.WriteByte(' ')
		statusQuery.WriteString(whereClause)
	}
	statusQuery.WriteString(" GROUP BY pr.status")

	rowsStatus, err := p.db.Query(ctx, statusQuery.String(), args...)
	if err != nil {
		return res, fmt.Errorf("summary status: %w", err)
	}
	defer rowsStatus.Close()
	for rowsStatus.Next() {
		var s entities.StatusStat
		if err := rowsStatus.Scan(&s.Status, &s.PRCount); err != nil {
			return res, fmt.Errorf("scan status summary: %w", err)
		}
		res.PRStatusCounts = append(res.PRStatusCounts, s)
	}
	if err := rowsStatus.Err(); err != nil {
		return res, fmt.Errorf("iterate status summary: %w", err)
	}

	teamQuery := strings.Builder{}
	teamQuery.WriteString("SELECT t.name, COUNT(*) AS assign_cnt FROM pr_reviewers r JOIN users u ON u.id = r.reviewer_id JOIN teams t ON t.id = u.team_id JOIN pull_requests pr ON pr.id = r.pr_id")
	if whereClause != "" {
		teamQuery.WriteByte(' ')
		teamQuery.WriteString(whereClause)
	}
	teamQuery.WriteString(" GROUP BY t.name ORDER BY assign_cnt DESC")
	rowsTeam, err := p.db.Query(ctx, teamQuery.String(), args...)
	if err != nil {
		return res, fmt.Errorf("summary teams: %w", err)
	}
	defer rowsTeam.Close()
	for rowsTeam.Next() {
		var s entities.TeamStat
		if err := rowsTeam.Scan(&s.TeamName, &s.AssignCnt); err != nil {
			return res, fmt.Errorf("scan team summary: %w", err)
		}
		res.TeamAssignments = append(res.TeamAssignments, s)
	}
	if err := rowsTeam.Err(); err != nil {
		return res, fmt.Errorf("iterate team summary: %w", err)
	}

	return res, nil
}

// ReviewerStats returns per-user stats.
func (p *Postgres) ReviewerStats(ctx context.Context, userID string, limit int) (entities.ReviewerStats, error) {
	res := entities.ReviewerStats{UserID: userID}
	var exists bool
	if err := p.db.QueryRow(ctx, reviewerExistsQuery, userID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, entities.ErrUserNotFound
		}
		return res, fmt.Errorf("check user: %w", err)
	}

	if err := p.db.QueryRow(ctx, reviewerAssigns, userID).Scan(&res.AssignCnt); err != nil {
		return res, fmt.Errorf("count assignments: %w", err)
	}

	statusRows, err := p.db.Query(ctx, reviewerStatus, userID)
	if err != nil {
		return res, fmt.Errorf("reviewer status counts: %w", err)
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var status entities.PullRequestStatus
		var cnt int64
		if err := statusRows.Scan(&status, &cnt); err != nil {
			return res, fmt.Errorf("scan reviewer status: %w", err)
		}
		switch status {
		case entities.StatusOpen:
			res.OpenPRCnt = cnt
		case entities.StatusMerged:
			res.MergedPRCnt = cnt
		}
	}
	if err := statusRows.Err(); err != nil {
		return res, fmt.Errorf("iterate reviewer status: %w", err)
	}

	recentRows, err := p.db.Query(ctx, reviewerRecent, userID, limit)
	if err != nil {
		return res, fmt.Errorf("reviewer recent prs: %w", err)
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var pr entities.PullRequestShort
		if err := recentRows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return res, fmt.Errorf("scan reviewer prs: %w", err)
		}
		res.RecentPRs = append(res.RecentPRs, pr)
	}
	if err := recentRows.Err(); err != nil {
		return res, fmt.Errorf("iterate reviewer prs: %w", err)
	}

	return res, nil
}

// PRStats returns statistics for a single PR.
func (p *Postgres) PRStats(ctx context.Context, prID string) (entities.PRStats, error) {
	var res entities.PRStats
	var createdAt time.Time
	var mergedAt *time.Time
	if err := p.db.QueryRow(ctx, prStatsQuery, prID).
		Scan(&res.PRID, &res.Name, &res.AuthorID, &res.Status, &createdAt, &mergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, entities.ErrPRNotFound
		}
		return res, fmt.Errorf("pr stats: %w", err)
	}
	res.CreatedAt = &createdAt
	res.MergedAt = mergedAt

	revRows, err := p.db.Query(ctx, prReviewersQuery, prID)
	if err != nil {
		return res, fmt.Errorf("pr reviewers: %w", err)
	}
	defer revRows.Close()
	for revRows.Next() {
		var id string
		if err := revRows.Scan(&id); err != nil {
			return res, fmt.Errorf("scan pr reviewer: %w", err)
		}
		res.Reviewers = append(res.Reviewers, id)
	}
	if err := revRows.Err(); err != nil {
		return res, fmt.Errorf("iterate pr reviewers: %w", err)
	}

	histRows, err := p.db.Query(ctx, prHistoryQuery, prID)
	if err != nil {
		return res, fmt.Errorf("pr history: %w", err)
	}
	defer histRows.Close()
	for histRows.Next() {
		var ev entities.ReassignmentEvent
		var newReviewer sql.NullString
		if err := histRows.Scan(&ev.OldReviewerID, &newReviewer, &ev.ChangedAt); err != nil {
			return res, fmt.Errorf("scan history: %w", err)
		}
		if newReviewer.Valid {
			ev.NewReviewerID = &newReviewer.String
		}
		res.Reassignments = append(res.Reassignments, ev)
	}
	if err := histRows.Err(); err != nil {
		return res, fmt.Errorf("iterate history: %w", err)
	}

	res.TransferCount = int64(len(res.Reassignments))
	return res, nil
}

func buildPRFilter(filter entities.StatsFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)
	idx := 1
	if filter.From != nil {
		conditions = append(conditions, "pr.created_at >= $"+strconv.Itoa(idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		conditions = append(conditions, "pr.created_at <= $"+strconv.Itoa(idx))
		args = append(args, *filter.To)
		idx++
	}
	if filter.Status != nil {
		conditions = append(conditions, "pr.status = $"+strconv.Itoa(idx))
		args = append(args, *filter.Status)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
