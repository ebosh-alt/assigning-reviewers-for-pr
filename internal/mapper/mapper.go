// Package mapper converts between domain models and transport DTOs.
package mapper

import (
	"assigning-reviewers-for-pr/internal/entities"
	oapi "assigning-reviewers-for-pr/internal/oapi"
)

// FromOAPITeam builds a entities.Team from transport DTO.
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
