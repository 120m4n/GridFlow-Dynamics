package crew

import (
	"testing"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
)

func TestNewService(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.crews == nil {
		t.Error("Service.crews should not be nil")
	}
}

func TestServiceRegisterCrew(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	crew := &domain.Crew{
		ID:          "crew-001",
		Name:        "Alpha Team",
		LeaderName:  "John Doe",
		MemberCount: 5,
		Status:      domain.CrewStatusAvailable,
	}

	service.RegisterCrew(crew)

	registeredCrew, exists := service.GetCrew("crew-001")
	if !exists {
		t.Error("Crew should exist after registration")
	}

	if registeredCrew.ID != "crew-001" {
		t.Errorf("Crew.ID = %s; want crew-001", registeredCrew.ID)
	}

	if registeredCrew.Name != "Alpha Team" {
		t.Errorf("Crew.Name = %s; want Alpha Team", registeredCrew.Name)
	}

	if registeredCrew.LastUpdate.IsZero() {
		t.Error("Crew.LastUpdate should be set after registration")
	}
}

func TestServiceGetCrew(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	// Test non-existent crew
	_, exists := service.GetCrew("non-existent")
	if exists {
		t.Error("GetCrew should return false for non-existent crew")
	}

	// Register and retrieve
	crew := &domain.Crew{
		ID:   "crew-002",
		Name: "Beta Team",
	}
	service.RegisterCrew(crew)

	retrieved, exists := service.GetCrew("crew-002")
	if !exists {
		t.Error("GetCrew should return true for existing crew")
	}

	if retrieved.ID != "crew-002" {
		t.Errorf("Retrieved crew ID = %s; want crew-002", retrieved.ID)
	}
}

func TestServiceGetAllCrews(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	// Initially empty
	crews := service.GetAllCrews()
	if len(crews) != 0 {
		t.Errorf("GetAllCrews should return empty slice initially, got %d crews", len(crews))
	}

	// Add crews
	service.RegisterCrew(&domain.Crew{ID: "crew-001", Name: "Alpha"})
	service.RegisterCrew(&domain.Crew{ID: "crew-002", Name: "Beta"})
	service.RegisterCrew(&domain.Crew{ID: "crew-003", Name: "Gamma"})

	crews = service.GetAllCrews()
	if len(crews) != 3 {
		t.Errorf("GetAllCrews should return 3 crews, got %d", len(crews))
	}
}

func TestServiceHandleLocationUpdate(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	// Create a location update JSON
	updateJSON := []byte(`{
		"crew_id": "crew-001",
		"location": {
			"latitude": 40.7128,
			"longitude": -74.0060,
			"accuracy": 10.0
		},
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	err := service.handleLocationUpdate(updateJSON)
	if err != nil {
		t.Fatalf("handleLocationUpdate failed: %v", err)
	}

	crew, exists := service.GetCrew("crew-001")
	if !exists {
		t.Error("Crew should be created after location update")
	}

	if crew.Location.Latitude != 40.7128 {
		t.Errorf("Crew.Location.Latitude = %f; want 40.7128", crew.Location.Latitude)
	}

	if crew.Location.Longitude != -74.0060 {
		t.Errorf("Crew.Location.Longitude = %f; want -74.0060", crew.Location.Longitude)
	}
}

func TestServiceHandleStatusUpdate(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	// Create a status update JSON
	updateJSON := []byte(`{
		"crew_id": "crew-001",
		"status": "working",
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	err := service.handleStatusUpdate(updateJSON)
	if err != nil {
		t.Fatalf("handleStatusUpdate failed: %v", err)
	}

	crew, exists := service.GetCrew("crew-001")
	if !exists {
		t.Error("Crew should be created after status update")
	}

	if crew.Status != domain.CrewStatusWorking {
		t.Errorf("Crew.Status = %s; want %s", crew.Status, domain.CrewStatusWorking)
	}
}

func TestServiceHandleLocationUpdateInvalidJSON(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	err := service.handleLocationUpdate([]byte("invalid json"))
	if err == nil {
		t.Error("handleLocationUpdate should fail with invalid JSON")
	}
}

func TestServiceHandleStatusUpdateInvalidJSON(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	err := service.handleStatusUpdate([]byte("invalid json"))
	if err == nil {
		t.Error("handleStatusUpdate should fail with invalid JSON")
	}
}

func TestServiceLocationUpdateExistingCrew(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	// Pre-register a crew
	service.RegisterCrew(&domain.Crew{
		ID:     "crew-001",
		Name:   "Alpha Team",
		Status: domain.CrewStatusAvailable,
	})

	// Update location
	updateJSON := []byte(`{
		"crew_id": "crew-001",
		"location": {
			"latitude": 41.0,
			"longitude": -75.0,
			"accuracy": 5.0
		},
		"timestamp": "2024-01-15T11:00:00Z"
	}`)

	err := service.handleLocationUpdate(updateJSON)
	if err != nil {
		t.Fatalf("handleLocationUpdate failed: %v", err)
	}

	crew, _ := service.GetCrew("crew-001")
	if crew.Name != "Alpha Team" {
		t.Errorf("Crew name should be preserved, got %s", crew.Name)
	}

	if crew.Location.Latitude != 41.0 {
		t.Errorf("Crew.Location.Latitude = %f; want 41.0", crew.Location.Latitude)
	}
}

func TestConcurrentCrewAccess(t *testing.T) {
	service := &Service{
		crews: make(map[string]*domain.Crew),
	}

	// Simulate concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			service.RegisterCrew(&domain.Crew{
				ID:   "crew-concurrent",
				Name: "Concurrent Team",
			})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetAllCrews()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetCrew("crew-concurrent")
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
