// Package entities contains core business entities.
package entities

// Team aggregates members under a team name.
type Team struct {
	Name    string
	Members []User
}
