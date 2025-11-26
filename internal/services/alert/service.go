// Package alert provides the alert management microservice.
package alert

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

// Service handles alert management operations.
type Service struct {
	publisher *messaging.Publisher
	consumer  *messaging.Consumer
	alerts    map[string]*domain.Alert
	mu        sync.RWMutex
	idCounter int
}

// NewService creates a new alert management service.
func NewService(publisher *messaging.Publisher, consumer *messaging.Consumer) *Service {
	return &Service{
		publisher: publisher,
		consumer:  consumer,
		alerts:    make(map[string]*domain.Alert),
	}
}

// Start initializes and starts the alert service.
func (s *Service) Start() error {
	// Bind to alert created events
	err := s.consumer.BindQueue(
		messaging.ExchangeAlertEvents,
		messaging.RoutingKeyAlertCreated,
		s.handleAlertCreated,
	)
	if err != nil {
		return fmt.Errorf("failed to bind alert created events: %w", err)
	}

	// Bind to alert acknowledged events
	err = s.consumer.BindQueue(
		messaging.ExchangeAlertEvents,
		messaging.RoutingKeyAlertAcked,
		s.handleAlertAcked,
	)
	if err != nil {
		return fmt.Errorf("failed to bind alert acked events: %w", err)
	}

	// Bind to alert resolved events
	err = s.consumer.BindQueue(
		messaging.ExchangeAlertEvents,
		messaging.RoutingKeyAlertResolved,
		s.handleAlertResolved,
	)
	if err != nil {
		return fmt.Errorf("failed to bind alert resolved events: %w", err)
	}

	return s.consumer.Start()
}

func (s *Service) handleAlertCreated(body []byte) error {
	var event domain.AlertEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal alert event: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.idCounter++
	alertID := fmt.Sprintf("ALT-%d", s.idCounter)

	alert := &domain.Alert{
		ID:          alertID,
		CrewID:      event.CrewID,
		TaskID:      event.TaskID,
		Category:    event.Category,
		Severity:    event.Severity,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		CreatedAt:   event.Timestamp,
	}

	s.alerts[alertID] = alert

	log.Printf("Alert created: [%s] %s - %s (Severity: %s, Crew: %s)",
		alertID, alert.Title, alert.Description, alert.Severity, alert.CrewID)

	return nil
}

func (s *Service) handleAlertAcked(body []byte) error {
	var ack struct {
		AlertID   string    `json:"alert_id"`
		Timestamp time.Time `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &ack); err != nil {
		return fmt.Errorf("failed to unmarshal alert ack: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[ack.AlertID]
	if !exists {
		log.Printf("Alert %s not found for acknowledgment", ack.AlertID)
		return nil
	}

	alert.AckedAt = &ack.Timestamp
	log.Printf("Alert %s acknowledged", ack.AlertID)

	return nil
}

func (s *Service) handleAlertResolved(body []byte) error {
	var resolve struct {
		AlertID   string    `json:"alert_id"`
		Timestamp time.Time `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &resolve); err != nil {
		return fmt.Errorf("failed to unmarshal alert resolve: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[resolve.AlertID]
	if !exists {
		log.Printf("Alert %s not found for resolution", resolve.AlertID)
		return nil
	}

	alert.ResolvedAt = &resolve.Timestamp
	log.Printf("Alert %s resolved", resolve.AlertID)

	return nil
}

// PublishAlert publishes an alert event.
func (s *Service) PublishAlert(ctx context.Context, event domain.AlertEvent) error {
	return s.publisher.Publish(ctx, messaging.ExchangeAlertEvents, messaging.RoutingKeyAlertCreated, event)
}

// AcknowledgeAlert publishes an alert acknowledgment event.
func (s *Service) AcknowledgeAlert(ctx context.Context, alertID string) error {
	ack := struct {
		AlertID   string    `json:"alert_id"`
		Timestamp time.Time `json:"timestamp"`
	}{
		AlertID:   alertID,
		Timestamp: time.Now(),
	}
	return s.publisher.Publish(ctx, messaging.ExchangeAlertEvents, messaging.RoutingKeyAlertAcked, ack)
}

// ResolveAlert publishes an alert resolution event.
func (s *Service) ResolveAlert(ctx context.Context, alertID string) error {
	resolve := struct {
		AlertID   string    `json:"alert_id"`
		Timestamp time.Time `json:"timestamp"`
	}{
		AlertID:   alertID,
		Timestamp: time.Now(),
	}
	return s.publisher.Publish(ctx, messaging.ExchangeAlertEvents, messaging.RoutingKeyAlertResolved, resolve)
}

// GetAlert returns an alert by ID.
func (s *Service) GetAlert(alertID string) (*domain.Alert, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	alert, exists := s.alerts[alertID]
	return alert, exists
}

// GetAllAlerts returns all alerts.
func (s *Service) GetAllAlerts() []*domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*domain.Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetActiveAlerts returns all unresolved alerts.
func (s *Service) GetActiveAlerts() []*domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alerts []*domain.Alert
	for _, alert := range s.alerts {
		if alert.ResolvedAt == nil {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetAlertsBySeverity returns alerts filtered by severity.
func (s *Service) GetAlertsBySeverity(severity domain.AlertSeverity) []*domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alerts []*domain.Alert
	for _, alert := range s.alerts {
		if alert.Severity == severity {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetAlertsByCrew returns alerts from a specific crew.
func (s *Service) GetAlertsByCrew(crewID string) []*domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alerts []*domain.Alert
	for _, alert := range s.alerts {
		if alert.CrewID == crewID {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// Stop stops the alert service.
func (s *Service) Stop() error {
	return s.consumer.Stop()
}
