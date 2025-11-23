// Package entities contains core business entities.
package entities

import "time"

// Stats aggregates counters by user, PR, status and team.
type Stats struct {
	ByUser   []UserStat   `json:"by_user"`
	ByPR     []PRStat     `json:"by_pr"`
	ByStatus []StatusStat `json:"by_status"`
	ByTeam   []TeamStat   `json:"by_team"`
}

// StatsSummary is a filtered snapshot of activity.
type StatsSummary struct {
	TopReviewers    []UserStat   `json:"top_reviewers"`
	PRStatusCounts  []StatusStat `json:"pr_status_counts"`
	TeamAssignments []TeamStat   `json:"team_assignments"`
}

// StatsFilter limits stats by time range/status.
type StatsFilter struct {
	From   *time.Time
	To     *time.Time
	Status *PullRequestStatus
	Limit  int
}

// ReviewerStats contains aggregated data for a single reviewer.
type ReviewerStats struct {
	UserID      string             `json:"user_id"`
	AssignCnt   int64              `json:"assign_cnt"`
	OpenPRCnt   int64              `json:"open_pr_cnt"`
	MergedPRCnt int64              `json:"merged_pr_cnt"`
	RecentPRs   []PullRequestShort `json:"recent_prs"`
}

// ReassignmentEvent captures a reviewer replacement.
type ReassignmentEvent struct {
	OldReviewerID string    `json:"old_reviewer_id"`
	NewReviewerID *string   `json:"new_reviewer_id,omitempty"`
	ChangedAt     time.Time `json:"changed_at"`
}

// PRStats contains statistics about a specific PR.
type PRStats struct {
	PRID          string              `json:"pr_id"`
	Name          string              `json:"pr_name"`
	AuthorID      string              `json:"author_id"`
	Status        PullRequestStatus   `json:"status"`
	Reviewers     []string            `json:"reviewers"`
	CreatedAt     *time.Time          `json:"created_at,omitempty"`
	MergedAt      *time.Time          `json:"merged_at,omitempty"`
	Reassignments []ReassignmentEvent `json:"reassignments"`
	TransferCount int64               `json:"transfer_cnt"`
}

// UserStat contains assignments count per reviewer.
type UserStat struct {
	UserID    string `json:"user_id"`
	AssignCnt int64  `json:"assign_cnt"`
}

// PRStat contains reviewer count per PR.
type PRStat struct {
	PRID      string `json:"pr_id"`
	AssignCnt int64  `json:"assign_cnt"`
}

// StatusStat describes PR counts grouped by status.
type StatusStat struct {
	Status  PullRequestStatus `json:"status"`
	PRCount int64             `json:"pr_count"`
}

// TeamStat aggregates assignments grouped by team name.
type TeamStat struct {
	TeamName  string `json:"team_name"`
	AssignCnt int64  `json:"assign_cnt"`
}

// DeactivateResult contains info about bulk deactivation outcome.
type DeactivateResult struct {
	DeactivatedUsers int `json:"deactivated_users"`
	Reassigned       int `json:"reassigned"`
	Removed          int `json:"removed"`
}
