package domain

import (
	"testing"
	"time"
)

func TestTrackingStatus(t *testing.T) {
	statuses := []TrackingStatus{
		TrackingStatusEnRoute,
		TrackingStatusWorking,
		TrackingStatusPaused,
		TrackingStatusCompleted,
	}

	expected := []string{"en_route", "working", "paused", "completed"}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("TrackingStatus[%d] = %s; want %s", i, status, expected[i])
		}
	}
}

func TestGPSCoordinates(t *testing.T) {
	gps := GPSCoordinates{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	if gps.Latitude != 40.7128 {
		t.Errorf("Latitude = %f; want 40.7128", gps.Latitude)
	}

	if gps.Longitude != -74.0060 {
		t.Errorf("Longitude = %f; want -74.0060", gps.Longitude)
	}
}

func TestResourceConsumption(t *testing.T) {
	rc := ResourceConsumption{
		MaterialID:   "MAT-001",
		MaterialName: "Copper Wire",
		Quantity:     100.5,
		Unit:         "meters",
	}

	if rc.MaterialID != "MAT-001" {
		t.Errorf("MaterialID = %s; want MAT-001", rc.MaterialID)
	}
}

func TestSafetyAlert(t *testing.T) {
	now := time.Now()
	alert := SafetyAlert{
		Type:        "hazard",
		Description: "Unstable ground",
		Severity:    "high",
		Timestamp:   now,
	}

	if alert.Type != "hazard" {
		t.Errorf("Type = %s; want hazard", alert.Type)
	}

	if alert.Severity != "high" {
		t.Errorf("Severity = %s; want high", alert.Severity)
	}
}

func TestTrackingPayload(t *testing.T) {
	now := time.Now()
	payload := TrackingPayload{
		CrewID:    "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: now,
		GPSCoordinates: GPSCoordinates{
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
		TaskID:             "TASK-001",
		Status:             TrackingStatusWorking,
		ProgressPercentage: 75,
		ResourceConsumption: ResourceConsumption{
			MaterialID: "MAT-001",
			Quantity:   50,
		},
		SafetyAlerts: []SafetyAlert{
			{Type: "warning", Description: "Low visibility"},
		},
		BatteryLevel: 85,
		Region:       "north",
	}

	if payload.CrewID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("CrewID mismatch")
	}

	if payload.Status != TrackingStatusWorking {
		t.Errorf("Status = %s; want %s", payload.Status, TrackingStatusWorking)
	}

	if payload.ProgressPercentage != 75 {
		t.Errorf("ProgressPercentage = %d; want 75", payload.ProgressPercentage)
	}

	if len(payload.SafetyAlerts) != 1 {
		t.Errorf("SafetyAlerts length = %d; want 1", len(payload.SafetyAlerts))
	}
}

func TestTrackingEvent(t *testing.T) {
	now := time.Now()
	event := TrackingEvent{
		CrewID:    "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: now,
		Location: Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
		TaskID:             "TASK-001",
		Status:             TrackingStatusEnRoute,
		ProgressPercentage: 0,
		BatteryLevel:       90,
		Region:             "south",
		ReceivedAt:         now,
	}

	if event.CrewID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("CrewID mismatch")
	}

	if event.Region != "south" {
		t.Errorf("Region = %s; want south", event.Region)
	}

	if event.ReceivedAt.IsZero() {
		t.Error("ReceivedAt should not be zero")
	}
}
