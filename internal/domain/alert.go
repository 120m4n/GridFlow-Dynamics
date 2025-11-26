package domain

import (
	"time"
)

// AlertSeverity represents the severity level of an alert.
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)

// AlertCategory represents the category of an alert.
type AlertCategory string

const (
	AlertCategorySafety      AlertCategory = "safety"
	AlertCategoryEquipment   AlertCategory = "equipment"
	AlertCategoryWeather     AlertCategory = "weather"
	AlertCategoryLogistics   AlertCategory = "logistics"
	AlertCategoryCompliance  AlertCategory = "compliance"
)

// Alert represents an alert from the field.
type Alert struct {
	ID          string        `json:"id"`
	CrewID      string        `json:"crew_id"`
	TaskID      string        `json:"task_id,omitempty"`
	Category    AlertCategory `json:"category"`
	Severity    AlertSeverity `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Location    Location      `json:"location"`
	CreatedAt   time.Time     `json:"created_at"`
	AckedAt     *time.Time    `json:"acked_at,omitempty"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
}

// AlertEvent represents an alert event from a crew.
type AlertEvent struct {
	CrewID      string        `json:"crew_id"`
	TaskID      string        `json:"task_id,omitempty"`
	Category    AlertCategory `json:"category"`
	Severity    AlertSeverity `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Location    Location      `json:"location"`
	Timestamp   time.Time     `json:"timestamp"`
}
