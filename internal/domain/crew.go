// Package domain contains domain models for the GridFlow-Dynamics platform.
package domain

import (
	"time"
)

// CrewStatus represents the current status of a crew.
type CrewStatus string

const (
	CrewStatusAvailable  CrewStatus = "available"
	CrewStatusEnRoute    CrewStatus = "en_route"
	CrewStatusOnSite     CrewStatus = "on_site"
	CrewStatusWorking    CrewStatus = "working"
	CrewStatusOffline    CrewStatus = "offline"
)

// Location represents GPS coordinates.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"`
}

// Crew represents a field crew.
type Crew struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	LeaderName  string     `json:"leader_name"`
	MemberCount int        `json:"member_count"`
	Status      CrewStatus `json:"status"`
	Location    Location   `json:"location"`
	LastUpdate  time.Time  `json:"last_update"`
}

// CrewLocationUpdate represents a location update event from a crew.
type CrewLocationUpdate struct {
	CrewID    string    `json:"crew_id"`
	Location  Location  `json:"location"`
	Timestamp time.Time `json:"timestamp"`
}
