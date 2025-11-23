// Package entities contains core business entities and errors.
package entities

import "errors"

var (
	// ErrUserNotFound is returned when a user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidArgument signals failed input validation.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrTeamExists signals team name conflict.
	ErrTeamExists = errors.New("team exists")
	// ErrTeamNotFound signals missing team.
	ErrTeamNotFound = errors.New("team not found")
	// ErrPRExists signals duplicate PR id.
	ErrPRExists = errors.New("pr exists")
	// ErrPRNotFound signals missing PR.
	ErrPRNotFound = errors.New("pr not found")
	// ErrPRMerged signals modification attempt after merge.
	ErrPRMerged = errors.New("pr merged")
	// ErrNotAssigned signals user not assigned to PR.
	ErrNotAssigned = errors.New("reviewer not assigned")
	// ErrNoCandidate signals absence of replacement candidate.
	ErrNoCandidate = errors.New("no candidate")
)
