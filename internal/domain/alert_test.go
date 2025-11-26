package domain

import (
	"testing"
	"time"
)

func TestAlertSeverity(t *testing.T) {
	severities := []AlertSeverity{
		AlertSeverityInfo,
		AlertSeverityWarning,
		AlertSeverityCritical,
		AlertSeverityEmergency,
	}

	expected := []string{"info", "warning", "critical", "emergency"}

	for i, severity := range severities {
		if string(severity) != expected[i] {
			t.Errorf("AlertSeverity[%d] = %s; want %s", i, severity, expected[i])
		}
	}
}

func TestAlertCategory(t *testing.T) {
	categories := []AlertCategory{
		AlertCategorySafety,
		AlertCategoryEquipment,
		AlertCategoryWeather,
		AlertCategoryLogistics,
		AlertCategoryCompliance,
	}

	expected := []string{"safety", "equipment", "weather", "logistics", "compliance"}

	for i, category := range categories {
		if string(category) != expected[i] {
			t.Errorf("AlertCategory[%d] = %s; want %s", i, category, expected[i])
		}
	}
}

func TestAlert(t *testing.T) {
	now := time.Now()
	alert := Alert{
		ID:          "alert-001",
		CrewID:      "crew-001",
		TaskID:      "task-001",
		Category:    AlertCategorySafety,
		Severity:    AlertSeverityCritical,
		Title:       "Safety Hazard Detected",
		Description: "Unstable ground conditions at work site",
		Location: Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Accuracy:  5.0,
		},
		CreatedAt: now,
	}

	if alert.ID != "alert-001" {
		t.Errorf("Alert.ID = %s; want alert-001", alert.ID)
	}

	if alert.Category != AlertCategorySafety {
		t.Errorf("Alert.Category = %s; want %s", alert.Category, AlertCategorySafety)
	}

	if alert.Severity != AlertSeverityCritical {
		t.Errorf("Alert.Severity = %s; want %s", alert.Severity, AlertSeverityCritical)
	}

	if alert.AckedAt != nil {
		t.Error("Alert.AckedAt should be nil for new alert")
	}

	if alert.ResolvedAt != nil {
		t.Error("Alert.ResolvedAt should be nil for new alert")
	}
}

func TestAlertEvent(t *testing.T) {
	now := time.Now()
	event := AlertEvent{
		CrewID:      "crew-001",
		TaskID:      "task-001",
		Category:    AlertCategoryEquipment,
		Severity:    AlertSeverityWarning,
		Title:       "Equipment Malfunction",
		Description: "Drill motor overheating",
		Location: Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Accuracy:  10.0,
		},
		Timestamp: now,
	}

	if event.CrewID != "crew-001" {
		t.Errorf("AlertEvent.CrewID = %s; want crew-001", event.CrewID)
	}

	if event.Category != AlertCategoryEquipment {
		t.Errorf("AlertEvent.Category = %s; want %s", event.Category, AlertCategoryEquipment)
	}
}
