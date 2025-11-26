// Package crew provides the crew tracking microservice.
package crew

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
	"github.com/120m4n/GridFlow-Dynamics/internal/messaging"
)

// Service handles crew tracking operations.
type Service struct {
	publisher *messaging.Publisher
	consumer  *messaging.Consumer
	crews     map[string]*domain.Crew
	mu        sync.RWMutex
}

// NewService creates a new crew tracking service.
func NewService(publisher *messaging.Publisher, consumer *messaging.Consumer) *Service {
	return &Service{
		publisher: publisher,
		consumer:  consumer,
		crews:     make(map[string]*domain.Crew),
	}
}

// Start initializes and starts the crew service.
func (s *Service) Start() error {
	// Bind to crew location updates
	err := s.consumer.BindQueue(
		messaging.ExchangeCrewEvents,
		messaging.RoutingKeyCrewLocation,
		s.handleLocationUpdate,
	)
	if err != nil {
		return fmt.Errorf("failed to bind location updates: %w", err)
	}

	// Bind to crew status updates
	err = s.consumer.BindQueue(
		messaging.ExchangeCrewEvents,
		messaging.RoutingKeyCrewStatus,
		s.handleStatusUpdate,
	)
	if err != nil {
		return fmt.Errorf("failed to bind status updates: %w", err)
	}

	return s.consumer.Start()
}

func (s *Service) handleLocationUpdate(body []byte) error {
	var update domain.CrewLocationUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("failed to unmarshal location update: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	crew, exists := s.crews[update.CrewID]
	if !exists {
		crew = &domain.Crew{
			ID:     update.CrewID,
			Status: domain.CrewStatusOffline,
		}
		s.crews[update.CrewID] = crew
	}

	crew.Location = update.Location
	crew.LastUpdate = update.Timestamp

	log.Printf("Crew %s location updated: lat=%.6f, lon=%.6f",
		update.CrewID, update.Location.Latitude, update.Location.Longitude)

	return nil
}

func (s *Service) handleStatusUpdate(body []byte) error {
	var update struct {
		CrewID    string            `json:"crew_id"`
		Status    domain.CrewStatus `json:"status"`
		Timestamp time.Time         `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("failed to unmarshal status update: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	crew, exists := s.crews[update.CrewID]
	if !exists {
		crew = &domain.Crew{
			ID: update.CrewID,
		}
		s.crews[update.CrewID] = crew
	}

	crew.Status = update.Status
	crew.LastUpdate = update.Timestamp

	log.Printf("Crew %s status updated: %s", update.CrewID, update.Status)

	return nil
}

// PublishLocationUpdate publishes a crew location update event.
func (s *Service) PublishLocationUpdate(ctx context.Context, update domain.CrewLocationUpdate) error {
	return s.publisher.Publish(ctx, messaging.ExchangeCrewEvents, messaging.RoutingKeyCrewLocation, update)
}

// PublishStatusUpdate publishes a crew status update event.
func (s *Service) PublishStatusUpdate(ctx context.Context, crewID string, status domain.CrewStatus) error {
	update := struct {
		CrewID    string            `json:"crew_id"`
		Status    domain.CrewStatus `json:"status"`
		Timestamp time.Time         `json:"timestamp"`
	}{
		CrewID:    crewID,
		Status:    status,
		Timestamp: time.Now(),
	}
	return s.publisher.Publish(ctx, messaging.ExchangeCrewEvents, messaging.RoutingKeyCrewStatus, update)
}

// GetCrew returns a crew by ID.
func (s *Service) GetCrew(crewID string) (*domain.Crew, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	crew, exists := s.crews[crewID]
	return crew, exists
}

// GetAllCrews returns all tracked crews.
func (s *Service) GetAllCrews() []*domain.Crew {
	s.mu.RLock()
	defer s.mu.RUnlock()

	crews := make([]*domain.Crew, 0, len(s.crews))
	for _, crew := range s.crews {
		crews = append(crews, crew)
	}
	return crews
}

// RegisterCrew registers a new crew in the system.
func (s *Service) RegisterCrew(crew *domain.Crew) {
	s.mu.Lock()
	defer s.mu.Unlock()
	crew.LastUpdate = time.Now()
	s.crews[crew.ID] = crew
}

// Stop stops the crew service.
func (s *Service) Stop() error {
	return s.consumer.Stop()
}
