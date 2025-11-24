// Package mapper converts between domain models and transport DTOs.
package mapper

import (
	"assigning-reviewers-for-pr/internal/entities"
	oapi "assigning-reviewers-for-pr/internal/oapi"
)

// FromOAPITeam builds an entities.Team from transport DTO.
func FromOAPITeam(src oapi.Team) entities.Team {
	members := make([]entities.User, 0, len(src.Members))
	for _, m := range src.Members {
		members = append(members, entities.User{
			ID:       m.UserId,
			Username: m.Username,
			TeamName: src.TeamName,
			IsActive: m.IsActive,
		})
	}

	return entities.Team{
		Name:    src.TeamName,
		Members: members,
	}
}

// ToOAPITeam maps entities.Team to transport model.
func ToOAPITeam(team entities.Team) oapi.Team {
	members := make([]oapi.TeamMember, 0, len(team.Members))
	for _, m := range team.Members {
		members = append(members, oapi.TeamMember{
			UserId:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	return oapi.Team{
		TeamName: team.Name,
		Members:  members,
	}
}

// ToOAPIUser maps entities.User to transport model.
func ToOAPIUser(u entities.User) oapi.User {
	return oapi.User{
		UserId:   u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

// ToOAPIPull maps entities.PullRequest to transport model.
func ToOAPIPull(pr entities.PullRequest) oapi.PullRequest {
	return oapi.PullRequest{
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorId:          pr.AuthorID,
		Status:            oapi.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.Reviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

// ToOAPIPullShort maps entities.PullRequestShort to transport model.
func ToOAPIPullShort(pr entities.PullRequestShort) oapi.PullRequestShort {
	return oapi.PullRequestShort{
		PullRequestId:   pr.ID,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          oapi.PullRequestShortStatus(pr.Status),
	}
}

// ToOAPIPullShortList maps a slice of entities.PullRequestShort to transport slice.
func ToOAPIPullShortList(list []entities.PullRequestShort) []oapi.PullRequestShort {
	res := make([]oapi.PullRequestShort, 0, len(list))
	for _, pr := range list {
		res = append(res, ToOAPIPullShort(pr))
	}
	return res
}

// ToOAPIStats maps aggregated statistics to transport model.
func ToOAPIStats(src entities.Stats) oapi.Stats {
	byUser := make([]oapi.UserStat, 0, len(src.ByUser))
	for _, s := range src.ByUser {
		userID, cnt := s.UserID, s.AssignCnt
		byUser = append(byUser, oapi.UserStat{UserId: &userID, AssignCnt: &cnt})
	}

	byPR := make([]oapi.PRStat, 0, len(src.ByPR))
	for _, s := range src.ByPR {
		prID, cnt := s.PRID, s.AssignCnt
		byPR = append(byPR, oapi.PRStat{PrId: &prID, AssignCnt: &cnt})
	}

	byStatus := make([]oapi.StatusStat, 0, len(src.ByStatus))
	for _, s := range src.ByStatus {
		status, cnt := oapi.StatusStatStatus(s.Status), s.PRCount
		byStatus = append(byStatus, oapi.StatusStat{Status: &status, PrCount: &cnt})
	}

	byTeam := make([]oapi.TeamStat, 0, len(src.ByTeam))
	for _, s := range src.ByTeam {
		team, cnt := s.TeamName, s.AssignCnt
		byTeam = append(byTeam, oapi.TeamStat{TeamName: &team, AssignCnt: &cnt})
	}

	return oapi.Stats{
		ByUser:   &byUser,
		ByPr:     &byPR,
		ByStatus: &byStatus,
		ByTeam:   &byTeam,
	}
}

// ToOAPIStatsSummary maps filtered statistics to transport model.
func ToOAPIStatsSummary(src entities.StatsSummary) oapi.StatsSummary {
	top := make([]oapi.UserStat, 0, len(src.TopReviewers))
	for _, s := range src.TopReviewers {
		userID, cnt := s.UserID, s.AssignCnt
		top = append(top, oapi.UserStat{UserId: &userID, AssignCnt: &cnt})
	}

	status := make([]oapi.StatusStat, 0, len(src.PRStatusCounts))
	for _, s := range src.PRStatusCounts {
		st, cnt := oapi.StatusStatStatus(s.Status), s.PRCount
		status = append(status, oapi.StatusStat{Status: &st, PrCount: &cnt})
	}

	teams := make([]oapi.TeamStat, 0, len(src.TeamAssignments))
	for _, s := range src.TeamAssignments {
		name, cnt := s.TeamName, s.AssignCnt
		teams = append(teams, oapi.TeamStat{TeamName: &name, AssignCnt: &cnt})
	}

	return oapi.StatsSummary{
		TopReviewers:    &top,
		PrStatusCounts:  &status,
		TeamAssignments: &teams,
	}
}

// ToOAPIReviewerStats maps per-user stats to transport DTO.
func ToOAPIReviewerStats(src entities.ReviewerStats) oapi.ReviewerStats {
	userID, assign, openCnt, mergedCnt := src.UserID, src.AssignCnt, src.OpenPRCnt, src.MergedPRCnt
	recent := ToOAPIPullShortList(src.RecentPRs)
	return oapi.ReviewerStats{
		UserId:      &userID,
		AssignCnt:   &assign,
		OpenPrCnt:   &openCnt,
		MergedPrCnt: &mergedCnt,
		RecentPrs:   &recent,
	}
}

// ToOAPIPRStats maps PR stats to transport DTO.
func ToOAPIPRStats(src entities.PRStats) oapi.PRStats {
	prID, name, author := src.PRID, src.Name, src.AuthorID
	status := oapi.PRStatsStatus(src.Status)
	reviewers := make([]string, len(src.Reviewers))
	copy(reviewers, src.Reviewers)

	reassignments := make([]oapi.ReassignmentEvent, 0, len(src.Reassignments))
	for _, r := range src.Reassignments {
		oldID, newID := r.OldReviewerID, r.NewReviewerID
		reassignments = append(reassignments, oapi.ReassignmentEvent{
			OldReviewerId: &oldID,
			NewReviewerId: newID,
			ChangedAt:     &r.ChangedAt,
		})
	}

	return oapi.PRStats{
		PrId:          &prID,
		PrName:        &name,
		AuthorId:      &author,
		Status:        &status,
		Reviewers:     &reviewers,
		CreatedAt:     src.CreatedAt,
		MergedAt:      src.MergedAt,
		Reassignments: &reassignments,
		TransferCnt:   &src.TransferCount,
	}
}
