// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"

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

// Handle handles POST requests to the tracking endpoint using Fiber.
func (h *TrackingHandler) Handle(c *fiber.Ctx) error {
	// Validate HMAC signature
	body := c.Body()
	signature := c.Get(middleware.SignatureHeader)
	if !h.hmacValidator.ValidateSignature(body, signature) {
		return h.sendError(c, fiber.StatusUnauthorized, "Invalid or missing HMAC signature")
	}

	// Parse the payload
	var payload domain.TrackingPayload
	if err := c.BodyParser(&payload); err != nil {
		return h.sendError(c, fiber.StatusBadRequest, fmt.Sprintf("Invalid JSON payload: %v", err))
	}

	// Validate the payload
	if err := h.validatePayload(&payload); err != nil {
		return h.sendError(c, fiber.StatusBadRequest, err.Error())
	}

	// Check rate limit
	if !h.rateLimiter.Allow(payload.CrewID) {
		remaining := h.rateLimiter.Remaining(payload.CrewID)
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		return h.sendError(c, fiber.StatusTooManyRequests, "Rate limit exceeded (100 requests/minute)")
	}

	// Set rate limit header
	remaining := h.rateLimiter.Remaining(payload.CrewID)
	c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	c.Set("X-RateLimit-Limit", "100")

	// Convert to tracking event
	event := h.payloadToEvent(&payload)

	// Publish to RabbitMQ (if publisher is available)
	if h.publisher != nil {
		routingKey := fmt.Sprintf("crew.%s.%s", event.Region, payload.CrewID)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.publisher.Publish(ctx, messaging.ExchangeCrewLocations, routingKey, event); err != nil {
			log.Printf("Failed to publish tracking event: %v", err)
			return h.sendError(c, fiber.StatusInternalServerError, "Failed to process tracking update")
		}
	}

	log.Printf("Tracking update received from crew %s: status=%s, progress=%d%%, lat=%.6f, lon=%.6f",
		payload.CrewID, payload.Status, payload.ProgressPercentage,
		payload.GPSCoordinates.Latitude, payload.GPSCoordinates.Longitude)

	// Send success response
	return h.sendSuccess(c, "Tracking update processed successfully", payload.CrewID)
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

func (h *TrackingHandler) sendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(TrackingResponse{
		Success: false,
		Error:   message,
	})
}

func (h *TrackingHandler) sendSuccess(c *fiber.Ctx, message string, requestID string) error {
	return c.Status(fiber.StatusOK).JSON(TrackingResponse{
		Success:   true,
		Message:   message,
		RequestID: requestID,
	})
}
