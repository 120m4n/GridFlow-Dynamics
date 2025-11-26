package domain

import (
	"testing"
	"time"
)

func TestCrewStatus(t *testing.T) {
	statuses := []CrewStatus{
		CrewStatusAvailable,
		CrewStatusEnRoute,
		CrewStatusOnSite,
		CrewStatusWorking,
		CrewStatusOffline,
	}

	expected := []string{"available", "en_route", "on_site", "working", "offline"}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("CrewStatus[%d] = %s; want %s", i, status, expected[i])
		}
	}
}

func TestCrew(t *testing.T) {
	now := time.Now()
	crew := Crew{
		ID:          "crew-001",
		Name:        "Alpha Team",
		LeaderName:  "John Doe",
		MemberCount: 5,
		Status:      CrewStatusAvailable,
		Location: Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Accuracy:  10.0,
		},
		LastUpdate: now,
	}

	if crew.ID != "crew-001" {
		t.Errorf("Crew.ID = %s; want crew-001", crew.ID)
	}

	if crew.Status != CrewStatusAvailable {
		t.Errorf("Crew.Status = %s; want %s", crew.Status, CrewStatusAvailable)
	}

	if crew.Location.Latitude != 40.7128 {
		t.Errorf("Crew.Location.Latitude = %f; want 40.7128", crew.Location.Latitude)
	}
}

func TestCrewLocationUpdate(t *testing.T) {
	now := time.Now()
	update := CrewLocationUpdate{
		CrewID: "crew-001",
		Location: Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Accuracy:  5.0,
		},
		Timestamp: now,
	}

	if update.CrewID != "crew-001" {
		t.Errorf("CrewLocationUpdate.CrewID = %s; want crew-001", update.CrewID)
	}

	if update.Location.Accuracy != 5.0 {
		t.Errorf("CrewLocationUpdate.Location.Accuracy = %f; want 5.0", update.Location.Accuracy)
	}
}
