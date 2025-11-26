package domain

import (
	"time"
)

// TrackingStatus represents the status of a crew during tracking.
type TrackingStatus string

const (
	TrackingStatusEnRoute    TrackingStatus = "en_route"
	TrackingStatusWorking    TrackingStatus = "working"
	TrackingStatusPaused     TrackingStatus = "paused"
	TrackingStatusCompleted  TrackingStatus = "completed"
)

// GPSCoordinates represents GPS location data.
type GPSCoordinates struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// ResourceConsumption represents material consumption data.
type ResourceConsumption struct {
	MaterialID   string  `json:"material_id,omitempty"`
	MaterialName string  `json:"material_name,omitempty"`
	Quantity     float64 `json:"quantity,omitempty"`
	Unit         string  `json:"unit,omitempty"`
}

// SafetyAlert represents a safety incident from the field.
type SafetyAlert struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

// TrackingPayload represents the JSON payload from mobile app.
type TrackingPayload struct {
	CrewID              string              `json:"crewId"`
	Timestamp           time.Time           `json:"timestamp"`
	GPSCoordinates      GPSCoordinates      `json:"gpsCoordinates"`
	TaskID              string              `json:"taskId"`
	Status              TrackingStatus      `json:"status"`
	ProgressPercentage  int                 `json:"progressPercentage"`
	ResourceConsumption ResourceConsumption `json:"resourceConsumption"`
	SafetyAlerts        []SafetyAlert       `json:"safetyAlerts"`
	BatteryLevel        int                 `json:"batteryLevel"`
	Region              string              `json:"region,omitempty"`
}

// TrackingEvent represents the event published to RabbitMQ.
type TrackingEvent struct {
	CrewID              string              `json:"crew_id"`
	Timestamp           time.Time           `json:"timestamp"`
	Location            Location            `json:"location"`
	TaskID              string              `json:"task_id"`
	Status              TrackingStatus      `json:"status"`
	ProgressPercentage  int                 `json:"progress_percentage"`
	ResourceConsumption ResourceConsumption `json:"resource_consumption"`
	SafetyAlerts        []SafetyAlert       `json:"safety_alerts"`
	BatteryLevel        int                 `json:"battery_level"`
	Region              string              `json:"region"`
	ReceivedAt          time.Time           `json:"received_at"`
}
