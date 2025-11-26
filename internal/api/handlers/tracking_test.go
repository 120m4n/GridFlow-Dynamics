package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
)

func TestUUIDRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"invalid-uuid", false},
		{"550e8400e29b41d4a716446655440000", false},
		{"", false},
		{"550e8400-e29b-41d4-a716-44665544000G", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := UUIDRegex.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("UUIDRegex.MatchString(%s) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrackingHandlerMethodNotAllowed(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	// Only POST is allowed
	app.Post("/api/v1/tracking", handler.Handle)

	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/tracking", nil)
			resp, _ := app.Test(req)

			// Fiber returns 405 for methods not allowed when route is not registered
			if resp.StatusCode != fiber.StatusMethodNotAllowed {
				t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusMethodNotAllowed)
			}
		})
	}
}

func TestTrackingHandlerMissingSignature(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	body := `{"crewId":"550e8400-e29b-41d4-a716-446655440000"}`
	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestTrackingHandlerInvalidJSON(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	body := `{invalid json}`
	sig := hmacValidator.ComputeSignature([]byte(body))

	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestTrackingHandlerValidation(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	tests := []struct {
		name        string
		payload     domain.TrackingPayload
		expectError bool
	}{
		{
			name: "invalid UUID",
			payload: domain.TrackingPayload{
				CrewID:             "not-a-uuid",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid latitude",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 100.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid longitude",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -200.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid status",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             "invalid_status",
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid progress",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 150,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid battery level",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       -10,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			sig := hmacValidator.ComputeSignature(body)

			req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(middleware.SignatureHeader, sig)
			resp, _ := app.Test(req)

			if tt.expectError && resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("Status code = %d; want %d for %s", resp.StatusCode, fiber.StatusBadRequest, tt.name)
			}
		})
	}
}

func TestTrackingHandlerRateLimiting(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(2, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	payload := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body, _ := json.Marshal(payload)
	sig := hmacValidator.ComputeSignature(body)

	// First 2 requests should pass validation (may succeed since no publisher)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(middleware.SignatureHeader, sig)
		resp, _ := app.Test(req)
		// Should not be rate limited (will be 200 since no publisher)
		if resp.StatusCode == fiber.StatusTooManyRequests {
			t.Errorf("Request %d should not be rate limited", i+1)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusTooManyRequests)
	}
}

func TestTrackingHandlerSuccess(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	payload := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body, _ := json.Marshal(payload)
	sig := hmacValidator.ComputeSignature(body)

	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusOK)
	}

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)
	var response TrackingResponse
	json.Unmarshal(respBody, &response)

	if !response.Success {
		t.Error("Response should be successful")
	}
	if response.RequestID != payload.CrewID {
		t.Errorf("RequestID = %s; want %s", response.RequestID, payload.CrewID)
	}
}

func TestTrackingResponse(t *testing.T) {
	resp := TrackingResponse{
		Success:   true,
		Message:   "Success",
		RequestID: "test-123",
	}

	if !resp.Success {
		t.Error("Success should be true")
	}

	if resp.Message != "Success" {
		t.Errorf("Message = %s; want Success", resp.Message)
	}

	if resp.RequestID != "test-123" {
		t.Errorf("RequestID = %s; want test-123", resp.RequestID)
	}
}
