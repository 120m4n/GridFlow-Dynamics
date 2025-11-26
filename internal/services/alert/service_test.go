package alert

import (
	"testing"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
)

func TestNewService(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.alerts == nil {
		t.Error("Service.alerts should not be nil")
	}
}

func TestServiceHandleAlertCreated(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	eventJSON := []byte(`{
		"crew_id": "crew-001",
		"task_id": "task-001",
		"category": "safety",
		"severity": "critical",
		"title": "Safety Hazard",
		"description": "Unstable ground conditions",
		"location": {
			"latitude": 40.7128,
			"longitude": -74.0060,
			"accuracy": 5.0
		},
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	err := service.handleAlertCreated(eventJSON)
	if err != nil {
		t.Fatalf("handleAlertCreated failed: %v", err)
	}

	alerts := service.GetAllAlerts()
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.CrewID != "crew-001" {
		t.Errorf("Alert.CrewID = %s; want crew-001", alert.CrewID)
	}

	if alert.Category != domain.AlertCategorySafety {
		t.Errorf("Alert.Category = %s; want %s", alert.Category, domain.AlertCategorySafety)
	}

	if alert.Severity != domain.AlertSeverityCritical {
		t.Errorf("Alert.Severity = %s; want %s", alert.Severity, domain.AlertSeverityCritical)
	}

	if alert.AckedAt != nil {
		t.Error("New alert should not be acknowledged")
	}

	if alert.ResolvedAt != nil {
		t.Error("New alert should not be resolved")
	}
}

func TestServiceHandleAlertAcked(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// First create an alert
	eventJSON := []byte(`{
		"crew_id": "crew-001",
		"category": "equipment",
		"severity": "warning",
		"title": "Equipment Issue",
		"description": "Motor overheating",
		"location": {"latitude": 40.0, "longitude": -74.0, "accuracy": 10.0},
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	_ = service.handleAlertCreated(eventJSON)

	alerts := service.GetAllAlerts()
	alertID := alerts[0].ID

	// Acknowledge the alert
	ackJSON := []byte(`{
		"alert_id": "` + alertID + `",
		"timestamp": "2024-01-15T10:35:00Z"
	}`)

	err := service.handleAlertAcked(ackJSON)
	if err != nil {
		t.Fatalf("handleAlertAcked failed: %v", err)
	}

	alert, _ := service.GetAlert(alertID)
	if alert.AckedAt == nil {
		t.Error("Alert should be acknowledged")
	}
}

func TestServiceHandleAlertResolved(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// First create an alert
	eventJSON := []byte(`{
		"crew_id": "crew-001",
		"category": "logistics",
		"severity": "info",
		"title": "Material Delay",
		"description": "Materials arriving late",
		"location": {"latitude": 40.0, "longitude": -74.0, "accuracy": 10.0},
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	_ = service.handleAlertCreated(eventJSON)

	alerts := service.GetAllAlerts()
	alertID := alerts[0].ID

	// Resolve the alert
	resolveJSON := []byte(`{
		"alert_id": "` + alertID + `",
		"timestamp": "2024-01-15T11:00:00Z"
	}`)

	err := service.handleAlertResolved(resolveJSON)
	if err != nil {
		t.Fatalf("handleAlertResolved failed: %v", err)
	}

	alert, _ := service.GetAlert(alertID)
	if alert.ResolvedAt == nil {
		t.Error("Alert should be resolved")
	}
}

func TestServiceHandleAlertCreatedInvalidJSON(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	err := service.handleAlertCreated([]byte("invalid json"))
	if err == nil {
		t.Error("handleAlertCreated should fail with invalid JSON")
	}
}

func TestServiceHandleAlertAckedInvalidJSON(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	err := service.handleAlertAcked([]byte("invalid json"))
	if err == nil {
		t.Error("handleAlertAcked should fail with invalid JSON")
	}
}

func TestServiceHandleAlertResolvedInvalidJSON(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	err := service.handleAlertResolved([]byte("invalid json"))
	if err == nil {
		t.Error("handleAlertResolved should fail with invalid JSON")
	}
}

func TestServiceGetAlert(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// Test non-existent alert
	_, exists := service.GetAlert("non-existent")
	if exists {
		t.Error("GetAlert should return false for non-existent alert")
	}
}

func TestServiceGetAllAlerts(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// Initially empty
	alerts := service.GetAllAlerts()
	if len(alerts) != 0 {
		t.Errorf("GetAllAlerts should return empty slice initially, got %d alerts", len(alerts))
	}
}

func TestServiceGetActiveAlerts(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// Create alerts
	for i := 0; i < 3; i++ {
		eventJSON := []byte(`{
			"crew_id": "crew-001",
			"category": "safety",
			"severity": "warning",
			"title": "Test Alert",
			"description": "Test description",
			"location": {"latitude": 40.0, "longitude": -74.0, "accuracy": 10.0},
			"timestamp": "2024-01-15T10:30:00Z"
		}`)
		_ = service.handleAlertCreated(eventJSON)
	}

	// All should be active initially
	activeAlerts := service.GetActiveAlerts()
	if len(activeAlerts) != 3 {
		t.Errorf("Expected 3 active alerts, got %d", len(activeAlerts))
	}

	// Resolve one
	alerts := service.GetAllAlerts()
	resolveJSON := []byte(`{
		"alert_id": "` + alerts[0].ID + `",
		"timestamp": "2024-01-15T11:00:00Z"
	}`)
	_ = service.handleAlertResolved(resolveJSON)

	activeAlerts = service.GetActiveAlerts()
	if len(activeAlerts) != 2 {
		t.Errorf("Expected 2 active alerts after resolving one, got %d", len(activeAlerts))
	}
}

func TestServiceGetAlertsBySeverity(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// Create alerts with different severities
	severities := []string{"critical", "warning", "critical", "info"}
	for _, severity := range severities {
		eventJSON := []byte(`{
			"crew_id": "crew-001",
			"category": "safety",
			"severity": "` + severity + `",
			"title": "Test Alert",
			"description": "Test description",
			"location": {"latitude": 40.0, "longitude": -74.0, "accuracy": 10.0},
			"timestamp": "2024-01-15T10:30:00Z"
		}`)
		_ = service.handleAlertCreated(eventJSON)
	}

	criticalAlerts := service.GetAlertsBySeverity(domain.AlertSeverityCritical)
	if len(criticalAlerts) != 2 {
		t.Errorf("Expected 2 critical alerts, got %d", len(criticalAlerts))
	}

	warningAlerts := service.GetAlertsBySeverity(domain.AlertSeverityWarning)
	if len(warningAlerts) != 1 {
		t.Errorf("Expected 1 warning alert, got %d", len(warningAlerts))
	}

	infoAlerts := service.GetAlertsBySeverity(domain.AlertSeverityInfo)
	if len(infoAlerts) != 1 {
		t.Errorf("Expected 1 info alert, got %d", len(infoAlerts))
	}
}

func TestServiceGetAlertsByCrew(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// Create alerts from different crews
	crews := []string{"crew-001", "crew-001", "crew-002"}
	for _, crewID := range crews {
		eventJSON := []byte(`{
			"crew_id": "` + crewID + `",
			"category": "safety",
			"severity": "warning",
			"title": "Test Alert",
			"description": "Test description",
			"location": {"latitude": 40.0, "longitude": -74.0, "accuracy": 10.0},
			"timestamp": "2024-01-15T10:30:00Z"
		}`)
		_ = service.handleAlertCreated(eventJSON)
	}

	crew1Alerts := service.GetAlertsByCrew("crew-001")
	if len(crew1Alerts) != 2 {
		t.Errorf("Expected 2 alerts from crew-001, got %d", len(crew1Alerts))
	}

	crew2Alerts := service.GetAlertsByCrew("crew-002")
	if len(crew2Alerts) != 1 {
		t.Errorf("Expected 1 alert from crew-002, got %d", len(crew2Alerts))
	}

	crew3Alerts := service.GetAlertsByCrew("crew-003")
	if len(crew3Alerts) != 0 {
		t.Errorf("Expected 0 alerts from crew-003, got %d", len(crew3Alerts))
	}
}

func TestConcurrentAlertAccess(t *testing.T) {
	service := &Service{
		alerts: make(map[string]*domain.Alert),
	}

	// Simulate concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			eventJSON := []byte(`{
				"crew_id": "crew-001",
				"category": "safety",
				"severity": "warning",
				"title": "Test Alert",
				"description": "Test description",
				"location": {"latitude": 40.0, "longitude": -74.0, "accuracy": 10.0},
				"timestamp": "2024-01-15T10:30:00Z"
			}`)
			_ = service.handleAlertCreated(eventJSON)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetAllAlerts()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetActiveAlerts()
		}
		done <- true
	}()

	// Wait for all goroutines with timeout
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}
