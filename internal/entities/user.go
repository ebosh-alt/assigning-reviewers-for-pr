// Package entities contains core business entities.
package entities

// User is a domain representation of a team member.
type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}
