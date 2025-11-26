// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
	"github.com/120m4n/GridFlow-Dynamics/internal/messaging"
)

// UUIDRegex validates UUID format.
var UUIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// TrackingHandler handles crew tracking requests.
type TrackingHandler struct {
	publisher     *messaging.Publisher
	rateLimiter   *middleware.RateLimiter
	hmacValidator *middleware.HMACValidator
}

// NewTrackingHandler creates a new tracking handler.
func NewTrackingHandler(publisher *messaging.Publisher, rateLimiter *middleware.RateLimiter, hmacValidator *middleware.HMACValidator) *TrackingHandler {
	return &TrackingHandler{
		publisher:     publisher,
		rateLimiter:   rateLimiter,
		hmacValidator: hmacValidator,
	}
}

// TrackingResponse represents the API response.
type TrackingResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ServeHTTP handles POST requests to the tracking endpoint.
func (h *TrackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Validate HMAC signature and read body
	body, valid := h.hmacValidator.ValidateRequest(r)
	if !valid {
		h.sendError(w, http.StatusUnauthorized, "Invalid or missing HMAC signature")
		return
	}

	// Parse the payload
	var payload domain.TrackingPayload
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON payload: %v", err))
		return
	}

	// Validate the payload
	if err := h.validatePayload(&payload); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check rate limit
	if !h.rateLimiter.Allow(payload.CrewID) {
		remaining := h.rateLimiter.Remaining(payload.CrewID)
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		h.sendError(w, http.StatusTooManyRequests, "Rate limit exceeded (100 requests/minute)")
		return
	}

	// Set rate limit header
	remaining := h.rateLimiter.Remaining(payload.CrewID)
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Limit", "100")

	// Convert to tracking event
	event := h.payloadToEvent(&payload)

	// Publish to RabbitMQ (if publisher is available)
	if h.publisher != nil {
		routingKey := fmt.Sprintf("crew.%s.%s", event.Region, payload.CrewID)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.publisher.Publish(ctx, messaging.ExchangeCrewLocations, routingKey, event); err != nil {
			log.Printf("Failed to publish tracking event: %v", err)
			h.sendError(w, http.StatusInternalServerError, "Failed to process tracking update")
			return
		}
	}

	log.Printf("Tracking update received from crew %s: status=%s, progress=%d%%, lat=%.6f, lon=%.6f",
		payload.CrewID, payload.Status, payload.ProgressPercentage,
		payload.GPSCoordinates.Latitude, payload.GPSCoordinates.Longitude)

	// Send success response
	h.sendSuccess(w, "Tracking update processed successfully", payload.CrewID)
}

func (h *TrackingHandler) validatePayload(p *domain.TrackingPayload) error {
	// Validate crewId (UUID format)
	if !UUIDRegex.MatchString(p.CrewID) {
		return fmt.Errorf("crewId must be a valid UUID")
	}

	// Validate timestamp
	if p.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	// Validate GPS coordinates
	if p.GPSCoordinates.Latitude < -90 || p.GPSCoordinates.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if p.GPSCoordinates.Longitude < -180 || p.GPSCoordinates.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	// Validate status
	switch p.Status {
	case domain.TrackingStatusEnRoute, domain.TrackingStatusWorking,
		domain.TrackingStatusPaused, domain.TrackingStatusCompleted:
		// Valid status
	default:
		return fmt.Errorf("status must be one of: en_route, working, paused, completed")
	}

	// Validate progressPercentage
	if p.ProgressPercentage < 0 || p.ProgressPercentage > 100 {
		return fmt.Errorf("progressPercentage must be between 0 and 100")
	}

	// Validate batteryLevel
	if p.BatteryLevel < 0 || p.BatteryLevel > 100 {
		return fmt.Errorf("batteryLevel must be between 0 and 100")
	}

	// Set default region if not provided
	if p.Region == "" {
		p.Region = "default"
	}

	return nil
}

func (h *TrackingHandler) payloadToEvent(p *domain.TrackingPayload) *domain.TrackingEvent {
	return &domain.TrackingEvent{
		CrewID:    p.CrewID,
		Timestamp: p.Timestamp,
		Location: domain.Location{
			Latitude:  p.GPSCoordinates.Latitude,
			Longitude: p.GPSCoordinates.Longitude,
			Accuracy:  0, // Not provided in tracking payload
		},
		TaskID:              p.TaskID,
		Status:              p.Status,
		ProgressPercentage:  p.ProgressPercentage,
		ResourceConsumption: p.ResourceConsumption,
		SafetyAlerts:        p.SafetyAlerts,
		BatteryLevel:        p.BatteryLevel,
		Region:              p.Region,
		ReceivedAt:          time.Now(),
	}
}

func (h *TrackingHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(TrackingResponse{
		Success: false,
		Error:   message,
	})
}

func (h *TrackingHandler) sendSuccess(w http.ResponseWriter, message string, requestID string) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TrackingResponse{
		Success:   true,
		Message:   message,
		RequestID: requestID,
	})
}
